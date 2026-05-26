package common

import "testing"

func TestNormalizeModelName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"GPT-4o", "gpt-4o"},
		{"GPT 4o", "gpt-4o"},
		{"Claude 3.5 Sonnet", "claude-3.5-sonnet"},
		{"DeepSeek Chat", "deepseek-chat"},
		{"gpt-4o\t turbo", "gpt-4o--turbo"},
		{"\u00a0test", "-test"},
	}
	for _, tt := range tests {
		got := NormalizeModelName(tt.input)
		if got != tt.expected {
			t.Errorf("NormalizeModelName(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
