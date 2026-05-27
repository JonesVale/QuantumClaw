// Package quantumclaw provides a Go client for QuantumClaw AI API Gateway.
package quantumclaw

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is the QuantumClaw API client.
type Client struct {
	apiKey    string
	baseURL   string
	http      *http.Client
	userAgent string
}

// NewClient creates a new QuantumClaw client.
func NewClient(apiKey, baseURL string) *Client {
	if baseURL == "" {
		baseURL = "http://localhost:3666"
	}
	return &Client{
		apiKey:    apiKey,
		baseURL:   strings.TrimRight(baseURL, "/"),
		http:      &http.Client{Timeout: 120 * time.Second},
		userAgent: "quantumclaw-go-sdk/0.1.0",
	}
}

// ChatCompletionRequest represents a chat completion request.
type ChatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

// ChatMessage represents a single message in a conversation.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionResponse represents a non-streaming response.
type ChatCompletionResponse struct {
	ID      string              `json:"id"`
	Object  string              `json:"object"`
	Created int64               `json:"created"`
	Model   string              `json:"model"`
	Choices []ChatCompletionChoice `json:"choices"`
	Usage   *UsageInfo          `json:"usage,omitempty"`
}

// ChatCompletionChoice represents a single choice.
type ChatCompletionChoice struct {
	Index        int          `json:"index"`
	Message      ChatMessage  `json:"message"`
	FinishReason string       `json:"finish_reason"`
}

// UsageInfo contains token usage information.
type UsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamEvent represents a single SSE event in streaming mode.
type StreamEvent struct {
	ID                string `json:"id"`
	Object            string `json:"object"`
	Created           int64  `json:"created"`
	Model             string `json:"model"`
	Choices           []StreamChoice `json:"choices"`
	Usage             *UsageInfo `json:"usage,omitempty"`
	Error             *StreamError `json:"error,omitempty"`
}

// StreamChoice represents a streaming choice delta.
type StreamChoice struct {
	Index        int            `json:"index"`
	Delta        StreamDelta    `json:"delta"`
	FinishReason string         `json:"finish_reason"`
}

// StreamDelta contains the incremental content.
type StreamDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// StreamError represents an error in the stream.
type StreamError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// Balance represents the user's balance info.
type Balance struct {
	Quota      float64 `json:"quota"`
	UsedQuota  float64 `json:"used_quota"`
}

// doRequest performs an HTTP request and unmarshals JSON.
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return resp, nil
}

// ChatCompletion sends a non-streaming chat completion request.
func (c *Client) ChatCompletion(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	req.Stream = false
	resp, err := c.doRequest(ctx, "POST", "/v1/chat/completions", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &result, nil
}

// StreamChatCompletion sends a streaming chat completion request.
func (c *Client) StreamChatCompletion(ctx context.Context, req *ChatCompletionRequest) (<-chan *StreamEvent, error) {
	req.Stream = true
	resp, err := c.doRequest(ctx, "POST", "/v1/chat/completions", req)
	if err != nil {
		return nil, err
	}

	ch := make(chan *StreamEvent)
	go func() {
		defer resp.Body.Close()
		defer close(ch)

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				return
			}
			var event StreamEvent
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				continue
			}
			select {
			case ch <- &event:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}

// ListModels returns the available models.
func (c *Client) ListModels(ctx context.Context) ([]map[string]interface{}, error) {
	resp, err := c.doRequest(ctx, "GET", "/api/models", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Success bool                   `json:"success"`
		Data    []map[string]interface{} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode models: %w", err)
	}
	return result.Data, nil
}

// GetBalance returns the current user's balance information.
func (c *Client) GetBalance(ctx context.Context) (*Balance, error) {
	resp, err := c.doRequest(ctx, "GET", "/api/user/balance", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Success bool     `json:"success"`
		Data    *Balance `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode balance: %w", err)
	}
	return result.Data, nil
}
