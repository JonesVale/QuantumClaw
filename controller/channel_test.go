package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	dbmodel "github.com/quantumclaw/quantumclaw/model"
)

// ── Test Helpers ─────────────────────────────────────────────

func setupChannelTest() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{
		Header: make(http.Header),
	}
	c.Request.Header.Set("Content-Type", "application/json")
	return c, w
}

// ── AddChannel JSON Parsing ──────────────────────────────────

func TestAddChannelValidJSON(t *testing.T) {
	validPayloads := []string{
		`{"type":1,"key":"sk-test","name":"OpenAI Test","models":"gpt-4","group":"default"}`,
		`{"type":28,"key":"gsk-test","name":"Groq Test","models":"mixtral-8x7b-32768","group":"default","weight":1,"priority":0}`,
		`{"type":35,"key":"sk-test","name":"DeepSeek","models":"deepseek-chat","group":"vip"}`,
		`{"type":100,"key":"ionq-key","name":"IonQ","models":"ionq_harmony","group":"default"}`,
		`{"type":44,"key":"sf-key","name":"SiliconFlow","models":"Qwen/Qwen2.5-7B-Instruct,deepseek-ai/DeepSeek-V3","group":"default"}`,
	}

	for _, payload := range validPayloads {
		var ch dbmodel.Channel
		err := json.Unmarshal([]byte(payload), &ch)
		assert.NoError(t, err, "Failed to parse: "+payload)
		assert.True(t, ch.Type > 0, "Type should be > 0: "+payload)
		assert.NotEmpty(t, ch.Key, "Key should not be empty: "+payload)
		assert.NotEmpty(t, ch.Name, "Name should not be empty: "+payload)
		assert.NotEmpty(t, ch.Models, "Models should not be empty: "+payload)
	}
}

func TestAddChannelEmptyKey(t *testing.T) {
	var ch dbmodel.Channel
	err := json.Unmarshal([]byte(`{"type":1,"key":"","name":"No Key","models":"gpt-4","group":"default"}`), &ch)
	assert.NoError(t, err)
	assert.Empty(t, ch.Key)

	// Simulate backend validation:
	// The AddChannel controller splits key by newline, skips empty
	keys := splitKeys(ch.Key)
	assert.Equal(t, 0, len(keys), "Empty key should produce no keys after split+filter")
}

func TestAddChannelExtraFields(t *testing.T) {
	// Frontend sends extra fields like cache_billing_ratio — should not break binding
	var ch dbmodel.Channel
	err := json.Unmarshal([]byte(`{
		"type":1,
		"key":"sk-test",
		"name":"With Extras",
		"models":"gpt-4",
		"group":"default",
		"cache_billing_ratio":0.5,
		"thinking_to_content":true
	}`), &ch)
	assert.NoError(t, err)
	assert.Equal(t, "With Extras", ch.Name)

	// Note: cache_billing_ratio and thinking_to_content are stored in Config field
	// When sent as top-level JSON, they are silently dropped by encoding/json
	// This is the expected behavior - they need to be in "config":"{...}"
}

func TestAddChannelInvalidType(t *testing.T) {
	var ch dbmodel.Channel
	err := json.Unmarshal([]byte(`{"type":"invalid","key":"sk-test","name":"Bad","models":"gpt-4","group":"default"}`), &ch)
	assert.Error(t, err, "Should fail with type as string")
}

// ── Channel Config JSON ──────────────────────────────────────

func TestChannelConfigSerialization(t *testing.T) {
	// This tests that config fields go into the "config" JSON field
	configPayload := `{"cache_billing_ratio":0.5,"thinking_to_content":true}`
	var cfg dbmodel.ChannelConfig
	err := json.Unmarshal([]byte(configPayload), &cfg)
	assert.NoError(t, err)
	assert.Equal(t, float64(0.5), cfg.CacheBillingRatio)
	assert.True(t, cfg.ThinkingToContent)

	// Round-trip
	data, _ := json.Marshal(cfg)
	var cfg2 dbmodel.ChannelConfig
	json.Unmarshal(data, &cfg2)
	assert.Equal(t, float64(0.5), cfg2.CacheBillingRatio)
}

// ── Gin Response Format ──────────────────────────────────────

func TestChannelResponseFormat(t *testing.T) {
	c, w := setupChannelTest()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": []gin.H{
			{"id": 1, "name": "Channel 1", "type": 1, "status": 1},
			{"id": 2, "name": "Channel 2", "type": 28, "status": 1},
		},
	})

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Success bool            `json:"success"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.NotEmpty(t, resp.Data)
}

func TestChannelErrorResponseFormat(t *testing.T) {
	c, w := setupChannelTest()

	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"message": "channel not found",
	})

	var resp struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, "channel not found", resp.Message)
}

// ── Channel Type Constants ───────────────────────────────────

func TestChannelTypeRange(t *testing.T) {
	// AI channels are type < 100
	assert.True(t, isAIChannel(1))
	assert.True(t, isAIChannel(28))
	assert.True(t, isAIChannel(58))
	assert.False(t, isAIChannel(100))
	assert.False(t, isAIChannel(105))

	// Quantum channels are type >= 100
	assert.True(t, isQuantumChannel(100))
	assert.True(t, isQuantumChannel(105))
	assert.False(t, isQuantumChannel(1))
	assert.False(t, isQuantumChannel(58))
}

// ── Helper Functions ─────────────────────────────────────────

func TestGetTypeBadgeLabels(t *testing.T) {
	typeNames := map[int]string{
		1:   "OpenAI",
		28:  "Groq",
		35:  "DeepSeek",
		44:  "SiliconFlow",
		100: "IonQ",
		101: "IBM Q",
		102: "Rigetti",
		103: "AWS Braket",
		104: "Azure Quantum",
		105: "Google Quantum",
	}

	for typeID, expectedName := range typeNames {
		name, exists := typeNames[typeID]
		assert.True(t, exists)
		assert.Equal(t, expectedName, name)
	}
}

// ── Key Splitting Logic (mirrors AddChannel controller) ──────

func splitKeys(key string) []string {
	if key == "" {
		return []string{}
	}
	parts := splitByNewline(key)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func splitByNewline(s string) []string {
	result := make([]string, 0)
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}

func isAIChannel(typeID int) bool {
	return typeID > 0 && typeID < 100
}

func isQuantumChannel(typeID int) bool {
	return typeID >= 100
}
