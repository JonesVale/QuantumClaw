package dto

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/types"
)

// ClaudeRequest represents an Anthropic Claude API request.
type ClaudeRequest struct {
	Model             string                `json:"model"`
	Messages          []ClaudeMessage       `json:"messages"`
	System            json.RawMessage       `json:"system,omitempty"`
	MaxTokens         int                   `json:"max_tokens"`
	Metadata          *ClaudeMetadata       `json:"metadata,omitempty"`
	StopSequences     []string              `json:"stop_sequences,omitempty"`
	Stream            bool                  `json:"stream"`
	Temperature       *float64              `json:"temperature,omitempty"`
	TopP              *float64              `json:"top_p,omitempty"`
	TopK              *int                  `json:"top_k,omitempty"`
	Tools             []ClaudeTool          `json:"tools,omitempty"`
	ToolChoice        any                   `json:"tool_choice,omitempty"`
	Thinking          *ClaudeThinkingConfig `json:"thinking,omitempty"`
}

// ClaudeMessage represents a single message in Claude format.
type ClaudeMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

// ClaudeMetadata holds optional metadata for Claude requests.
type ClaudeMetadata struct {
	UserID string `json:"user_id"`
}

// ClaudeTool represents a tool definition for Claude.
type ClaudeTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"input_schema,omitempty"`
}

// ClaudeThinkingConfig configures extended thinking for Claude.
type ClaudeThinkingConfig struct {
	Type      string `json:"type"`
	BudgetTokens int `json:"budget_tokens"`
}

func (r *ClaudeRequest) GetTokenCountMeta() *types.TokenCountMeta {
	return &types.TokenCountMeta{
		TokenType: types.TokenTypeTokenizer,
	}
}

func (r *ClaudeRequest) IsStream(c *gin.Context) bool {
	return r.Stream
}

func (r *ClaudeRequest) SetModelName(modelName string) {
	if modelName != "" {
		r.Model = modelName
	}
}

// ClaudeResponse represents a Claude API response.
type ClaudeResponse struct {
	ID           string            `json:"id"`
	Type         string            `json:"type"`
	Role         string            `json:"role"`
	Content      []ClaudeContentBlock `json:"content"`
	Model        string            `json:"model"`
	StopReason   *string           `json:"stop_reason,omitempty"`
	StopSequence *string           `json:"stop_sequence,omitempty"`
	Usage        *ClaudeUsage      `json:"usage,omitempty"`
}

// ClaudeContentBlock represents a content block in Claude responses.
type ClaudeContentBlock struct {
	Type   string          `json:"type"`
	Text   string          `json:"text,omitempty"`
	ID     string          `json:"id,omitempty"`
	Name   string          `json:"name,omitempty"`
	Input  json.RawMessage `json:"input,omitempty"`
	Source *ClaudeSource   `json:"source,omitempty"`
	Thinking string        `json:"thinking,omitempty"`
	Signature string       `json:"signature,omitempty"`
}

// ClaudeSource represents a media source in Claude.
type ClaudeSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

// ClaudeUsage holds token usage for Claude responses.
type ClaudeUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
	ThinkingTokens           int `json:"thinking_tokens,omitempty"`
}
