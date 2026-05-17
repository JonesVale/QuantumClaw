package dto

import (
	"encoding/json"
)

// SimpleResponse is a minimal response for simple relay cases.
type SimpleResponse struct {
	Usage *Usage `json:"usage"`
	Error any    `json:"error"`
}

// PromptTokensDetails holds detailed prompt token breakdowns.
type PromptTokensDetails struct {
	CachedTokens int `json:"cached_tokens"`
}

// Usage holds token usage information.
type Usage struct {
	PromptTokens     int                    `json:"prompt_tokens"`
	CompletionTokens int                    `json:"completion_tokens"`
	TotalTokens      int                    `json:"total_tokens"`
	PromptTokensDetails     *PromptTokensDetails     `json:"prompt_tokens_details,omitempty"`
	CompletionTokensDetails *CompletionTokensDetails `json:"completion_tokens_details,omitempty"`
}

// CompletionTokensDetails holds detailed completion token breakdowns.
type CompletionTokensDetails struct {
	ReasoningTokens          int `json:"reasoning_tokens"`
	AcceptedPredictionTokens int `json:"accepted_prediction_tokens"`
	RejectedPredictionTokens int `json:"rejected_prediction_tokens"`
}

// OpenAITextResponseChoice represents a single choice in a text response.
type OpenAITextResponseChoice struct {
	Index        int             `json:"index"`
	Message      Message         `json:"message"`
	FinishReason string          `json:"finish_reason"`
	Logprobs     json.RawMessage `json:"logprobs,omitempty"`
}

// OpenAITextResponse represents a non-streaming OpenAI chat completion response.
type OpenAITextResponse struct {
	ID      string                     `json:"id"`
	Model   string                     `json:"model"`
	Object  string                     `json:"object"`
	Created any                        `json:"created"`
	Choices []OpenAITextResponseChoice `json:"choices"`
	Error   any                        `json:"error,omitempty"`
	Usage   *Usage                     `json:"usage,omitempty"`
}

// OpenAIImageResponseData represents one image in a generation response.
type OpenAIImageResponseData struct {
	URL           string `json:"url,omitempty"`
	B64JSON       string `json:"b64_json,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

// OpenAIImageResponse represents an image generation response.
type OpenAIImageResponse struct {
	Created int                      `json:"created"`
	Data    []OpenAIImageResponseData `json:"data"`
}

// OpenAIEmbeddingResponseItem represents a single embedding vector.
type OpenAIEmbeddingResponseItem struct {
	Object    string    `json:"object"`
	Index     int       `json:"index"`
	Embedding []float64 `json:"embedding"`
}

// OpenAIEmbeddingResponse represents an embedding API response.
type OpenAIEmbeddingResponse struct {
	Object string                        `json:"object"`
	Data   []OpenAIEmbeddingResponseItem `json:"data"`
	Model  string                        `json:"model"`
	Usage  *Usage                        `json:"usage,omitempty"`
}

// Message represents a single message in a chat response.
type Message struct {
	Role             string             `json:"role"`
	Content          any                `json:"content"`
	ToolCalls        []ToolCall         `json:"tool_calls,omitempty"`
	ToolCallID       string             `json:"tool_call_id,omitempty"`
	Name             string             `json:"name,omitempty"`
	ReasoningContent string             `json:"reasoning_content,omitempty"`
	Audio            *MessageAudio      `json:"audio,omitempty"`
}

// ToolCall represents a function call in a response.
type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
}

// ToolCallFunction represents the function details in a tool call.
type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// MessageAudio represents audio output in a message.
type MessageAudio struct {
	ID string `json:"id"`
}

// ChatCompletionsStreamResponseChoice represents a streaming choice.
type ChatCompletionsStreamResponseChoice struct {
	Delta        ChatCompletionsStreamResponseChoiceDelta `json:"delta,omitempty"`
	Logprobs     *any                                     `json:"logprobs"`
	FinishReason *string                                  `json:"finish_reason"`
	Index        int                                      `json:"index"`
}

// ChatCompletionsStreamResponseChoiceDelta represents the delta content in a streaming response.
type ChatCompletionsStreamResponseChoiceDelta struct {
	Content          *string            `json:"content,omitempty"`
	ReasoningContent *string            `json:"reasoning_content,omitempty"`
	Reasoning        *string            `json:"reasoning,omitempty"`
	Role             string             `json:"role,omitempty"`
	ToolCalls        []ToolCallResponse `json:"tool_calls,omitempty"`
}

func (c *ChatCompletionsStreamResponseChoiceDelta) SetContentString(s string) {
	c.Content = &s
}

func (c *ChatCompletionsStreamResponseChoiceDelta) GetContentString() string {
	if c.Content != nil {
		return *c.Content
	}
	return ""
}

// ToolCallResponse represents a tool call in streaming mode.
type ToolCallResponse struct {
	ID       string                      `json:"id,omitempty"`
	Type     string                      `json:"type,omitempty"`
	Index    *int                        `json:"index,omitempty"`
	Function *ToolCallResponseFunction   `json:"function,omitempty"`
}

// ToolCallResponseFunction represents the function fields in a streaming tool call.
type ToolCallResponseFunction struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

// ChatCompletionsStreamResponse is the top-level SSE streamed response.
type ChatCompletionsStreamResponse struct {
	ID      string                                `json:"id"`
	Object  string                                `json:"object"`
	Created int64                                 `json:"created"`
	Model   string                                `json:"model"`
	Choices []ChatCompletionsStreamResponseChoice `json:"choices"`
	Usage   *Usage                                `json:"usage,omitempty"`
}
