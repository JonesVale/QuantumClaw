package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/logger"
)

//  Prompt Optimizer Middleware 
// Intercepts /v1/chat/completions requests and automatically enhances
// user prompts to improve AI response quality.
//
// Levels:
//   - off:      no optimization
//   - basic:    add quality-enhancing system prompt
//   - advanced: rewrite user messages for clarity (rule-based)

type PromptOptimizerConfig struct {
	Enabled bool   `json:"enabled"`
	Level   string `json:"level"` // "off", "basic", "advanced"
}

var PromptOptimizer = PromptOptimizerConfig{
	Enabled: false,
	Level:   "basic",
}

// qualitySystemPrompts contains role-specific enhancement prompts.
var qualitySystemPrompts = map[string]string{
	"default": `You are a helpful, thorough, and accurate AI assistant. When responding:
- Provide well-structured answers with clear sections where appropriate
- Use bullet points for lists
- Be precise about facts; if uncertain, acknowledge the uncertainty
- Use code blocks with language tags for code
- Think step by step for complex questions`,
}

// advancedRewriteRules enhance user messages for clarity.
var advancedRewriteRules = []struct {
	Match   string
	Replace string
}{
	{"write code", "write working production-ready code with error handling"},
	{"translate", "provide an accurate translation with any cultural nuances explained"},
	{"summarize", "provide a comprehensive summary covering all key points"},
}

// PromptOptimizerMiddleware enhances prompts for better AI responses.
func PromptOptimizerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !PromptOptimizer.Enabled || PromptOptimizer.Level == "off" {
			c.Next()
			return
		}

		path := c.Request.URL.Path
		if !strings.HasSuffix(path, "/chat/completions") {
			c.Next()
			return
		}

		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.Next()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		var req struct {
			Model    string          `json:"model"`
			Messages json.RawMessage `json:"messages"`
		}
		if err := json.Unmarshal(bodyBytes, &req); err != nil || req.Messages == nil {
			c.Next()
			return
		}

		var messages []map[string]interface{}
		if err := json.Unmarshal(req.Messages, &messages); err != nil || len(messages) == 0 {
			c.Next()
			return
		}

		modified := false

		if PromptOptimizer.Level == "basic" || PromptOptimizer.Level == "advanced" {
			// Add/enhance system prompt
			hasSystem := false
			for i, msg := range messages {
				role, _ := msg["role"].(string)
				if role == "system" {
					hasSystem = true
					content, _ := msg["content"].(string)
					if prompt, ok := qualitySystemPrompts["default"]; ok {
						msg["content"] = content + "\n\n" + prompt
						messages[i] = msg
						modified = true
					}
					break
				}
			}
			if !hasSystem {
				if prompt, ok := qualitySystemPrompts["default"]; ok {
					messages = append([]map[string]interface{}{
						{"role": "system", "content": prompt},
					}, messages...)
					modified = true
				}
			}
		}

		if PromptOptimizer.Level == "advanced" {
			// Rewrite user messages
			for i, msg := range messages {
				role, _ := msg["role"].(string)
				if role != "user" {
					continue
				}
				content, _ := msg["content"].(string)
				newContent := content
				for _, rule := range advancedRewriteRules {
					if strings.Contains(strings.ToLower(content), rule.Match) {
						newContent = strings.Replace(
							strings.ToLower(content),
							rule.Match, rule.Replace, 1,
						)
						modified = true
					}
				}
				if newContent != content {
					msg["content"] = newContent
					messages[i] = msg
					logger.Info(context.Background(),
						fmt.Sprintf("PromptOptimizer: rewrote user message [%d]: %q �?%q", i, content, newContent))
				}
			}
		}

		if modified {
			var rawMap map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &rawMap); err == nil {
				rawMap["messages"] = messages
				newBody, err := json.Marshal(rawMap)
				if err == nil {
					c.Request.Body = io.NopCloser(bytes.NewBuffer(newBody))
					c.Request.ContentLength = int64(len(newBody))
					logger.Debug(context.Background(),
						fmt.Sprintf("PromptOptimizer: optimized prompt for model %s", req.Model))
				}
			}
		}

		c.Next()
	}
}

// SetPromptOptimizerConfig updates the optimizer config at runtime.
func SetPromptOptimizerConfig(cfg PromptOptimizerConfig) {
	PromptOptimizer = cfg
}
