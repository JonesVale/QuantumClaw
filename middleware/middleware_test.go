package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ==================== isModelInList ====================

func TestIsModelInList(t *testing.T) {
	tests := []struct {
		name      string
		modelName string
		models    string
		want      bool
	}{
		{
			name:      "exact match",
			modelName: "gpt-4",
			models:    "gpt-4",
			want:      true,
		},
		{
			name:      "first of multiple models",
			modelName: "gpt-4",
			models:    "gpt-4,gpt-3.5-turbo,claude-3",
			want:      true,
		},
		{
			name:      "middle of multiple models",
			modelName: "gpt-3.5-turbo",
			models:    "gpt-4,gpt-3.5-turbo,claude-3",
			want:      true,
		},
		{
			name:      "last of multiple models",
			modelName: "claude-3",
			models:    "gpt-4,gpt-3.5-turbo,claude-3",
			want:      true,
		},
		{
			name:      "no match",
			modelName: "claude-2",
			models:    "gpt-4,gpt-3.5-turbo",
			want:      false,
		},
		{
			name:      "empty model list",
			modelName: "gpt-4",
			models:    "",
			want:      false,
		},
		{
			name:      "empty model name",
			modelName: "",
			models:    "gpt-4",
			want:      false,
		},
		{
			name:      "both empty (split returns [''], so ''=='' is true)",
			modelName: "",
			models:    "",
			want:      true,
		},
		{
			name:      "partial match not enough",
			modelName: "gpt",
			models:    "gpt-4,gpt-3.5",
			want:      false,
		},
		{
			name:      "single model with trailing comma",
			modelName: "gpt-4",
			models:    "gpt-4,",
			want:      true,
		},
		{
			name:      "case-sensitive match",
			modelName: "GPT-4",
			models:    "gpt-4",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isModelInList(tt.modelName, tt.models)
			assert.Equal(t, tt.want, got,
				"isModelInList(%q, %q) = %v, want %v",
				tt.modelName, tt.models, got, tt.want)
		})
	}
}

// ==================== isModelInList Edge Cases ====================

func TestIsModelInList_SingleElement(t *testing.T) {
	assert.True(t, isModelInList("gpt-4", "gpt-4"))
	assert.False(t, isModelInList("gpt-4o", "gpt-4"))
	assert.False(t, isModelInList("gpt-4", "gpt-4o"))
}

func TestIsModelInList_Whitespace(t *testing.T) {
	// The current implementation uses strings.Split, so "gpt-4" != " gpt-4" (exact match)
	assert.True(t, isModelInList("gpt-4", "gpt-4,gpt-3.5"))
	assert.False(t, isModelInList(" gpt-4", "gpt-4,gpt-3.5"), "whitespace is not trimmed")
}

func TestIsModelInList_ManyModels(t *testing.T) {
	models := "model-1,model-2,model-3,model-4,model-5,model-6,model-7,model-8,model-9,model-10"
	assert.True(t, isModelInList("model-1", models))
	assert.True(t, isModelInList("model-10", models))
	assert.False(t, isModelInList("model-11", models))
}
