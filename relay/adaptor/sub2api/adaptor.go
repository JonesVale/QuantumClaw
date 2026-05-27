// Package sub2api implements the relay.Adaptor interface for schema-driven web model providers.
package sub2api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/ctxkey"
	dbmodel "github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/relay/adaptor"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/web_shared"
	"github.com/quantumclaw/quantumclaw/relay/meta"
	"github.com/quantumclaw/quantumclaw/relay/model"
)

type Adaptor struct {
	Meta *meta.Meta
}

func (a *Adaptor) Init(meta *meta.Meta) {
	a.Meta = meta
}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	// URL is resolved dynamically in DoRequest
	return "http://sub2api.local/_internal", nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	adaptor.SetupCommonRequestHeader(c, req, meta)
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, fmt.Errorf("request is nil")
	}
	if request.Stream {
		if request.StreamOptions == nil {
			request.StreamOptions = &model.StreamOptions{}
		}
		request.StreamOptions.IncludeUsage = true
	}
	return request, nil
}

func (a *Adaptor) ConvertImageRequest(request *model.ImageRequest) (any, error) {
	return request, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	// Read the already-marshaled request body
	bodyBytes, err := io.ReadAll(requestBody)
	if err != nil {
		return nil, fmt.Errorf("read request body: %w", err)
	}
	var textRequest model.GeneralOpenAIRequest
	if err := json.Unmarshal(bodyBytes, &textRequest); err != nil {
		return nil, fmt.Errorf("parse request body: %w", err)
	}

	// Restore body for potential retries
	c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	provider := c.GetString(ProviderKey)
	if provider == "" {
		return nil, fmt.Errorf("no sub2api provider set in context")
	}

	// Execute the request against the provider backend
	httpResp, errWithCode := web_shared.GlobalAdapter.ExecuteRequest(c, meta, provider, &textRequest)
	if errWithCode != nil {
		// Return an error response with status code
		jsonBody, _ := json.Marshal(map[string]interface{}{
			"error": map[string]interface{}{
				"message": errWithCode.Error.Message,
				"type":    errWithCode.Error.Type,
			},
		})
		dummyResp := &http.Response{
			StatusCode: errWithCode.StatusCode,
			Body:       io.NopCloser(bytes.NewReader(jsonBody)),
			Header:     make(http.Header),
		}
		dummyResp.Header.Set("Content-Type", "application/json")
		return dummyResp, nil
	}

	return httpResp, nil
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	provider := c.GetString(ProviderKey)
	if provider == "" {
		return nil, &model.ErrorWithStatusCode{
			StatusCode: 500,
			Error: model.Error{
				Message: "no sub2api provider in context",
				Type:    "sub2api_error",
			},
		}
	}

	// Get schema's response path for content extraction
	schema, dbErr := dbmodel.GetActiveSchema(provider)
	if dbErr != nil || schema == nil {
		// Fallback: try to find schema from any source
		schemas, listErr := dbmodel.ListSub2APISchemas()
		if listErr != nil || len(schemas) == 0 {
			return nil, &model.ErrorWithStatusCode{
				StatusCode: 502,
				Error: model.Error{
					Message: "no schema found",
					Type:    "sub2api_error",
				},
			}
		}
		schema = &schemas[0]
	}

	if meta.IsStream {
		u := web_shared.GlobalAdapter.ParseStreamResponse(c, resp, schema.ResponsePath, meta)
		return u, nil
	}

	u := web_shared.GlobalAdapter.ParseNonStreamResponse(c, resp, schema.ResponsePath, meta.OriginModelName, meta)
	return u, nil
}

func (a *Adaptor) GetModelList() []string {
	return []string{
		"gpt-4o", "gpt-4", "gpt-4-turbo", "gpt-3.5-turbo",
		"claude-3-opus", "claude-3-sonnet", "claude-3-haiku",
		"gemini-2.0-flash", "gemini-2.0-pro",
		"deepseek-chat", "deepseek-reasoner",
		"grok-2", "grok-3",
	}
}

func (a *Adaptor) GetChannelName() string {
	return "sub2api"
}

// ProviderKey is the context key used by Sub2API middleware.
const ProviderKey = "sub2api_provider"

// MatchAndSetProvider checks if the request can be served by Sub2API and sets context.
func MatchAndSetProvider(c *gin.Context) bool {
	userId := c.GetInt(ctxkey.Id)
	requestModel := c.GetString(ctxkey.RequestModel)
	if userId == 0 || requestModel == "" {
		return false
	}
	provider, matched, err := web_shared.GlobalAdapter.MatchRequest(userId, requestModel)
	if err != nil || !matched {
		return false
	}
	c.Set(ProviderKey, provider)
	return true
}
