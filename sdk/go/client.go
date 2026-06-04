// Package quantumclaw provides a Go client for QuantumClaw AI API Gateway.
//
// Key format: sk-xxxxxxxx... (51 chars, OpenAI-compatible)
// Pass the full key including "sk-" prefix.
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
// apiKey should include the "sk-" prefix.
// baseURL defaults to "http://localhost:3666".
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

// ==================== DTO Types ====================

// ChatCompletionRequest represents a chat completion request.
type ChatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
	TopP        float64       `json:"top_p,omitempty"`
}

// ChatMessage represents a single message in a conversation.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionResponse represents a non-streaming response.
type ChatCompletionResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChatCompletionChoice `json:"choices"`
	Usage   *UsageInfo             `json:"usage,omitempty"`
}

// ChatCompletionChoice represents a single choice.
type ChatCompletionChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// UsageInfo contains token usage information.
type UsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamEvent represents a single SSE event in streaming mode.
type StreamEvent struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []StreamChoice `json:"choices"`
	Usage   *UsageInfo     `json:"usage,omitempty"`
	Error   *StreamError   `json:"error,omitempty"`
}

// StreamChoice represents a streaming choice delta.
type StreamChoice struct {
	Index        int          `json:"index"`
	Delta        StreamDelta  `json:"delta"`
	FinishReason string       `json:"finish_reason"`
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
	Quota     float64 `json:"quota"`
	UsedQuota float64 `json:"used_quota"`
}

// Token represents an API key.
type Token struct {
	ID             int     `json:"id"`
	UserID         int     `json:"user_id"`
	Key            string  `json:"key"` // Raw key with "sk-" prefix (only on create)
	Name           string  `json:"name"`
	Status         int     `json:"status"`
	CreatedTime    int64   `json:"created_time"`
	ExpiredTime    int64   `json:"expired_time"`
	RemainQuota    int64   `json:"remain_quota"`
	UnlimitedQuota bool    `json:"unlimited_quota"`
	UsedQuota      int64   `json:"used_quota"`
	Models         *string `json:"models"`
	Subnet         *string `json:"subnet"`
}

// Channel represents an AI provider channel.
type Channel struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Type   int    `json:"type"`
	Key    string `json:"key"`
	Status int    `json:"status"`
	Models string `json:"models"`
}

// apiResponse is the standard QuantumClaw API response wrapper.
type apiResponse[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

// ==================== HTTP Client ====================

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

// ==================== OpenAI Compatible Relay ====================

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
// Returns a channel of StreamEvent; close when ctx is cancelled or stream ends.
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
	resp, err := c.doRequest(ctx, "GET", "/v1/models", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode models: %w", err)
	}
	return result.Data, nil
}

// ==================== User & Balance ====================

// GetSelfInfo returns the current user's information.
func (c *Client) GetSelfInfo(ctx context.Context) (map[string]interface{}, error) {
	resp, err := c.doRequest(ctx, "GET", "/api/user/self", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result apiResponse[map[string]interface{}]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode user info: %w", err)
	}
	return result.Data, nil
}

// GetBalance returns the current user's balance information.
func (c *Client) GetBalance(ctx context.Context) (*Balance, error) {
	resp, err := c.doRequest(ctx, "GET", "/api/user/self/balance", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result apiResponse[*Balance]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode balance: %w", err)
	}
	return result.Data, nil
}

// ==================== Token (API Key) Management ====================

// ListTokens returns all API keys for the current user.
func (c *Client) ListTokens(ctx context.Context, page int) ([]Token, error) {
	path := fmt.Sprintf("/api/token/?p=%d", page)
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result apiResponse[[]Token]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode tokens: %w", err)
	}
	return result.Data, nil
}

// CreateToken creates a new API key.
// Returns the Token with the raw key (including "sk-" prefix) in the Key field.
func (c *Client) CreateToken(ctx context.Context, name string, remainQuota int64, unlimitedQuota bool) (*Token, error) {
	body := map[string]interface{}{
		"name":            name,
		"remain_quota":    remainQuota,
		"unlimited_quota": unlimitedQuota,
	}
	resp, err := c.doRequest(ctx, "POST", "/api/token/", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result apiResponse[*Token]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode token: %w", err)
	}
	return result.Data, nil
}

// DeleteToken deletes an API key.
func (c *Client) DeleteToken(ctx context.Context, tokenID int) error {
	path := fmt.Sprintf("/api/token/%d", tokenID)
	resp, err := c.doRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// ==================== Channel Management ====================

// ==================== Quantum Resources ====================

// ListQuantumBackends returns all available quantum computing backends.
func (c *Client) ListQuantumBackends(ctx context.Context) ([]map[string]interface{}, error) {
	resp, err := c.doRequest(ctx, "GET", "/api/quantum/backends", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result apiResponse[[]map[string]interface{}]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode quantum backends: %w", err)
	}
	return result.Data, nil
}

// ListQuantumProviders returns quantum provider configuration stats.
func (c *Client) ListQuantumProviders(ctx context.Context) ([]map[string]interface{}, error) {
	resp, err := c.doRequest(ctx, "GET", "/api/quantum/providers", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result apiResponse[[]map[string]interface{}]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode quantum providers: %w", err)
	}
	return result.Data, nil
}

// SubmitQuantumTask submits a QASM quantum circuit for execution.
func (c *Client) SubmitQuantumTask(ctx context.Context, provider, qasm string, shots int, wait bool) (map[string]interface{}, error) {
	body := map[string]interface{}{
		"provider": provider,
		"qasm":     qasm,
		"shots":    shots,
		"wait":     wait,
	}
	resp, err := c.doRequest(ctx, "POST", "/api/quantum/submit", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result apiResponse[map[string]interface{}]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode quantum task: %w", err)
	}
	return result.Data, nil
}

// ==================== Async Tasks ====================

// ListTasks returns all async tasks for the current user.
func (c *Client) ListTasks(ctx context.Context) ([]map[string]interface{}, error) {
	resp, err := c.doRequest(ctx, "GET", "/api/task/", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result apiResponse[[]map[string]interface{}]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode tasks: %w", err)
	}
	return result.Data, nil
}

// GetTaskStatus returns the status of an async task by ID.
func (c *Client) GetTaskStatus(ctx context.Context, taskID string) (map[string]interface{}, error) {
	resp, err := c.doRequest(ctx, "GET", "/api/task/"+taskID, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result apiResponse[map[string]interface{}]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode task: %w", err)
	}
	return result.Data, nil
}

// ListChannels returns all channels.
func (c *Client) ListChannels(ctx context.Context) ([]Channel, error) {
	resp, err := c.doRequest(ctx, "GET", "/api/channel/", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result apiResponse[[]Channel]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode channels: %w", err)
	}
	return result.Data, nil
}
