package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
)

// WebhookEvent represents an event to dispatch via webhook.
type WebhookEvent struct {
	Type      string      `json:"type"`
	Timestamp int64       `json:"timestamp"`
	Data      interface{} `json:"data"`
}

var (
	webhookURL string
	whMu       sync.RWMutex
)

// SetWebhookURL sets the global webhook URL.
func SetWebhookURL(url string) {
	whMu.Lock()
	defer whMu.Unlock()
	webhookURL = url
	logger.SysLog(fmt.Sprintf("webhook URL set: %s", url))
}

// GetWebhookURL returns the current webhook URL.
func GetWebhookURL() string {
	whMu.RLock()
	defer whMu.RUnlock()
	return webhookURL
}

// DispatchWebhook sends a webhook event to the configured URL asynchronously.
func DispatchWebhook(eventType string, data interface{}) {
	url := GetWebhookURL()
	if url == "" {
		return
	}

	event := WebhookEvent{
		Type:      eventType,
		Timestamp: time.Now().Unix(),
		Data:      data,
	}

	go func() {
		body, err := json.Marshal(event)
		if err != nil {
			logger.SysError(fmt.Sprintf("failed to marshal webhook event: %s", err.Error()))
			return
		}

		req, err := http.NewRequest("POST", url, bytes.NewReader(body))
		if err != nil {
			logger.SysError(fmt.Sprintf("failed to create webhook request: %s", err.Error()))
			return
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			logger.SysError(fmt.Sprintf("webhook dispatch failed: %s", err.Error()))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 300 {
			logger.SysError(fmt.Sprintf("webhook returned status %d", resp.StatusCode))
		}
	}()
}
