package common

import (
	"strconv"
	"strings"
)

// ParseModelSuffix parses known reasoning/thinking suffixes from model names.
//
// Supported suffix patterns:
//   - o3-mini-high         → base="o3-mini",          ReasoningEffort="high"
//   - o3-mini-medium       → base="o3-mini",          ReasoningEffort="medium"
//   - o3-mini-low          → base="o3-mini",          ReasoningEffort="low"
//   - claude-3-7-sonnet-20250219-thinking → base="claude-3-7-sonnet-20250219", Thinking=true
//   - gemini-2.5-flash-thinking        → base="gemini-2.5-flash", Thinking=true
//   - gemini-2.5-flash-nothinking      → base="gemini-2.5-flash", Thinking=false
//   - gemini-2.5-pro-thinking-128      → base="gemini-2.5-pro",   Thinking=true, ThinkingBudget=128
func ParseModelSuffix(modelName string) (base string, reasoningEffort string, thinking bool, thinkingBudget int) {
	if modelName == "" {
		return modelName, "", false, 0
	}

	// 1. Check for o3-mini / o1 / o1-mini reasoning effort suffixes
	//    Pattern: <base>-high | <base>-medium | <base>-low
	for _, prefix := range []string{"o3-mini", "o1-mini", "o1-preview", "o1"} {
		if strings.HasPrefix(modelName, prefix) {
			rest := modelName[len(prefix):]
			if rest == "-high" || rest == "-medium" || rest == "-low" {
				return prefix, rest[1:], false, 0 // trim leading '-'
			}
			if rest == "" {
				return modelName, "", false, 0
			}
		}
	}

	// 2. Check for Claude thinking suffix: -thinking
	if strings.HasSuffix(modelName, "-thinking") {
		base = strings.TrimSuffix(modelName, "-thinking")
		return base, "", true, 0
	}

	// 3. Check for Gemini thinking/nothinking suffix
	//    Patterns:
	//      gemini-2.5-flash-thinking       → thinking=true
	//      gemini-2.5-flash-nothinking     → thinking=false
	//      gemini-2.5-flash-thinking-128   → thinking=true, budget=128
	//    The thinking/nothinking suffix is always the last meaningful token
	if strings.HasPrefix(modelName, "gemini-") || strings.Contains(modelName, "gemini") {
		// Try -thinking-NNN  (thinking with budget)
		if idx := strings.LastIndex(modelName, "-thinking-"); idx > 0 {
			budgetStr := modelName[idx+len("-thinking-"):]
			if budget, err := strconv.Atoi(budgetStr); err == nil && budget > 0 {
				base = modelName[:idx]
				return base, "", true, budget
			}
		}
		// Try -thinking (no budget)
		if strings.HasSuffix(modelName, "-thinking") {
			base = strings.TrimSuffix(modelName, "-thinking")
			return base, "", true, 0
		}
		// Try -nothinking
		if strings.HasSuffix(modelName, "-nothinking") {
			base = strings.TrimSuffix(modelName, "-nothinking")
			return base, "", false, 0
		}
	}

	// No recognized suffix
	return modelName, "", false, 0
}

// StripModelSuffix removes recognized reasoning/thinking suffixes from a model name,
// returning the base model name. It is a convenience wrapper around ParseModelSuffix.
func StripModelSuffix(modelName string) string {
	base, _, _, _ := ParseModelSuffix(modelName)
	return base
}
