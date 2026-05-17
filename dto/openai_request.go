package dto

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/types"
)

// OpenAIRequest is a general-purpose OpenAI-compatible chat completion request.
type OpenAIRequest struct {
	Model               string          `json:"model"`
	Messages            json.RawMessage `json:"messages"`
	Stream              bool            `json:"stream"`
	MaxTokens           int             `json:"max_tokens,omitempty"`
	MaxCompletionTokens *int            `json:"max_completion_tokens,omitempty"`
	Temperature         *float64        `json:"temperature,omitempty"`
	TopP                *float64        `json:"top_p,omitempty"`
	TopK                int             `json:"top_k,omitempty"`
	FrequencyPenalty    *float64        `json:"frequency_penalty,omitempty"`
	PresencePenalty     *float64        `json:"presence_penalty,omitempty"`
	N                   int             `json:"n,omitempty"`
	Stop                any             `json:"stop,omitempty"`
	Tools               json.RawMessage `json:"tools,omitempty"`
	ToolChoice          any             `json:"tool_choice,omitempty"`
	ParallelToolCalls   *bool           `json:"parallel_tool_calls,omitempty"`
	ResponseFormat      *ResponseFormat `json:"response_format,omitempty"`
	Seed                float64         `json:"seed,omitempty"`
	User                string          `json:"user,omitempty"`
	Metadata            json.RawMessage `json:"metadata,omitempty"`
	Modalities          []string        `json:"modalities,omitempty"`
	Audio               *AudioParam     `json:"audio,omitempty"`
	StreamOptions       *StreamOptions  `json:"stream_options,omitempty"`
	Store               *bool           `json:"store,omitempty"`
	ReasoningEffort     *string         `json:"reasoning_effort,omitempty"`
	ServiceTier         *string         `json:"service_tier,omitempty"`
	Input               any             `json:"input,omitempty"`
}

type ResponseFormat struct {
	Type       string          `json:"type,omitempty"`
	JSONSchema *JSONSchema     `json:"json_schema,omitempty"`
}

type JSONSchema struct {
	Description string                 `json:"description,omitempty"`
	Name        string                 `json:"name"`
	Schema      map[string]interface{} `json:"schema,omitempty"`
	Strict      *bool                  `json:"strict,omitempty"`
}

type AudioParam struct {
	Voice  string `json:"voice,omitempty"`
	Format string `json:"format,omitempty"`
}

type StreamOptions struct {
	IncludeUsage bool `json:"include_usage,omitempty"`
}

func (r *OpenAIRequest) GetTokenCountMeta() *types.TokenCountMeta {
	return &types.TokenCountMeta{
		TokenType: types.TokenTypeTokenizer,
	}
}

func (r *OpenAIRequest) IsStream(c *gin.Context) bool {
	return r.Stream
}

func (r *OpenAIRequest) SetModelName(modelName string) {
	if modelName != "" {
		r.Model = modelName
	}
}

// ImageRequest represents an OpenAI image generation request.
type ImageRequest struct {
	Model          string `json:"model"`
	Prompt         string `json:"prompt"`
	N              int    `json:"n,omitempty"`
	Quality        string `json:"quality,omitempty"`
	Size           string `json:"size,omitempty"`
	Style          string `json:"style,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
	User           string `json:"user,omitempty"`
}
