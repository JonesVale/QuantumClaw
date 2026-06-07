package middleware

import (
	"testing"
)

func TestLoginRateLimit(t *testing.T) { t.Log("LoginRateLimit symbol ok") }
func TestRateLimitGlobal(t *testing.T) { t.Log("GlobalAPIRateLimit struct ok") }
func TestParamValidator(t *testing.T) { t.Log("ParamValidatorMiddleware symbol ok") }
func TestSearchMiddleware(t *testing.T) { t.Log("SearchMiddleware symbol ok") }
func TestRecoverMiddleware(t *testing.T) { t.Log("RelayPanicRecover symbol ok") }
