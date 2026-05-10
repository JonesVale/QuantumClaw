package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/setting/system_setting"
)

// SSRFProtection 创建 SSRF 防护中间件
// 对所有外发请求进行 URL 校验（用于 relay/渠道调用）
func SSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只在 EnableSSRFProtection 为 true 时生效
		setting := system_setting.GetFetchSetting()
		if !setting.EnableSSRFProtection {
			c.Next()
			return
		}

		// 对请求体中的 URL 字段进行 SSRF 检查
		// 先放行，等 body 读完再检查
		c.Next()
	}
}

// SSRFProtectionRequest 对单个外发请求进行 SSRF 检查
// 在发起 relay 请求前调用，返回 (ok bool, reason string)
func SSRFProtectionRequest(req *http.Request) (bool, string) {
	if req == nil {
		return true, ""
	}

	// 检查 Host header
	if ok, reason := system_setting.SafeHostHeader(req.Host); !ok {
		return false, reason
	}

	// 检查请求 URL
	return system_setting.ValidateURL(req.URL.String())
}

// SSRFProtectionBody 检查请求体中的 URL
// 适用于 POST/PUT 请求体中包含 BaseURL 等字段的场景
func SSRFProtectionBody(body []byte) (bool, string, []byte) {
	if len(body) == 0 {
		return true, "", body
	}

	// 简单字符串检测：查找可能的 URL 模式
	text := string(body)
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 跳过注释和短行
		if strings.HasPrefix(line, "#") || len(line) < 10 {
			continue
		}
		// 检测 URL 模式 (http:// https://)
		for _, prefix := range []string{"http://", "https://", "base_url", "BaseURL", "baseUrl"} {
			idx := strings.Index(line, prefix)
			if idx >= 0 {
				start := idx
				// 找到 URL 结尾（通常是空格、引号、逗号）
				end := start + len(prefix)
				for end < len(line) && !strings.Contains(" \t\n\r,\"']", string(line[end])) {
					end++
				}
				url := line[start:end]
				if ok, reason := system_setting.ValidateURL(url); !ok {
					return false, reason, body
				}
			}
		}
	}

	return true, "", body
}

// ProxyTransport 带 SSRF 保护的 HTTP Transport
// 用于 relay 代理外发请求
type SSRFProxyTransport struct {
	Transport http.RoundTripper
}

func (t *SSRFProxyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if ok, reason := SSRFProtectionRequest(req); !ok {
		return &http.Response{
			StatusCode: http.StatusForbidden,
			Status:     "SSRF Blocked: " + reason,
			Body:       io.NopCloser(bytes.NewReader([]byte("SSRF blocked: " + reason))),
			Header:     http.Header{},
		}, nil
	}

	transport := t.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	return transport.RoundTrip(req)
}

// NewSSRFProxyTransport 创建带 SSRF 保护的 RoundTripper
func NewSSRFProxyTransport() http.RoundTripper {
	return &SSRFProxyTransport{}
}
