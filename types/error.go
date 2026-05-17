package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// OpenAIError represents an OpenAI-style error response.
type OpenAIError struct {
	Message  string          `json:"message"`
	Type     string          `json:"type"`
	Param    string          `json:"param"`
	Code     any             `json:"code"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

// ClaudeError represents an Anthropic-style error response.
type ClaudeError struct {
	Type    string `json:"type,omitempty"`
	Message string `json:"message,omitempty"`
}

// ErrorType categorizes relay errors.
type ErrorType string

const (
	ErrorTypeQuantumClawError ErrorType = "quantumclaw_error"
	ErrorTypeOpenAIError      ErrorType = "openai_error"
	ErrorTypeClaudeError      ErrorType = "claude_error"
	ErrorTypeGeminiError      ErrorType = "gemini_error"
	ErrorTypeUpstreamError    ErrorType = "upstream_error"
)

// ErrorCode enumerates common error codes.
type ErrorCode string

const (
	ErrorCodeInvalidRequest           ErrorCode = "invalid_request"
	ErrorCodeCountTokenFailed         ErrorCode = "count_token_failed"
	ErrorCodeModelPriceError          ErrorCode = "model_price_error"
	ErrorCodeInvalidAPIType           ErrorCode = "invalid_api_type"
	ErrorCodeJSONMarshalFailed        ErrorCode = "json_marshal_failed"
	ErrorCodeDoRequestFailed          ErrorCode = "do_request_failed"
	ErrorCodeGetChannelFailed         ErrorCode = "get_channel_failed"
	ErrorCodeGenRelayInfoFailed       ErrorCode = "gen_relay_info_failed"
	ErrorCodeReadRequestBodyFailed    ErrorCode = "read_request_body_failed"
	ErrorCodeConvertRequestFailed     ErrorCode = "convert_request_failed"
	ErrorCodeAccessDenied             ErrorCode = "access_denied"
	ErrorCodeBadRequestBody           ErrorCode = "bad_request_body"
	ErrorCodeReadResponseBodyFailed   ErrorCode = "read_response_body_failed"
	ErrorCodeBadResponseStatusCode    ErrorCode = "bad_response_status_code"
	ErrorCodeBadResponse              ErrorCode = "bad_response"
	ErrorCodeEmptyResponse            ErrorCode = "empty_response"
	ErrorCodeModelNotFound            ErrorCode = "model_not_found"
	ErrorCodePromptBlocked            ErrorCode = "prompt_blocked"
	ErrorCodeQueryDataError           ErrorCode = "query_data_error"
	ErrorCodeUpdateDataError          ErrorCode = "update_data_error"
	ErrorCodeInsufficientUserQuota    ErrorCode = "insufficient_user_quota"
	ErrorCodePreConsumeTokenQuotaFail ErrorCode = "pre_consume_token_quota_failed"
)

// QuantumClawError is the unified error type for the relay system.
type QuantumClawError struct {
	Err            error
	RelayError     any
	skipRetry      bool
	recordErrorLog *bool
	errorType      ErrorType
	errorCode      ErrorCode
	StatusCode     int
	Metadata       json.RawMessage
}

// Unwrap enables errors.Is / errors.As.
func (e *QuantumClawError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func (e *QuantumClawError) GetErrorCode() ErrorCode {
	if e == nil {
		return ""
	}
	return e.errorCode
}

func (e *QuantumClawError) GetErrorType() ErrorType {
	if e == nil {
		return ""
	}
	return e.errorType
}

func (e *QuantumClawError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err == nil {
		return string(e.errorCode)
	}
	return e.Err.Error()
}

func (e *QuantumClawError) ErrorWithStatusCode() string {
	if e == nil {
		return ""
	}
	msg := e.Error()
	if e.StatusCode == 0 {
		return msg
	}
	if msg == "" {
		return fmt.Sprintf("status_code=%d", e.StatusCode)
	}
	return fmt.Sprintf("status_code=%d, %s", e.StatusCode, msg)
}

func (e *QuantumClawError) ToOpenAIError() OpenAIError {
	var result OpenAIError
	switch e.errorType {
	case ErrorTypeOpenAIError:
		if openAIError, ok := e.RelayError.(OpenAIError); ok {
			result = openAIError
		}
	case ErrorTypeClaudeError:
		if claudeError, ok := e.RelayError.(ClaudeError); ok {
			result = OpenAIError{
				Message: e.Error(),
				Type:    claudeError.Type,
				Param:   "",
				Code:    e.errorCode,
			}
		}
	default:
		result = OpenAIError{
			Message: e.Error(),
			Type:    string(e.errorType),
			Param:   "",
			Code:    e.errorCode,
		}
	}
	if result.Message == "" {
		result.Message = string(e.errorType)
	}
	return result
}

func (e *QuantumClawError) ToClaudeError() ClaudeError {
	var result ClaudeError
	switch e.errorType {
	case ErrorTypeOpenAIError:
		if openAIError, ok := e.RelayError.(OpenAIError); ok {
			result = ClaudeError{
				Message: e.Error(),
				Type:    fmt.Sprintf("%v", openAIError.Code),
			}
		}
	case ErrorTypeClaudeError:
		if claudeError, ok := e.RelayError.(ClaudeError); ok {
			result = claudeError
		}
	default:
		result = ClaudeError{
			Message: e.Error(),
			Type:    string(e.errorType),
		}
	}
	if result.Message == "" {
		result.Message = string(e.errorType)
	}
	return result
}

// QuantumClawErrorOptions is a functional option for creating errors.
type QuantumClawErrorOptions func(*QuantumClawError)

func NewError(err error, errorCode ErrorCode, ops ...QuantumClawErrorOptions) *QuantumClawError {
	var newErr *QuantumClawError
	if errors.As(err, &newErr) {
		for _, op := range ops {
			op(newErr)
		}
		return newErr
	}
	e := &QuantumClawError{
		Err:        err,
		RelayError: nil,
		errorType:  ErrorTypeQuantumClawError,
		StatusCode: http.StatusInternalServerError,
		errorCode:  errorCode,
	}
	for _, op := range ops {
		op(e)
	}
	return e
}

func NewOpenAIError(err error, errorCode ErrorCode, statusCode int, ops ...QuantumClawErrorOptions) *QuantumClawError {
	var newErr *QuantumClawError
	if errors.As(err, &newErr) {
		if newErr.RelayError == nil {
			newErr.RelayError = OpenAIError{
				Message: newErr.Error(),
				Type:    string(errorCode),
				Code:    errorCode,
			}
		}
		for _, op := range ops {
			op(newErr)
		}
		return newErr
	}
	openAIError := OpenAIError{
		Message: err.Error(),
		Type:    string(errorCode),
		Code:    errorCode,
	}
	return WithOpenAIError(openAIError, statusCode, ops...)
}

func WithOpenAIError(openAIError OpenAIError, statusCode int, ops ...QuantumClawErrorOptions) *QuantumClawError {
	code, ok := openAIError.Code.(string)
	if !ok {
		if openAIError.Code != nil {
			code = fmt.Sprintf("%v", openAIError.Code)
		} else {
			code = "unknown_error"
		}
	}
	if openAIError.Type == "" {
		openAIError.Type = "upstream_error"
	}
	e := &QuantumClawError{
		RelayError: openAIError,
		errorType:  ErrorTypeOpenAIError,
		StatusCode: statusCode,
		Err:        errors.New(openAIError.Message),
		errorCode:  ErrorCode(code),
	}
	// OpenRouter metadata
	if len(openAIError.Metadata) > 0 {
		openAIError.Message = fmt.Sprintf("%s (%s)", openAIError.Message, openAIError.Metadata)
		e.Metadata = openAIError.Metadata
		e.RelayError = openAIError
		e.Err = errors.New(openAIError.Message)
	}
	for _, op := range ops {
		op(e)
	}
	return e
}

func WithClaudeError(claudeError ClaudeError, statusCode int, ops ...QuantumClawErrorOptions) *QuantumClawError {
	if claudeError.Type == "" {
		claudeError.Type = "upstream_error"
	}
	e := &QuantumClawError{
		RelayError: claudeError,
		errorType:  ErrorTypeClaudeError,
		StatusCode: statusCode,
		Err:        errors.New(claudeError.Message),
		errorCode:  ErrorCode(claudeError.Type),
	}
	for _, op := range ops {
		op(e)
	}
	return e
}

func IsChannelError(err *QuantumClawError) bool {
	if err == nil {
		return false
	}
	return strings.HasPrefix(string(err.errorCode), "channel:")
}

func IsSkipRetryError(err *QuantumClawError) bool {
	if err == nil {
		return false
	}
	return err.skipRetry
}

func ErrOptionWithSkipRetry() QuantumClawErrorOptions {
	return func(e *QuantumClawError) {
		e.skipRetry = true
	}
}

func ErrOptionWithStatusCode(statusCode int) QuantumClawErrorOptions {
	return func(e *QuantumClawError) {
		e.StatusCode = statusCode
	}
}
