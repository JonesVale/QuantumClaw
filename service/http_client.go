package service

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
)

var HttpClient *http.Client

func InitHttpClient() {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: false},
		DisableCompression:  false,
	}

	if proxyURL := os.Getenv("RELAY_PROXY"); proxyURL != "" {
		if parsed, err := url.Parse(proxyURL); err == nil {
			transport.Proxy = http.ProxyURL(parsed)
			logger.SysLog("relay proxy configured: " + proxyURL)
		}
	}

	HttpClient = &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}
}
