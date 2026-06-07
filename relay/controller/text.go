package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/relay"
	"github.com/quantumclaw/quantumclaw/relay/adaptor"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/openai"
	"github.com/quantumclaw/quantumclaw/relay/apitype"
	billingratio "github.com/quantumclaw/quantumclaw/relay/billing/ratio"
	"github.com/quantumclaw/quantumclaw/relay/channeltype"
	common_handler "github.com/quantumclaw/quantumclaw/relay/common_handler"
	relaycommon "github.com/quantumclaw/quantumclaw/relay/common"
	metapkg "github.com/quantumclaw/quantumclaw/relay/meta"
	"github.com/quantumclaw/quantumclaw/relay/model"
)

func RelayTextHelper(c *gin.Context) *model.ErrorWithStatusCode {
	ctx := c.Request.Context()
	meta := metapkg.GetByContext(c)
	// get & validate textRequest
	textRequest, err := getAndValidateTextRequest(c, meta.Mode)
	if err != nil {
		logger.Errorf(ctx, "getAndValidateTextRequest failed: %s", err.Error())
		return openai.ErrorWrapper(err, "invalid_text_request", http.StatusBadRequest)
	}
	meta.IsStream = textRequest.Stream

	// map model name
	// Parse reasoning/thinking suffixes from the model name and set defaults.
	// e.g. "o3-mini-high" → base="o3-mini", ReasoningEffort="high"
	//      "claude-3-7-sonnet-20250219-thinking" → Thinking=true
	applyReasoningSuffix(textRequest)

	meta.OriginModelName = textRequest.Model
	textRequest.Model, _ = getMappedModelName(textRequest.Model, meta.ModelMapping)
	meta.ActualModelName = textRequest.Model
	// set system prompt if not empty
	systemPromptReset := setSystemPrompt(ctx, textRequest, meta.ForcedSystemPrompt)
	// get model ratio & group ratio
	modelRatio := billingratio.GetModelRatio(textRequest.Model, meta.ChannelType)
	groupRatio := billingratio.GetGroupRatio(meta.Group)
	ratio := modelRatio * groupRatio
	// pre-consume — 使用现金计费
	promptTokens := getPromptTokens(textRequest, meta.Mode)
	meta.PromptTokens = promptTokens
	preConsumedQuota, billingSource, bizErr := preConsumeBalance(ctx, textRequest, promptTokens, ratio, meta)
	if bizErr != nil {
		logger.Warnf(ctx, "preConsumeBalance failed: %+v", *bizErr)
		return bizErr
	}

	// defer 保护：防止 panic 導致額度無法退還
	var quotaRefunded bool
	defer func() {
		if r := recover(); r != nil {
			if !quotaRefunded {
				logger.Warnf(ctx, "panic after pre-consume, refunding %d quota for user=%d token=%d",
					preConsumedQuota, meta.UserId, meta.TokenId)
				// 调用 PostConsumeBilling 退還全部預扣額度
				relayInfo := &relaycommon.RelayInfo{
					UserID:  meta.UserId,
					TokenID: meta.TokenId,
				}
				common_handler.PostConsumeBilling(ctx, billingSource, relayInfo, meta.TokenId, 0, preConsumedQuota)
			}
			panic(r) // 重新 panic，让 RelayPanicRecover 记录堆栈
		}
	}()

	adaptor := relay.GetAdaptor(meta.APIType)
	if adaptor == nil {
		return openai.ErrorWrapper(fmt.Errorf("invalid api type: %d", meta.APIType), "invalid_api_type", http.StatusBadRequest)
	}
	adaptor.Init(meta)

	// get request body
	requestBody, err := getRequestBody(c, meta, textRequest, adaptor)
	if err != nil {
		return openai.ErrorWrapper(err, "convert_request_failed", http.StatusInternalServerError)
	}

	// do request
	resp, err := adaptor.DoRequest(c, meta, requestBody)
	if err != nil {
		logger.Errorf(ctx, "DoRequest failed: %s", err.Error())
		return openai.ErrorWrapper(err, "do_request_failed", http.StatusInternalServerError)
	}
	if isErrorHappened(meta, resp) {
		return RelayErrorHandler(resp)
	}

	// do response
	usage, respErr := adaptor.DoResponse(c, resp, meta)

	// ── 关键修复：即使 DoResponse 返回 error，只要 usage 不为 nil 就要扣款 ──
	// 原因：上游可能已经消耗了 token（比如部分成功、超时但已计费等情况），
	// 如果不扣款，平台会亏损。
	// 扣款前刷新 meta（可能发生了回退，ChannelId 已变化）
	meta = metapkg.GetByContext(c)

	if respErr != nil {
		logger.Errorf(ctx, "respErr is not nil: %+v, usage=%+v", respErr, usage)
		// 即使有 error，只要 usage 不为 nil 且消耗了 token，就执行扣款
		if usage != nil && (usage.PromptTokens > 0 || usage.CompletionTokens > 0) {
			logger.Warnf(ctx, "[BILLING_AUDIT] DoResponse error but usage present, forcing deduct: user=%d usage=%+v", meta.UserId, usage)
			quotaRefunded = true
			postConsumeDeduct(ctx, usage, meta, textRequest, ratio, preConsumedQuota, modelRatio, groupRatio, systemPromptReset, billingSource)
		}
		return respErr
	}

	// ── 关键：刷新 meta 以获取回退后的最新上下文 ──
	// adaptor.DoRequest/DoResponse 过程中可能触发了 schema 回退（tryFallbackRequest），
	// 回退代码会更新 Gin Context（is_fallback、fallback_chain 等）。
	// 此处必须重新从 context 读取，否则计费使用的是过期的 ChannelId。
	meta = metapkg.GetByContext(c)

	// post-consume — 现金扣款 + 分账（同步执行，失败返回错误）
	quotaRefunded = true
	postConsumeDeduct(ctx, usage, meta, textRequest, ratio, preConsumedQuota, modelRatio, groupRatio, systemPromptReset, billingSource)
	return nil
}

func getRequestBody(c *gin.Context, meta *metapkg.Meta, textRequest *model.GeneralOpenAIRequest, adaptor adaptor.Adaptor) (io.Reader, error) {
	if !config.EnforceIncludeUsage &&
		meta.APIType == apitype.OpenAI &&
		meta.OriginModelName == meta.ActualModelName &&
		meta.ChannelType != channeltype.Baichuan &&
		meta.ForcedSystemPrompt == "" {
		// no need to convert request for openai
		return c.Request.Body, nil
	}

	// get request body
	var requestBody io.Reader
	convertedRequest, err := adaptor.ConvertRequest(c, meta.Mode, textRequest)
	if err != nil {
		logger.Debugf(c.Request.Context(), "converted request failed: %s\n", err.Error())
		return nil, err
	}
	jsonData, err := json.Marshal(convertedRequest)
	if err != nil {
		logger.Debugf(c.Request.Context(), "converted request json_marshal_failed: %s\n", err.Error())
		return nil, err
	}
	logger.Debugf(c.Request.Context(), "converted request: \n%s", string(jsonData))
	requestBody = bytes.NewBuffer(jsonData)
	return requestBody, nil
}

// applyReasoningSuffix parses reasoning/thinking suffixes from the model name
// and sets the corresponding fields (ReasoningEffort, Thinking) on the request.
// It also strips the suffix from the model name so downstream code sees the base model.
//
// Supported patterns:
//   - o3-mini-high / o3-mini-medium / o3-mini-low
//   - claude-*-thinking
//   - gemini-*-thinking / gemini-*-nothinking / gemini-*-thinking-<budget>
func applyReasoningSuffix(textRequest *model.GeneralOpenAIRequest) {
	if textRequest == nil || textRequest.Model == "" {
		return
	}
	base, reasoningEffort, thinking, _ := relaycommon.ParseModelSuffix(textRequest.Model)
	if base == textRequest.Model {
		// No suffix was recognized; no changes needed
		return
	}
	// Update model to base name (suffix stripped)
	textRequest.Model = base

	// Set ReasoningEffort only if not already explicitly provided in the request body
	if reasoningEffort != "" && textRequest.ReasoningEffort == nil {
		textRequest.ReasoningEffort = &reasoningEffort
	}

	// Set Thinking flag
	if textRequest.Thinking == nil {
		boolVal := thinking
		textRequest.Thinking = &boolVal
	}
}
