package dto

import (
	"encoding/json"

	"github.com/quantumclaw/quantumclaw/types"
)

// OpenAIErrorWithStatusCode wraps an OpenAIError with an HTTP status code.
type OpenAIErrorWithStatusCode struct {
	Error      types.OpenAIError `json:"error"`
	StatusCode int               `json:"status_code"`
	LocalError bool
}

// GeneralErrorResponse is a best-effort parser for upstream error bodies.
type GeneralErrorResponse struct {
	Error    json.RawMessage `json:"error"`
	Message  string          `json:"message"`
	Msg      string          `json:"msg"`
	Err      string          `json:"err"`
	ErrorMsg string          `json:"error_msg"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
	Detail   string          `json:"detail,omitempty"`
	Header   struct {
		Message string `json:"message"`
	} `json:"header"`
	Response struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	} `json:"response"`
}

func (e GeneralErrorResponse) TryToOpenAIError() *types.OpenAIError {
	var openAIError types.OpenAIError
	if len(e.Error) > 0 {
		if err := json.Unmarshal(e.Error, &openAIError); err == nil && openAIError.Message != "" {
			return &openAIError
		}
	}
	return nil
}

func (e GeneralErrorResponse) ToMessage() string {
	if len(e.Error) > 0 {
		var openAIError types.OpenAIError
		if err := json.Unmarshal(e.Error, &openAIError); err == nil && openAIError.Message != "" {
			return openAIError.Message
		}
		var msg string
		if err := json.Unmarshal(e.Error, &msg); err == nil && msg != "" {
			return msg
		}
		return string(e.Error)
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Msg != "" {
		return e.Msg
	}
	if e.Err != "" {
		return e.Err
	}
	if e.ErrorMsg != "" {
		return e.ErrorMsg
	}
	if e.Detail != "" {
		return e.Detail
	}
	if e.Header.Message != "" {
		return e.Header.Message
	}
	if e.Response.Error.Message != "" {
		return e.Response.Error.Message
	}
	return ""
}
