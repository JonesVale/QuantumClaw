package web_shared

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ── SSE Event Types ──

// SSEEvent represents a single Server-Sent Event from a streaming response.
type SSEEvent struct {
	Content string
	Done    bool
	Error   error
}

// StreamResult is the final parsed response for non-streaming mode.
type StreamResult struct {
	Content      string
	Usage        *ModelUsage
	FinishReason string
}

type ModelUsage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// ── Stream Parsers ──

// StreamSSE parses a Server-Sent Events (SSE) stream from an HTTP response.
// It extracts content using the specified JSON response path.
// responsePath supports dot notation, e.g. "message.content.parts.0" or "choices.0.delta.content".
func StreamSSE(ctx context.Context, resp *http.Response, responsePath string) chan SSEEvent {
	ch := make(chan SSEEvent, 64)
	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // 1MB buffer

		for scanner.Scan() {
			line := scanner.Text()

			select {
			case <-ctx.Done():
				ch <- SSEEvent{Error: ctx.Err()}
				return
			default:
			}

			// Skip empty lines and comments
			if line == "" || strings.HasPrefix(line, ":") {
				continue
			}

			// Parse SSE format: "event: ..." or "data: ..."
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")
				if data == "[DONE]" {
					ch <- SSEEvent{Done: true}
					return
				}
				if data == "" {
					continue
				}

				// Try to extract content from the JSON using the response path
				content := extractByPath(data, responsePath)
				ch <- SSEEvent{Content: content}
			}
		}

		if err := scanner.Err(); err != nil {
			ch <- SSEEvent{Error: fmt.Errorf("sse scanner error: %w", err)}
		}
		ch <- SSEEvent{Done: true}
	}()
	return ch
}

// StreamPoll reads a complete non-streaming response body and extracts content.
// Used for providers that return the full response at once.
func StreamPoll(ctx context.Context, resp *http.Response, responsePath string) (*StreamResult, error) {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("poll read error: %w", err)
	}

	content := extractByPath(string(body), responsePath)

	// Try to extract usage info too
	usage := extractUsage(string(body))

	return &StreamResult{
		Content: content,
		Usage:   usage,
	}, nil
}

// ── JSON Path Extraction ──

// extractByPath navigates a JSON response using dot-notation path.
// Examples:
//
//	"choices.0.message.content"
//	"message.content.parts.0"
//	"content"
func extractByPath(jsonStr string, path string) string {
	if jsonStr == "" || path == "" {
		return jsonStr
	}

	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return jsonStr // fallback: return raw data
	}

	parts := strings.Split(path, ".")
	current := data
	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			if next, ok := v[part]; ok {
				current = next
			} else {
				return "" // path not found
			}
		case []interface{}:
			idx := 0
			if _, err := fmt.Sscanf(part, "%d", &idx); err == nil && idx >= 0 && idx < len(v) {
				current = v[idx]
			} else {
				return "" // array index out of bounds
			}
		default:
			return "" // can't navigate further
		}
	}

	// Convert result to string
	switch v := current.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case float64:
		return fmt.Sprintf("%v", v)
	default:
		data, _ := json.Marshal(v)
		return string(data)
	}
}

// extractUsage attempts to parse token usage from a JSON response.
func extractUsage(jsonStr string) *ModelUsage {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil
	}

	usage := &ModelUsage{}
	// Try standard OpenAI format: usage.prompt_tokens, usage.completion_tokens
	if u, ok := data["usage"].(map[string]interface{}); ok {
		if pt, ok := u["prompt_tokens"].(float64); ok {
			usage.PromptTokens = int(pt)
		}
		if ct, ok := u["completion_tokens"].(float64); ok {
			usage.CompletionTokens = int(ct)
		}
		if tt, ok := u["total_tokens"].(float64); ok {
			usage.TotalTokens = int(tt)
		}
	}

	if usage.TotalTokens == 0 {
		return nil // no usage data
	}
	return usage
}

// FormatSSE formats content as an SSE data frame for the relay response.
func FormatSSE(content string) string {
	if content == "" {
		return ""
	}
	return fmt.Sprintf("data: %s\n\n", content)
}

// FormatSSEDone returns the standard [DONE] SSE marker.
func FormatSSEDone() string {
	return "data: [DONE]\n\n"
}
