package common

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
)

type ErrorCode string

const (
	ErrCodeUnknown           ErrorCode = "UNKNOWN"
	ErrCodeDatabase          ErrorCode = "DATABASE"
	ErrCodeRedis             ErrorCode = "REDIS"
	ErrCodeValidation        ErrorCode = "VALIDATION"
	ErrCodeAuthentication    ErrorCode = "AUTH"
	ErrCodeAuthorization     ErrorCode = "FORBIDDEN"
	ErrCodeNotFound          ErrorCode = "NOT_FOUND"
	ErrCodeRateLimit        ErrorCode = "RATE_LIMIT"
	ErrCodeQuotaExceeded    ErrorCode = "QUOTA_EXCEEDED"
	ErrCodeChannelUnavailable ErrorCode = "CHANNEL_UNAVAILABLE"
	ErrCodeTimeout          ErrorCode = "TIMEOUT"
	ErrCodeInternal         ErrorCode = "INTERNAL"
)

type AppError struct {
	Code       ErrorCode   `json:"code"`
	Message    string      `json:"message"`
	Detail     string      `json:"detail,omitempty"`
	HTTPStatus int         `json:"-"`
	Timestamp  time.Time   `json:"timestamp"`
	Err        error       `json:"-"`
}

var errorMessages = map[ErrorCode]string{
	ErrCodeUnknown:           "An unknown error occurred",
	ErrCodeDatabase:          "Database operation failed",
	ErrCodeRedis:            "Cache operation failed",
	ErrCodeValidation:       "Validation failed",
	ErrCodeAuthentication:   "Authentication required",
	ErrCodeAuthorization:     "Access denied",
	ErrCodeNotFound:         "Resource not found",
	ErrCodeRateLimit:        "Rate limit exceeded",
	ErrCodeQuotaExceeded:    "Quota exceeded",
	ErrCodeChannelUnavailable: "Channel unavailable",
	ErrCodeTimeout:          "Operation timeout",
	ErrCodeInternal:         "Internal server error",
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) WithDetail(detail string) *AppError {
	e.Detail = detail
	return e
}

func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	return e
}

func NewAppError(code ErrorCode, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Timestamp:  time.Now(),
	}
}

func NewDatabaseError(err error) *AppError {
	return NewAppError(ErrCodeDatabase, errorMessages[ErrCodeDatabase], http.StatusInternalServerError).WithError(err)
}

func NewRedisError(err error) *AppError {
	return NewAppError(ErrCodeRedis, errorMessages[ErrCodeRedis], http.StatusInternalServerError).WithError(err)
}

func NewValidationError(message string) *AppError {
	return NewAppError(ErrCodeValidation, message, http.StatusBadRequest)
}

func NewAuthError(message string) *AppError {
	if message == "" {
		message = errorMessages[ErrCodeAuthentication]
	}
	return NewAppError(ErrCodeAuthentication, message, http.StatusUnauthorized)
}

func NewForbiddenError() *AppError {
	return NewAppError(ErrCodeAuthorization, errorMessages[ErrCodeAuthorization], http.StatusForbidden)
}

func NewNotFoundError(resource string) *AppError {
	return NewAppError(ErrCodeNotFound, fmt.Sprintf("%s not found", resource), http.StatusNotFound)
}

func NewRateLimitError() *AppError {
	return NewAppError(ErrCodeRateLimit, errorMessages[ErrCodeRateLimit], http.StatusTooManyRequests)
}

func NewQuotaExceededError() *AppError {
	return NewAppError(ErrCodeQuotaExceeded, errorMessages[ErrCodeQuotaExceeded], http.StatusPaymentRequired)
}

func NewChannelUnavailableError(channel string) *AppError {
	return NewAppError(ErrCodeChannelUnavailable, fmt.Sprintf("Channel %s is unavailable", channel), http.StatusServiceUnavailable)
}

func NewTimeoutError() *AppError {
	return NewAppError(ErrCodeTimeout, errorMessages[ErrCodeTimeout], http.StatusGatewayTimeout)
}

func NewInternalError(err error) *AppError {
	return NewAppError(ErrCodeInternal, errorMessages[ErrCodeInternal], http.StatusInternalServerError).WithError(err)
}

func RecoverAndLog() {
	if r := recover(); r != nil {
		stack := debug.Stack()
		logger.SysError(fmt.Sprintf("Panic recovered: %v\n%s", r, string(stack)))
	}
}

func SafeGo(fn func(), onError func(err error)) {
	go func() {
		defer RecoverAndLog()
		defer func() {
			if r := recover(); r != nil {
				onError(fmt.Errorf("panic: %v", r))
			}
		}()
		fn()
	}()
}

type RetryConfig struct {
	MaxAttempts int
	Interval    time.Duration
	MaxInterval time.Duration
	Multiplier  float64
}

var DefaultRetryConfig = RetryConfig{
	MaxAttempts: 3,
	Interval:    100 * time.Millisecond,
	MaxInterval: 5 * time.Second,
	Multiplier:  2.0,
}

func Retry[T any](fn func() (T, error), config *RetryConfig) (T, error) {
	if config == nil {
		config = &DefaultRetryConfig
	}

	var lastErr error
	interval := config.Interval

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err

		if attempt < config.MaxAttempts {
			time.Sleep(interval)
			interval = time.Duration(float64(interval) * config.Multiplier)
			if interval > config.MaxInterval {
				interval = config.MaxInterval
			}
		}
	}

	var zero T
	return zero, fmt.Errorf("retry failed after %d attempts: %w", config.MaxAttempts, lastErr)
}

type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

type CircuitBreaker struct {
	mu           sync.Mutex
	state        CircuitState
	failures     int
	threshold    int
	resetTimeout time.Duration
	lastFailure  time.Time
}

func NewCircuitBreaker(threshold int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:        CircuitClosed,
		threshold:    threshold,
		resetTimeout: resetTimeout,
	}
}

func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.state = CircuitHalfOpen
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	default:
		return false
	}
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0
	cb.state = CircuitClosed
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailure = time.Now()

	if cb.failures >= cb.threshold {
		cb.state = CircuitOpen
	}
}

func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

type ErrorHandler struct {
	circuitBreakers map[string]*CircuitBreaker
	mu              sync.RWMutex
}

var globalErrorHandler = &ErrorHandler{
	circuitBreakers: make(map[string]*CircuitBreaker),
}

func GetErrorHandler() *ErrorHandler {
	return globalErrorHandler
}

func (h *ErrorHandler) GetCircuitBreaker(name string) *CircuitBreaker {
	h.mu.RLock()
	cb, exists := h.circuitBreakers[name]
	h.mu.RUnlock()

	if exists {
		return cb
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if cb, exists = h.circuitBreakers[name]; exists {
		return cb
	}

	cb = NewCircuitBreaker(5, 30*time.Second)
	h.circuitBreakers[name] = cb
	return cb
}

func (h *ErrorHandler) Execute(name string, fn func() error) error {
	cb := h.GetCircuitBreaker(name)

	if !cb.Allow() {
		return NewChannelUnavailableError(name)
	}

	err := fn()
	if err != nil {
		cb.RecordFailure()
		return err
	}

	cb.RecordSuccess()
	return nil
}

func ClassifyError(err error) ErrorCode {
	if err == nil {
		return ErrCodeUnknown
	}

	errStr := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errStr, "database") || strings.Contains(errStr, "sql") || strings.Contains(errStr, "gorm"):
		return ErrCodeDatabase
	case strings.Contains(errStr, "redis") || strings.Contains(errStr, "cache"):
		return ErrCodeRedis
	case strings.Contains(errStr, "validation") || strings.Contains(errStr, "invalid"):
		return ErrCodeValidation
	case strings.Contains(errStr, "auth") || strings.Contains(errStr, "unauthorized"):
		return ErrCodeAuthentication
	case strings.Contains(errStr, "forbidden") || strings.Contains(errStr, "permission"):
		return ErrCodeAuthorization
	case strings.Contains(errStr, "not found") || strings.Contains(errStr, "不存在"):
		return ErrCodeNotFound
	case strings.Contains(errStr, "rate limit") || strings.Contains(errStr, "too many"):
		return ErrCodeRateLimit
	case strings.Contains(errStr, "quota") || strings.Contains(errStr, "额度"):
		return ErrCodeQuotaExceeded
	case strings.Contains(errStr, "timeout") || strings.Contains(errStr, "超时"):
		return ErrCodeTimeout
	default:
		return ErrCodeInternal
	}
}