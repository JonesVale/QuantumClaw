package ratio

import (
	"testing"
)

// Helper to verify GetModelRatio returns valid values
func TestGetModelRatio_Basic(t *testing.T) {
	// Popular models should have positive ratios
	models := []struct {
		name        string
		channelType int
	}{
		{"gpt-4o", 1},
		{"gpt-4o-mini", 1},
		{"gpt-4-turbo", 1},
		{"claude-3-5-sonnet-20241022", 1},
		{"gpt-4", 1},
	}

	for _, m := range models {
		ratio := GetModelRatio(m.name, m.channelType)
		if ratio <= 0 {
			t.Errorf("GetModelRatio(%q, %d) = %f; expected > 0", m.name, m.channelType, ratio)
		}
	}
}

func TestGetModelRatio_UnknownModel_ReturnsZero(t *testing.T) {
	ratio := GetModelRatio("nonexistent-model-v99", 1)
	if ratio != 0 {
		t.Errorf("Expected 0 for unknown model, got %f", ratio)
	}
}

func TestGetModelRatio_NewModels(t *testing.T) {
	// Verify recently added model ratios from the fix
	models := []string{
		"gpt-4.1",
		"gpt-4.1-mini",
		"gpt-4.1-nano",
		"claude-sonnet-4-20250514",
		"claude-sonnet-4",
		"gemini-2.5-pro-preview-05-07",
		"gemini-2.5-pro",
		"deepseek-v3-0324",
		"deepseek-r1-0528",
	}
	for _, model := range models {
		ratio := GetModelRatio(model, 1)
		if ratio <= 0 {
			t.Errorf("Missing ratio for new model %q", model)
		}
	}
}

func TestGetCompletionRatio_Basic(t *testing.T) {
	models := []struct {
		name        string
		channelType int
	}{
		{"gpt-4o", 1},
		{"gpt-4o-mini", 1},
		{"claude-3-5-sonnet-20241022", 1},
	}

	for _, m := range models {
		ratio := GetCompletionRatio(m.name, m.channelType)
		if ratio <= 0 {
			t.Errorf("GetCompletionRatio(%q, %d) = %f; expected > 0", m.name, m.channelType, ratio)
		}
	}
}

func TestGetCompletionRatio_Fallback(t *testing.T) {
	ratio := GetCompletionRatio("unknown-model", 1)
	if ratio != 1.0 {
		t.Errorf("Expected 1.0 fallback for unknown model, got %f", ratio)
	}
}

func TestModelRatioJSONRoundTrip(t *testing.T) {
	// JSON serialization should round-trip
	jsonStr := ModelRatio2JSONString()
	if jsonStr == "" {
		t.Fatal("ModelRatio2JSONString() returned empty")
	}

	err := UpdateModelRatioByJSONString(jsonStr)
	if err != nil {
		t.Errorf("UpdateModelRatioByJSONString() failed: %v", err)
	}
}

func TestCompletionRatioJSONRoundTrip(t *testing.T) {
	jsonStr := CompletionRatio2JSONString()
	if jsonStr == "" {
		t.Fatal("CompletionRatio2JSONString() returned empty")
	}

	err := UpdateCompletionRatioByJSONString(jsonStr)
	if err != nil {
		t.Errorf("UpdateCompletionRatioByJSONString() failed: %v", err)
	}
}

func TestGetModelRatio_DifferentChannelTypes(t *testing.T) {
	// Same model may have different ratios for different channel types
	ratio1 := GetModelRatio("gpt-4o", 1) // OpenAI
	ratio2 := GetModelRatio("gpt-4o", 14) // Azure
	ratio3 := GetModelRatio("gpt-4o", 32) // Custom

	if ratio1 <= 0 || ratio2 <= 0 || ratio3 <= 0 {
		t.Errorf("All channel types should return positive ratios: type1=%f type14=%f type32=%f", ratio1, ratio2, ratio3)
	}
}
