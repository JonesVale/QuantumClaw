package operation_setting

import (
	"encoding/json"
	"sync"

	"github.com/quantumclaw/quantumclaw/common/logger"
)

// ModelRateLimitRule defines rate limit for a specific model or model pattern.
type ModelRateLimitRule struct {
	// Model is the exact model name or glob pattern (e.g., "gpt-4*", "claude-*").
	// When empty, applies to all models (catch-all default).
	Model string `json:"model"`

	// MaxRequests is the maximum number of requests allowed within the duration.
	MaxRequests int `json:"max_requests"`

	// Duration in seconds for the rate limit window.
	Duration int64 `json:"duration"`
}

// ModelRateLimitSetting stores per-model rate limit configuration.
type ModelRateLimitSetting struct {
	// Enabled toggles model-level rate limiting globally.
	Enabled bool `json:"enabled"`

	// Rules contains per-model rate limit rules.
	// Rules are evaluated in order; the first matching rule applies.
	Rules []ModelRateLimitRule `json:"rules"`
}

var (
	modelRateLimitSetting   *ModelRateLimitSetting
	modelRateLimitSettingMu sync.RWMutex
)

// GetModelRateLimitSetting returns the current model rate limit setting.
func GetModelRateLimitSetting() *ModelRateLimitSetting {
	modelRateLimitSettingMu.RLock()
	defer modelRateLimitSettingMu.RUnlock()
	if modelRateLimitSetting == nil {
		return &ModelRateLimitSetting{
			Enabled: false,
			Rules:   []ModelRateLimitRule{},
		}
	}
	// Return a copy
	rules := make([]ModelRateLimitRule, len(modelRateLimitSetting.Rules))
	copy(rules, modelRateLimitSetting.Rules)
	return &ModelRateLimitSetting{
		Enabled: modelRateLimitSetting.Enabled,
		Rules:   rules,
	}
}

// SetModelRateLimitSetting updates the model rate limit setting.
func SetModelRateLimitSetting(s *ModelRateLimitSetting) {
	modelRateLimitSettingMu.Lock()
	defer modelRateLimitSettingMu.Unlock()
	if s == nil {
		s = &ModelRateLimitSetting{}
	}
	if s.Rules == nil {
		s.Rules = []ModelRateLimitRule{}
	}
	modelRateLimitSetting = &ModelRateLimitSetting{
		Enabled: s.Enabled,
		Rules:   make([]ModelRateLimitRule, len(s.Rules)),
	}
	copy(modelRateLimitSetting.Rules, s.Rules)
	logger.SysLogf("model rate limit setting updated: enabled=%v, rules=%d", s.Enabled, len(s.Rules))
}

// ParseModelRateLimitSetting parses a JSON string into a ModelRateLimitSetting.
func ParseModelRateLimitSetting(data string) (*ModelRateLimitSetting, error) {
	var s ModelRateLimitSetting
	if data == "" {
		return &ModelRateLimitSetting{
			Enabled: false,
			Rules:   []ModelRateLimitRule{},
		}, nil
	}
	if err := json.Unmarshal([]byte(data), &s); err != nil {
		return nil, err
	}
	if s.Rules == nil {
		s.Rules = []ModelRateLimitRule{}
	}
	return &s, nil
}
