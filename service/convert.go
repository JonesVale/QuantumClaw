package service

import (
	"encoding/json"

	"github.com/quantumclaw/quantumclaw/dto"
)

// ClaudeToOpenAIRequest converts a Claude-format request to an OpenAI-format request.
func ClaudeToOpenAIRequest(request dto.ClaudeRequest, info any) (*dto.OpenAIRequest, error) {
	// Convert Claude messages to raw JSON for OpenAI request
	msgBytes, err := json.Marshal(request.Messages)
	if err != nil {
		return nil, err
	}

	return &dto.OpenAIRequest{
		Model:       request.Model,
		Messages:    msgBytes,
		MaxTokens:   request.MaxTokens,
		Stream:      request.Stream,
		Temperature: request.Temperature,
	}, nil
}

// OpenAIToClaudeRequest converts an OpenAI-format request to a Claude-format request.
func OpenAIToClaudeRequest(request *dto.OpenAIRequest) (*dto.ClaudeRequest, error) {
	// Parse raw messages into ClaudeMessage format
	var msgs []dto.Message
	if err := json.Unmarshal(request.Messages, &msgs); err != nil {
		// Try to parse as array of ClaudeMessage directly
		var claudeMsgs []dto.ClaudeMessage
		if err2 := json.Unmarshal(request.Messages, &claudeMsgs); err2 != nil {
			return nil, err
		}
		return &dto.ClaudeRequest{
			Model:     request.Model,
			Messages:  claudeMsgs,
			MaxTokens: 4096,
			Stream:    request.Stream,
		}, nil
	}

	systemMsg := ""
	cleanMessages := make([]dto.ClaudeMessage, 0)
	for _, msg := range msgs {
		if msg.Role == "system" {
			if s, ok := msg.Content.(string); ok {
				systemMsg = s
			}
		} else {
			contentBytes, _ := json.Marshal(msg.Content)
			cleanMessages = append(cleanMessages, dto.ClaudeMessage{
				Role:    msg.Role,
				Content: contentBytes,
			})
		}
	}

	claudeReq := &dto.ClaudeRequest{
		Model:     request.Model,
		Messages:  cleanMessages,
		MaxTokens: 4096,
		Stream:    request.Stream,
	}

	if systemMsg != "" {
		claudeReq.System = json.RawMessage(`"` + systemMsg + `"`)
	}

	// Set temperature if present
	if request.Temperature != nil {
		claudeReq.Temperature = request.Temperature
	}

	return claudeReq, nil
}
