// Package web_shared provides a generic, schema-driven adapter engine for web-based LLM providers.
// It translates OpenAI-compatible requests into provider-specific API calls using
// configurable templates, credential injection, and streaming response parsing.
package web_shared

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/quantumclaw/quantumclaw/model"
	relaymodel "github.com/quantumclaw/quantumclaw/relay/model"
)

// ── Placeholder Constants ──

const (
	PH_MODEL        = "{{model}}"
	PH_MESSAGES     = "{{messages}}"
	PH_ROLE         = "{{role}}"
	PH_CONTENT      = "{{content}}"
	PH_STREAM       = "{{stream}}"
	PH_MAX_TOKENS   = "{{max_tokens}}"
	PH_TEMPERATURE  = "{{temperature}}"
	PH_UUID         = "{{uuid}}"
	PH_TIMESTAMP    = "{{timestamp}}"
	PH_ORG_ID       = "{{org_id}}"
)

// ── Render Input ──

type RenderInput struct {
	Schema      *model.Sub2APISchema
	APIModel    string // original model name from the API request (e.g. "gpt-4o")
	Messages    []relaymodel.Message
	Stream      bool
	MaxTokens   int
	Temperature float64
	CookieToken string // decrypted credential value (cookie/bearer/apikey)
	OrgID       string // optional org id for providers that need it
}

// ── Render Output ──

type RenderOutput struct {
	EndpointURL string
	Headers     map[string]string
	Body        string
	Method      string // default "POST"
}

// RenderRequest renders a full HTTP request from a schema template and input.
func RenderRequest(input *RenderInput) (*RenderOutput, error) {
	s := input.Schema

	// 1. Map the model name
	mappedModel, ok := s.MapModel(input.APIModel)
	if !ok {
		mappedModel = input.APIModel
	}

	// 2. Build the messages block
	messagesBlock := buildMessagesJSON(input.Messages)

	// 3. Replace placeholders in request_template
	body := s.RequestTemplate
	body = strings.ReplaceAll(body, PH_MODEL, mappedModel)
	body = strings.ReplaceAll(body, PH_MESSAGES, messagesBlock)
	body = strings.ReplaceAll(body, PH_STREAM, fmt.Sprintf("%v", input.Stream))
	body = strings.ReplaceAll(body, PH_MAX_TOKENS, fmt.Sprintf("%d", input.MaxTokens))
	body = strings.ReplaceAll(body, PH_TEMPERATURE, fmt.Sprintf("%.2f", input.Temperature))
	body = strings.ReplaceAll(body, PH_UUID, uuid.New().String())
	body = strings.ReplaceAll(body, PH_TIMESTAMP, fmt.Sprintf("%d", time.Now().UnixMilli()))
	body = strings.ReplaceAll(body, PH_ORG_ID, input.OrgID)

	// If the template has single-message format {{role}}/{{content}},
	// use the LAST user message
	if strings.Contains(body, PH_ROLE) || strings.Contains(body, PH_CONTENT) {
		lastRole, lastContent := extractLastMessage(input.Messages)
		body = strings.ReplaceAll(body, PH_ROLE, lastRole)
		body = strings.ReplaceAll(body, PH_CONTENT, lastContent)
	}

	// 4. Parse headers template
	headers := make(map[string]string)
	if s.HeadersTemplate != "" && s.HeadersTemplate != "{}" {
		var hdrs map[string]interface{}
		if err := json.Unmarshal([]byte(s.HeadersTemplate), &hdrs); err == nil {
			for k, v := range hdrs {
				if vStr, ok := v.(string); ok {
					// Replace placeholders in header values too
					vStr = strings.ReplaceAll(vStr, PH_UUID, uuid.New().String())
					headers[k] = vStr
				}
			}
		}
	}

	// 5. Inject auth
	injectAuth(headers, s, input.CookieToken)

	// 6. Resolve endpoint URL
	endpoint := s.EndpointURL
	endpoint = strings.ReplaceAll(endpoint, PH_ORG_ID, input.OrgID)
	endpoint = strings.ReplaceAll(endpoint, PH_MODEL, mappedModel)
	endpoint = strings.ReplaceAll(endpoint, PH_UUID, uuid.New().String())

	return &RenderOutput{
		EndpointURL: endpoint,
		Headers:     headers,
		Body:        body,
		Method:      "POST",
	}, nil
}

// injectAuth adds credential headers based on auth type.
func injectAuth(headers map[string]string, s *model.Sub2APISchema, token string) {
	if token == "" {
		return
	}
	switch s.AuthType {
	case "cookie":
		headers["Cookie"] = fmt.Sprintf("%s=%s", s.AuthKeyName, token)
	case "bearer":
		headers["Authorization"] = "Bearer " + token
	case "apikey":
		if s.AuthKeyName != "" {
			headers[s.AuthKeyName] = token
		} else {
			headers["X-API-Key"] = token
		}
	}
}

// buildMessagesJSON converts the relay message array to a JSON string.
// Some providers want the full array, others want single messages.
func buildMessagesJSON(msgs []relaymodel.Message) string {
	if len(msgs) == 0 {
		return "[]"
	}
	type msg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	simplified := make([]msg, 0, len(msgs))
	for _, m := range msgs {
		content := m.StringContent()
		role := m.Role
		if role == "" {
			role = "user"
		}
		simplified = append(simplified, msg{Role: role, Content: content})
	}
	data, err := json.Marshal(simplified)
	if err != nil {
		return "[]"
	}
	return string(data)
}

// extractLastMessage returns the role and content of the last message.
func extractLastMessage(msgs []relaymodel.Message) (string, string) {
	if len(msgs) == 0 {
		return "user", ""
	}
	last := msgs[len(msgs)-1]
	content := last.StringContent()
	role := last.Role
	if role == "" {
		role = "user"
	}
	return role, content
}
