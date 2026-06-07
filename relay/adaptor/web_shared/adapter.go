// Package web_shared provides the generic schema-driven web model adapter engine.
package web_shared

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/client"
	"github.com/quantumclaw/quantumclaw/common/ctxkey"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/relay/meta"
	relaymodel "github.com/quantumclaw/quantumclaw/relay/model"
)

var (
	ErrNoSchema      = fmt.Errorf("no active sub2api schema found for provider")
	ErrNoCredential  = fmt.Errorf("no valid sub2api credential found")
	ErrCircuitOpen   = fmt.Errorf("circuit breaker is open for this schema")
	ErrRequestFailed = fmt.Errorf("sub2api request failed")
)

type Adapter struct{}

var GlobalAdapter = &Adapter{}

// MatchRequest checks if model can be served by Sub2API and user has credentials.
func (a *Adapter) MatchRequest(userId int, apiModel string) (string, bool, error) {
	schemas, err := model.ListSub2APISchemas()
	if err != nil {
		return "", false, err
	}
	for _, s := range schemas {
		if s.Status != 1 || DefaultCircuitBreaker.IsCircuitOpen(s.Id) {
			continue
		}
		if !schemaMatchesModel(&s, apiModel) {
			continue
		}
		_, found, err := ResolveCredential(userId, s.Provider)
		if err != nil || !found {
			continue
		}
		return s.Provider, true, nil
	}
	return "", false, nil
}

func schemaMatchesModel(s *model.Sub2APISchema, apiModel string) bool {
	mapping, err := s.ParseModelMapping()
	if err != nil || len(mapping) == 0 {
		return false
	}
	if _, ok := mapping[apiModel]; ok {
		return true
	}
	if _, ok := mapping["*"]; ok {
		return true
	}
	return false
}

// ExecuteRequest performs the HTTP request against the provider backend.
// Returns the raw *http.Response from the provider.
func (a *Adapter) ExecuteRequest(c *gin.Context, meta *meta.Meta, provider string,
	textRequest *relaymodel.GeneralOpenAIRequest) (*http.Response, *relaymodel.ErrorWithStatusCode) {

	ctx := c.Request.Context()
	userId := c.GetInt("id")

	schema, err := model.GetActiveSchema(provider)
	if err != nil || schema == nil {
		return nil, errorWithCode(fmt.Errorf("no active schema for %s: %w", provider, err), 502)
	}
	if DefaultCircuitBreaker.IsCircuitOpen(schema.Id) {
		if fb, fbErr := FindFallbackSchema(provider, schema.Version); fbErr == nil && fb != nil {
			schema = fb
		} else {
			return nil, errorWithCode(ErrCircuitOpen, 502)
		}
	}

	token, found, err := ResolveCredential(userId, provider)
	if err != nil || !found {
		return nil, errorWithCode(ErrNoCredential, 401)
	}

	input := &RenderInput{
		Schema:      schema,
		APIModel:    textRequest.Model,
		Messages:    textRequest.Messages,
		Stream:      meta.IsStream,
		MaxTokens:   textRequest.MaxTokens,
		Temperature: getTemperatureFn(textRequest),
		CookieToken: token,
		OrgID:       "",
	}
	output, renderErr := RenderRequest(input)
	if renderErr != nil {
		return nil, errorWithCode(fmt.Errorf("template rendering failed: %w", renderErr), 500)
	}

	req, reqErr := http.NewRequestWithContext(ctx, output.Method, output.EndpointURL,
		bytes.NewBufferString(output.Body))
	if reqErr != nil {
		return nil, errorWithCode(fmt.Errorf("http request create failed: %w", reqErr), 500)
	}
	for k, v := range output.Headers {
		req.Header.Set(k, v)
	}

	httpResp, httpErr := client.HTTPClient.Do(req)
	if httpErr != nil {
		DefaultCircuitBreaker.RecordFailure(schema.Id)
		return nil, errorWithCode(fmt.Errorf("http call failed: %w", httpErr), 502)
	}

	if httpResp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		httpResp.Body.Close()
		bodyStr := string(bodyBytes)
		if len(bodyStr) > 500 {
			bodyStr = bodyStr[:500]
		}
		DefaultCircuitBreaker.RecordFailure(schema.Id)
		return a.tryFallbackRequest(c, meta, provider, schema, textRequest, httpResp.StatusCode, bodyStr)
	}

	DefaultCircuitBreaker.RecordSuccess(schema.Id)
	return httpResp, nil
}

// ==================== 回退链追踪 ====================
// 核心修复：当 relay 层发生回退时，必须更新 Gin Context，
// 否则下游 PostConsumeDeduct 使用的是过期的 ChannelId，
// 导致计费对象错误（渠道商 A 被扣款但实际服务由渠道商 B 提供）。
//
// 追踪字段：
//   - original_channel_id : 分发器最初选择的渠道（始终不变）
//   - fallback_chain       : JSON 数组记录每次回退的 [from_schema, to_schema]
//   - is_fallback          : 标记是否发生过回退
//   - actual_schema_id     : 最终实际使用的 schema ID

func (a *Adapter) tryFallbackRequest(c *gin.Context, meta *meta.Meta, provider string,
	failedSchema *model.Sub2APISchema, textRequest *relaymodel.GeneralOpenAIRequest,
	statusCode int, body string) (*http.Response, *relaymodel.ErrorWithStatusCode) {

	fb, err := FindFallbackSchema(provider, failedSchema.Version)
	if err != nil || fb == nil {
		return nil, errorWithCode(
			fmt.Errorf("all schemas failed for %s: status=%d body=%s", provider, statusCode, body), 502)
	}

	// ── 核心：回退成功时更新 Gin Context ──
	// 记录原始渠道 ID（仅首次回退时记录）
	if _, exists := c.Get("original_channel_id"); !exists {
		c.Set("original_channel_id", c.GetInt(ctxkey.ChannelId))
	}
	// 标记回退状态
	c.Set(ctxkey.IsFallback, true)

	// 记录回退链（用于审计）
	var fallbackChain []map[string]int
	if existing, exists := c.Get("fallback_chain"); exists {
		fallbackChain = existing.([]map[string]int)
	} else {
		fallbackChain = make([]map[string]int, 0)
	}
	fallbackChain = append(fallbackChain, map[string]int{
		"from_schema": failedSchema.Id,
		"to_schema":   fb.Id,
	})
	c.Set("fallback_chain", fallbackChain)

	// 记录最终实际使用的 schema
	c.Set("actual_schema_id", fb.Id)

	logger.Warnf(c.Request.Context(),
		"[FALLBACK] provider=%s schema:%d→%d channel_id=%d status=%d",
		provider, failedSchema.Id, fb.Id, c.GetInt(ctxkey.ChannelId), statusCode)

	return a.ExecuteRequest(c, meta, provider, textRequest)
}

// ParseStreamResponse parses a streaming (SSE) response and writes SSE chunks to the client.
func (a *Adapter) ParseStreamResponse(c *gin.Context, httpResp *http.Response,
	responsePath string, meta *meta.Meta) *relaymodel.Usage {

	ctx := c.Request.Context()
	eventCh := StreamSSE(ctx, httpResp, responsePath)
	var fullContent strings.Builder

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.WriteHeader(http.StatusOK)

	for event := range eventCh {
		if event.Error != nil {
			break
		}
		if event.Done {
			finishChunk := `{"choices":[{"delta":{"content":""},"finish_reason":"stop","index":0}]}`
			c.Writer.Write([]byte(FormatSSE(finishChunk)))
			c.Writer.Write([]byte(FormatSSEDone()))
			c.Writer.Flush()
			break
		}
		fullContent.WriteString(event.Content)
		escaped, _ := json.Marshal(event.Content)
		chunk := fmt.Sprintf(`{"choices":[{"delta":{"content":%s},"index":0,"finish_reason":null}]}`, string(escaped))
		c.Writer.Write([]byte(FormatSSE(chunk)))
		c.Writer.Flush()
	}

	return &relaymodel.Usage{
		PromptTokens:     meta.PromptTokens,
		CompletionTokens: estimateTokens(fullContent.String()),
		TotalTokens:      meta.PromptTokens + estimateTokens(fullContent.String()),
	}
}

// ParseNonStreamResponse parses a non-streaming response and writes OpenAI-compatible JSON.
func (a *Adapter) ParseNonStreamResponse(c *gin.Context, httpResp *http.Response,
	responsePath string, originModel string, meta *meta.Meta) *relaymodel.Usage {

	result, err := StreamPoll(context.Background(), httpResp, responsePath)
	if err != nil {
		return &relaymodel.Usage{
			PromptTokens:     meta.PromptTokens,
			CompletionTokens: 0,
			TotalTokens:      meta.PromptTokens,
		}
	}

	usage := &relaymodel.Usage{
		PromptTokens:     meta.PromptTokens,
		CompletionTokens: estimateTokens(result.Content),
		TotalTokens:      meta.PromptTokens + estimateTokens(result.Content),
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"choices": []map[string]interface{}{
			{
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": result.Content,
				},
				"finish_reason": "stop",
				"index": 0,
			},
		},
		"usage": map[string]interface{}{
			"prompt_tokens":     usage.PromptTokens,
			"completion_tokens": usage.CompletionTokens,
			"total_tokens":      usage.TotalTokens,
		},
		"model": originModel,
	})
	return usage
}

func getTemperatureFn(req *relaymodel.GeneralOpenAIRequest) float64 {
	if req == nil || req.Temperature == nil {
		return 1.0
	}
	return *req.Temperature
}

func errorWithCode(err error, code int) *relaymodel.ErrorWithStatusCode {
	return &relaymodel.ErrorWithStatusCode{
		StatusCode: code,
		Error: relaymodel.Error{
			Message: err.Error(),
			Type:    "sub2api_error",
		},
	}
}

func estimateTokens(text string) int {
	if len(text) == 0 {
		return 0
	}
	n := len(text) / 3
	if n < 1 {
		return 1
	}
	return n
}
