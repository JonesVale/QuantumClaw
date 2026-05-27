package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/logger"
)

//  Model parameter constraints 
// These are sensible defaults; model-specific overrides come from ModelMetadata.

type modelParamBounds struct {
	MaxTokensMax int     `json:"max_tokens_max"`
	TemperatureMin float64 `json:"temperature_min"`
	TemperatureMax float64 `json:"temperature_max"`
	TopPMin       float64 `json:"top_p_min"`
	TopPMax       float64 `json:"top_p_max"`
}

// defaultBounds applies to all models unless overridden.
var defaultBounds = modelParamBounds{
	MaxTokensMax:   131072, // 128K
	TemperatureMin: 0,
	TemperatureMax: 2.0,
	TopPMin:        0,
	TopPMax:        1.0,
}

// modelOverrides contains model-specific parameter overrides.
// Key is normalized model name (lowercase, hyphens).
var modelOverrides = map[string]modelParamBounds{
	// OpenAI o-series reasoning models: temperature must be 1
	"o1":                {TemperatureMin: 1, TemperatureMax: 1, MaxTokensMax: 100000, TopPMin: 0, TopPMax: 1},
	"o1-2024-12-17":     {TemperatureMin: 1, TemperatureMax: 1, MaxTokensMax: 100000, TopPMin: 0, TopPMax: 1},
	"o1-preview":        {TemperatureMin: 1, TemperatureMax: 1, MaxTokensMax: 32768, TopPMin: 0, TopPMax: 1},
	"o1-preview-2024-09-12": {TemperatureMin: 1, TemperatureMax: 1, MaxTokensMax: 32768, TopPMin: 0, TopPMax: 1},
	"o1-mini":           {TemperatureMin: 1, TemperatureMax: 1, MaxTokensMax: 65536, TopPMin: 0, TopPMax: 1},
	"o1-mini-2024-09-12": {TemperatureMin: 1, TemperatureMax: 1, MaxTokensMax: 65536, TopPMin: 0, TopPMax: 1},
	"o3-mini":           {TemperatureMin: 1, TemperatureMax: 1, MaxTokensMax: 100000, TopPMin: 0, TopPMax: 1},
	"o3-mini-2025-01-31": {TemperatureMin: 1, TemperatureMax: 1, MaxTokensMax: 100000, TopPMin: 0, TopPMax: 1},
	// Claude models: temperature 0-1
	"claude-3-5-sonnet":  {TemperatureMin: 0, TemperatureMax: 1, MaxTokensMax: 8192, TopPMin: 0, TopPMax: 1},
	"claude-3-opus":      {TemperatureMin: 0, TemperatureMax: 1, MaxTokensMax: 4096, TopPMin: 0, TopPMax: 1},
	"claude-3-haiku":     {TemperatureMin: 0, TemperatureMax: 1, MaxTokensMax: 4096, TopPMin: 0, TopPMax: 1},
	// DeepSeek
	"deepseek-chat":      {MaxTokensMax: 8192, TemperatureMin: 0, TemperatureMax: 2, TopPMin: 0, TopPMax: 1},
	"deepseek-reasoner":  {MaxTokensMax: 8192, TemperatureMin: 0, TemperatureMax: 2, TopPMin: 0, TopPMax: 1},
	// Gemini
	"gemini-1-5-pro":     {TemperatureMin: 0, TemperatureMax: 2, MaxTokensMax: 8192, TopPMin: 0, TopPMax: 1},
	"gemini-1-5-flash":   {TemperatureMin: 0, TemperatureMax: 2, MaxTokensMax: 8192, TopPMin: 0, TopPMax: 1},
}

// ParamValidationConfig holds the toggle state for param validation.
type ParamValidationConfig struct {
	Enabled bool `json:"enabled"`
	// AutoFix: when true, silently fix out-of-range parameters.
	// When false, return 400 with explanation.
	AutoFix bool `json:"auto_fix"`
	// LogOnly: when true, only log violations without fixing.
	LogOnly bool `json:"log_only"`
}

var ParamValidation = ParamValidationConfig{
	Enabled: true,
	AutoFix: true,
	LogOnly: false,
}

//  OpenAI chat completion request (partial, only validation-relevant fields) 

type validationRequest struct {
	Model       string    `json:"model"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature *float64  `json:"temperature,omitempty"`
	TopP        *float64  `json:"top_p,omitempty"`
}

// getBounds returns parameter bounds for the given model name.
func getBounds(model string) modelParamBounds {
	name := strings.ToLower(model)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")

	// Exact match first
	if b, ok := modelOverrides[name]; ok {
		return b
	}
	// Prefix match for versioned models (e.g., "o1-2024-12-17" ?"o1")
	// Try progressively shorter prefixes
	parts := strings.Split(name, "-")
	for i := len(parts) - 1; i > 0; i-- {
		prefix := strings.Join(parts[:i], "-")
		if b, ok := modelOverrides[prefix]; ok {
			return b
		}
	}
	return defaultBounds
}

// clampInt clamps val to [min, max].
func clampInt(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// clampFloat clamps val to [min, max].
func clampFloat(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// ParamValidatorMiddleware validates and optionally fixes model request parameters.
// It should be placed BEFORE the relay chain (after auth, before distribute).
func ParamValidatorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ParamValidation.Enabled {
			c.Next()
			return
		}

		// Only intercept chat/completions and completions endpoints
		path := c.Request.URL.Path
		if !strings.HasSuffix(path, "/chat/completions") && !strings.HasSuffix(path, "/completions") {
			c.Next()
			return
		}

		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.Warn(context.Background(), fmt.Sprintf("ParamValidator: failed to read body: %v", err))
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"message": "Failed to read request body",
					"type":    "invalid_request_error",
				},
			})
			return
		}
		// Restore body for downstream handlers
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		var req validationRequest
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			// Not a JSON body we understand, skip validation
			c.Next()
			return
		}

		if req.Model == "" {
			c.Next()
			return
		}

		bounds := getBounds(req.Model)
		var corrections []string

		// Validate max_tokens
		if req.MaxTokens > 0 && req.MaxTokens > bounds.MaxTokensMax {
			oldVal := req.MaxTokens
			if ParamValidation.AutoFix && !ParamValidation.LogOnly {
				req.MaxTokens = bounds.MaxTokensMax
				corrections = append(corrections, fmt.Sprintf("max_tokens: %d �?%d", oldVal, req.MaxTokens))
			} else if !ParamValidation.LogOnly {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error": gin.H{
						"message": fmt.Sprintf("max_tokens exceeds maximum of %d for model %s", bounds.MaxTokensMax, req.Model),
						"type":    "invalid_request_error",
					},
				})
				return
			}
		}

		// Validate temperature
		if req.Temperature != nil {
			if *req.Temperature < bounds.TemperatureMin || *req.Temperature > bounds.TemperatureMax {
				oldVal := *req.Temperature
				if ParamValidation.AutoFix && !ParamValidation.LogOnly {
					*req.Temperature = clampFloat(*req.Temperature, bounds.TemperatureMin, bounds.TemperatureMax)
					corrections = append(corrections, fmt.Sprintf("temperature: %.2f �?%.2f", oldVal, *req.Temperature))
				} else if !ParamValidation.LogOnly {
					c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
						"error": gin.H{
							"message": fmt.Sprintf("temperature %.2f out of range [%.1f, %.1f] for model %s",
								*req.Temperature, bounds.TemperatureMin, bounds.TemperatureMax, req.Model),
							"type": "invalid_request_error",
						},
					})
					return
				}
			}
		}

		// Validate top_p
		if req.TopP != nil {
			if *req.TopP < bounds.TopPMin || *req.TopP > bounds.TopPMax {
				oldVal := *req.TopP
				if ParamValidation.AutoFix && !ParamValidation.LogOnly {
					*req.TopP = clampFloat(*req.TopP, bounds.TopPMin, bounds.TopPMax)
					corrections = append(corrections, fmt.Sprintf("top_p: %.2f �?%.2f", oldVal, *req.TopP))
				} else if !ParamValidation.LogOnly {
					c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
						"error": gin.H{
							"message": fmt.Sprintf("top_p %.2f out of range [%.1f, %.1f] for model %s",
								*req.TopP, bounds.TopPMin, bounds.TopPMax, req.Model),
							"type": "invalid_request_error",
						},
					})
					return
				}
			}
		}

		// If we auto-fixed something, write the modified body back
		if len(corrections) > 0 {
			logger.Info(context.Background(), fmt.Sprintf("ParamValidator [%s]: %s", req.Model, strings.Join(corrections, "; ")))

			if !ParamValidation.LogOnly {
				// Rebuild the body with corrected values
				var rawMap map[string]interface{}
				if err := json.Unmarshal(bodyBytes, &rawMap); err == nil {
					if req.MaxTokens > 0 {
						rawMap["max_tokens"] = req.MaxTokens
					}
					if req.Temperature != nil {
						rawMap["temperature"] = req.Temperature
					}
					if req.TopP != nil {
						rawMap["top_p"] = req.TopP
					}
					newBody, err := json.Marshal(rawMap)
					if err == nil {
						c.Request.Body = io.NopCloser(bytes.NewBuffer(newBody))
						c.Request.ContentLength = int64(len(newBody))
					}
				}
			}
		}

		c.Next()
	}
}

// SetParamValidationConfig updates the validation config at runtime.
func SetParamValidationConfig(cfg ParamValidationConfig) {
	ParamValidation = cfg
}

// GetParamBounds returns parameter bounds for a given model (for frontend display).
func GetParamBounds(model string) modelParamBounds {
	return getBounds(model)
}
