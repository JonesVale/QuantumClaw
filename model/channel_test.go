package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ── Channel JSON Binding ─────────────────────────────────────

func TestChannelJSONBinding(t *testing.T) {
	jsonStr := `{
		"type": 28,
		"key": "gsk_test_abc123",
		"name": "Test Groq",
		"base_url": "https://api.groq.com/openai",
		"models": "mixtral-8x7b-32768,llama3-70b-8192",
		"group": "default",
		"weight": 2,
		"priority": 1,
		"model_mapping": "",
		"config": ""
	}`

	var ch Channel
	err := json.Unmarshal([]byte(jsonStr), &ch)
	assert.NoError(t, err)
	assert.Equal(t, 28, ch.Type)
	assert.Equal(t, "gsk_test_abc123", ch.Key)
	assert.Equal(t, "Test Groq", ch.Name)
	assert.NotNil(t, ch.BaseURL)
	assert.Equal(t, "https://api.groq.com/openai", *ch.BaseURL)
	assert.Equal(t, "mixtral-8x7b-32768,llama3-70b-8192", ch.Models)
	assert.Equal(t, "default", ch.Group)
	assert.NotNil(t, ch.Weight)
	assert.Equal(t, uint(2), *ch.Weight)
}

func TestChannelJSONBindingMinimal(t *testing.T) {
	// Minimal payload the frontend would send
	jsonStr := `{
		"type": 1,
		"key": "sk-test-key",
		"name": "Minimal Test",
		"models": "gpt-4",
		"group": "default"
	}`

	var ch Channel
	err := json.Unmarshal([]byte(jsonStr), &ch)
	assert.NoError(t, err)
	assert.Equal(t, 1, ch.Type)
	assert.Equal(t, "sk-test-key", ch.Key)
	assert.Equal(t, "Minimal Test", ch.Name)
	assert.Equal(t, "gpt-4", ch.Models)
	assert.Equal(t, "default", ch.Group)
	// These should have zero/nil values when not provided
	assert.Nil(t, ch.BaseURL)
	assert.Nil(t, ch.Weight)
}

func TestChannelJSONBindingExtraFields(t *testing.T) {
	// Extra fields like cache_billing_ratio should be ignored (not cause errors)
	jsonStr := `{
		"type": 35,
		"key": "sk-test",
		"name": "DeepSeek Extra",
		"models": "deepseek-chat",
		"group": "default",
		"cache_billing_ratio": 0.5,
		"thinking_to_content": true,
		"unknown_field": "should be ignored"
	}`

	var ch Channel
	err := json.Unmarshal([]byte(jsonStr), &ch)
	assert.NoError(t, err) // Extra fields should NOT cause unmarshal error
	assert.Equal(t, 35, ch.Type)
	assert.Equal(t, "DeepSeek Extra", ch.Name)
}

// ── Channel ModelMapping ─────────────────────────────────────

func TestChannelModelMappingNil(t *testing.T) {
	ch := Channel{}
	assert.Nil(t, ch.GetModelMapping())

	ch2 := Channel{ModelMapping: nil}
	assert.Nil(t, ch2.GetModelMapping())
}

func TestChannelModelMappingEmpty(t *testing.T) {
	emptyStr := ""
	ch := Channel{ModelMapping: &emptyStr}
	assert.Nil(t, ch.GetModelMapping())
}

func TestChannelModelMappingValid(t *testing.T) {
	mapping := `{"gpt-4":"gpt-4-turbo","claude-3":"claude-3-sonnet"}`
	ch := Channel{ModelMapping: &mapping}
	result := ch.GetModelMapping()
	assert.NotNil(t, result)
	assert.Equal(t, "gpt-4-turbo", result["gpt-4"])
	assert.Equal(t, "claude-3-sonnet", result["claude-3"])
	assert.Equal(t, 2, len(result))
}

// ── Channel Config Loading ───────────────────────────────────

func TestChannelConfigEmpty(t *testing.T) {
	ch := Channel{Config: ""}
	cfg, err := ch.LoadConfig()
	assert.NoError(t, err)
	assert.Equal(t, "", cfg.Region)
	assert.Equal(t, float64(0), cfg.CacheBillingRatio)
}

func TestChannelConfigCacheBilling(t *testing.T) {
	ch := Channel{Config: `{"cache_billing_ratio":0.5,"thinking_to_content":true}`}
	cfg, err := ch.LoadConfig()
	assert.NoError(t, err)
	assert.Equal(t, float64(0.5), cfg.CacheBillingRatio)
	assert.True(t, cfg.ThinkingToContent)
}

func TestChannelConfigFull(t *testing.T) {
	ch := Channel{Config: `{
		"region":"us-east-1",
		"sk":"secret-key",
		"ak":"access-key",
		"api_version":"2024-02-01",
		"cache_billing_ratio":0.3
	}`}
	cfg, err := ch.LoadConfig()
	assert.NoError(t, err)
	assert.Equal(t, "us-east-1", cfg.Region)
	assert.Equal(t, "secret-key", cfg.SK)
	assert.Equal(t, "access-key", cfg.AK)
	assert.Equal(t, "2024-02-01", cfg.APIVersion)
	assert.Equal(t, float64(0.3), cfg.CacheBillingRatio)
}

// ── Channel Key Splitting ────────────────────────────────────

func TestChannelSplitKeys(t *testing.T) {
	// This tests the key splitting logic from AddChannel controller
	keys := splitAndValidate("key1\nkey2\nkey3")
	assert.Equal(t, 3, len(keys))
	assert.Equal(t, "key1", keys[0])
	assert.Equal(t, "key2", keys[1])
	assert.Equal(t, "key3", keys[2])

	// Single key
	keys2 := splitAndValidate("single-key")
	assert.Equal(t, 1, len(keys2))
	assert.Equal(t, "single-key", keys2[0])
}

func TestChannelSplitKeysEmpty(t *testing.T) {
	keys := splitAndValidate("")
	assert.Equal(t, 0, len(keys))
}

func TestChannelSplitKeysWithEmptyLines(t *testing.T) {
	// Keys with empty lines should skip empties
	keys := splitAndValidate("key1\n\nkey2\n\n\nkey3")
	assert.Equal(t, 3, len(keys))
	assert.Equal(t, "key1", keys[0])
	assert.Equal(t, "key2", keys[1])
	assert.Equal(t, "key3", keys[2])
}



// ── Channel Ability Name Parsing ─────────────────────────────

func TestChannelModelsParsing(t *testing.T) {
	ch := Channel{Models: "gpt-4,gpt-3.5-turbo,claude-3"}
	models := parseModels(ch.Models)
	assert.Equal(t, 3, len(models))
	assert.Contains(t, models, "gpt-4")
	assert.Contains(t, models, "gpt-3.5-turbo")
	assert.Contains(t, models, "claude-3")
}

func TestChannelModelsParsingEmpty(t *testing.T) {
	ch := Channel{Models: ""}
	models := parseModels(ch.Models)
	assert.Equal(t, 1, len(models))
	assert.Equal(t, "", models[0]) // Split on empty string returns [""]
}

func TestChannelModelsParsingSpaces(t *testing.T) {
	// Models with spaces should be preserved as-is (backend handles them)
	ch := Channel{Models: "qwen-max, qwen-plus, qwen-turbo"}
	models := parseModels(ch.Models)
	assert.Equal(t, 3, len(models))
	assert.Equal(t, "qwen-max", models[0])
	assert.Equal(t, " qwen-plus", models[1]) // Space preserved - will be stored in DB
	assert.Equal(t, " qwen-turbo", models[2])
}

// ── Helper Functions (refactored from controller) ────────────

func splitAndValidate(key string) []string {
	parts := splitByNewline(key)
	result := make([]string, 0, len(parts))
	for _, k := range parts {
		if k != "" {
			result = append(result, k)
		}
	}
	return result
}

func splitByNewline(s string) []string {
	if s == "" {
		return []string{}
	}
	result := make([]string, 0)
	current := ""
	for _, c := range s {
		if c == '\n' {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	result = append(result, current)
	return result
}

func parseModels(models string) []string {
	if models == "" {
		return []string{""}
	}
	result := make([]string, 0)
	current := ""
	for _, c := range models {
		if c == ',' {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	result = append(result, current)
	return result
}
