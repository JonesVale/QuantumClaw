package client

import (
	"bytes"
	"io"
	"net/http"

	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/setting/system_setting"
)

// ssrfProxyTransport 带 SSRF 防护的 HTTP Transport
// 从 middleware/ssrf_protection.go 复制到此，打破 import cycle：
// common/client → middleware → relay/adaptor → common/client
type ssrfProxyTransport struct {
	Transport http.RoundTripper
}

func (t *ssrfProxyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if ok, reason := ssrfCheckRequest(req); !ok {
		logger.SysWarn("SSRF blocked: " + reason)
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

// ssrfCheckRequest 对单个外发请求进行 SSRF 检查
func ssrfCheckRequest(req *http.Request) (bool, string) {
	if req == nil {
		return true, ""
	}
	return system_setting.ValidateURL(req.URL.String())
}

// wrapWithSSRFProtection 如果需要 SSRF 防护，包装现有 Transport
func wrapWithSSRFProtection(base http.RoundTripper) http.RoundTripper {
	setting := system_setting.GetFetchSetting()
	if !setting.EnableSSRFProtection {
		logger.SysWarn("⚠️ SSRF protection is DISABLED. Enable it via system settings for production security.")
		return base
	}
	return &ssrfProxyTransport{Transport: base}
}
