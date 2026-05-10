package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/relay/model"
)

// ==================== Files API 中转 ====================


func RelayFilesHelper(c *gin.Context) *model.ErrorWithStatusCode {
	return relayGenericProxy(c)
}

// ==================== Assistants API 中转 ====================

func RelayAssistantsHelper(c *gin.Context) *model.ErrorWithStatusCode {
	return relayGenericProxy(c)
}

func RelayThreadsHelper(c *gin.Context) *model.ErrorWithStatusCode {
	return relayGenericProxy(c)
}

// ==================== Fine-tuning API 中转 ====================

func RelayFineTuningHelper(c *gin.Context) *model.ErrorWithStatusCode {
	return relayGenericProxy(c)
}

// ==================== 通用透传中转 ====================
// 适用于 Assistants / Files / Fine-tuning / Threads 等 API

func relayGenericProxy(c *gin.Context) *model.ErrorWithStatusCode {
	method := c.Request.Method
	path := c.Request.URL.Path

	adaptor := getRelayAdaptorFromContext(c)
	if adaptor == nil {
		return wrapError(fmt.Errorf("no adaptor configured"), "invalid_request", http.StatusBadRequest)
	}

	body, _ := common.GetRequestBody(c)
	upstreamBase := getUpstreamBase(c)
	if upstreamBase == "" {
		return wrapError(fmt.Errorf("no upstream base URL"), "configuration_error", http.StatusInternalServerError)
	}
	upstreamURL := upstreamBase + path

	proxyReq, err := http.NewRequestWithContext(c.Request.Context(), method, upstreamURL, bytes.NewReader(body))
	if err != nil {
		return wrapError(err, "invalid_request", http.StatusBadRequest)
	}

	// 复制关键 headers
	for _, h := range []string{"Content-Type", "Authorization", "OpenAI-Beta", "OpenAI-Organization"} {
		if v := c.GetHeader(h); v != "" {
			proxyReq.Header.Set(h, v)
		}
	}

	resp, err := http.DefaultClient.Do(proxyReq)
	if err != nil {
		logger.Errorf(c.Request.Context(), "relay proxy failed for %s %s: %v", method, path, err)
		return wrapError(err, "upstream_error", http.StatusBadGateway)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	for k, v := range resp.Header {
		if len(v) > 0 {
			c.Header(k, v[0])
		}
	}
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
	return nil
}

func wrapError(err error, code string, status int) *model.ErrorWithStatusCode {
	return &model.ErrorWithStatusCode{
		Error: model.Error{
			Message: err.Error(),
			Type:    "relay_error",
			Code:    code,
		},
		StatusCode: status,
	}
}

// getUpstreamBase 从 channel context 获取上游 base URL
func getUpstreamBase(c *gin.Context) string {
	// 从 middleware 设置的 context 中获取 channel base URL
	// 这里用 gin 的 GetString 尝试获取
	if baseURL, exists := c.Get("UpstreamBaseURL"); exists {
		if s, ok := baseURL.(string); ok {
			return s
		}
	}
	return ""
}

// getRelayAdaptorFromContext 获取 relay adaptor（来自 middleware）
func getRelayAdaptorFromContext(c *gin.Context) RelayAdaptorLike {
	if adaptor, exists := c.Get("RelayAdaptor"); exists {
		if a, ok := adaptor.(RelayAdaptorLike); ok {
			return a
		}
	}
	return nil
}

// RelayAdaptorLike relay adaptor 简化接口
type RelayAdaptorLike interface {
	GetAPIType() int
}

// json import reference (avoids unused import error)
var _ = json.Unmarshal
