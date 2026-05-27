package model

import (
	"encoding/json"
	"time"
)

// ── Schema Model ──

// Sub2APISchema defines the request/response contract for a web LLM provider.
// Multiple versions per provider create a fallback chain.
type Sub2APISchema struct {
	Id        int    `json:"id" gorm:"primaryKey"`
	Provider  string `json:"provider" gorm:"type:varchar(64);not null;index:idx_provider_version"` // "chatgpt","claude","gemini",...
	Version   int    `json:"version" gorm:"type:int;not null;default:1;index:idx_provider_version"`
	Label     string `json:"label" gorm:"type:varchar(255)"` // human-readable, e.g. "ChatGPT v4 (2026-03)"

	// ── Connection ──
	EndpointURL string `json:"endpoint_url" gorm:"type:text;not null"`
	AuthType    string `json:"auth_type" gorm:"type:varchar(32);default:'cookie'"` // cookie / bearer / apikey / none
	AuthKeyName string `json:"auth_key_name" gorm:"type:varchar(128);default:''"`  // cookie name or header key

	// ── Templates ──
	// Request body template with {{placeholders}}. This is the core abstraction.
	RequestTemplate string `json:"request_template" gorm:"type:mediumtext;not null"`
	// Default headers (JSON object)
	HeadersTemplate string `json:"headers_template" gorm:"type:text;not null;default:'{}'"`
	// JSONPath or dot-notation to extract response text
	ResponsePath string `json:"response_path" gorm:"type:varchar(255);not null;default:'message.content'"`
	// Streaming mode
	StreamMode string `json:"stream_mode" gorm:"type:varchar(32);default:'sse'"` // sse / websocket / poll / none

	// ── Model Mapping ──
	// JSON map: {"gpt-4o":"gpt-4o","gpt-4":"text-davinci-002-render-sha"}
	// Keys are our API model names, values are the provider's internal model identifiers.
	ModelMapping string `json:"model_mapping" gorm:"type:mediumtext;default:'{}'"`

	// ── Status ──
	Status    int `json:"status" gorm:"default:1"`         // 1=active, 2=draft, 3=deprecated, 4=broken
	Priority  int `json:"priority" gorm:"default:0;index"` // higher = preferred when multiple versions active
	IsBuiltin bool `json:"is_builtin" gorm:"default:false"` // true = shipped with platform, seed data

	// ── Health ──
	LastHealthAt int64  `json:"last_health_at" gorm:"bigint;default:0"`
	LastError    string `json:"last_error" gorm:"type:text"`

	CreatedTime int64 `json:"created_time" gorm:"bigint"`
	UpdatedTime int64 `json:"updated_time" gorm:"bigint"`
}

func (Sub2APISchema) TableName() string { return "sub2api_schemas" }

// ── Helper types ──

// Sub2APITemplatePlaceholders documents all available {{placeholders}} for request templates.
var Sub2APITemplatePlaceholders = map[string]string{
	"{{model}}":       "Mapped model identifier (e.g., 'gpt-4o' → 'gpt-4o' after ModelMapping)",
	"{{messages}}":    "JSON array of messages formatted per provider spec",
	"{{stream}}":      "true/false for streaming mode",
	"{{max_tokens}}":  "Maximum tokens to generate",
	"{{temperature}}": "Temperature value (0.0-2.0)",
	"{{uuid}}":        "Generated UUID for message/conversation IDs",
	"{{timestamp}}":   "Current timestamp in provider-expected format",
	"{{role}}":        "Message role (user/assistant/system)",
	"{{content}}":     "Message content text",
}

// ParseModelMapping parses the ModelMapping JSON into a usable map.
func (s *Sub2APISchema) ParseModelMapping() (map[string]string, error) {
	if s.ModelMapping == "" || s.ModelMapping == "{}" {
		return map[string]string{}, nil
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(s.ModelMapping), &m); err != nil {
		return nil, err
	}
	return m, nil
}

// MapModel converts our API model name to the provider's internal name.
func (s *Sub2APISchema) MapModel(apiModel string) (string, bool) {
	m, err := s.ParseModelMapping()
	if err != nil || len(m) == 0 {
		return apiModel, true // passthrough if no mapping
	}
	if internal, ok := m[apiModel]; ok {
		return internal, true
	}
	// Fallback: try wildcard
	if internal, ok := m["*"]; ok {
		return internal, true
	}
	return apiModel, false // no match, but still use the original
}

// ── Built-in Templates ──

// SeedSub2APISchemas returns the built-in schema templates shipped with the platform.
func SeedSub2APISchemas() []Sub2APISchema {
	now := time.Now().UnixMilli()
	return []Sub2APISchema{
		{
			Provider:  "chatgpt",
			Version:   4,
			Label:     "ChatGPT (backend-api/v1)",
			EndpointURL: "https://chatgpt.com/backend-api/conversation",
			AuthType:    "cookie",
			AuthKeyName: "__Secure-next-auth.session-token",
			RequestTemplate: `{
			  "action": "next",
			  "messages": [
			    {
			      "id": "{{uuid}}",
			      "author": { "role": "{{role}}" },
			      "content": { "content_type": "text", "parts": ["{{content}}"] }
			    }
			  ],
			  "model": "{{model}}",
			  "parent_message_id": "{{uuid}}",
			  "conversation_id": null,
			  "stream": {{stream}}
			}`,
			HeadersTemplate: `{
			  "Content-Type": "application/json",
			  "Accept": "text/event-stream",
			  "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			  "Oai-Device-Id": "{{uuid}}"
			}`,
			ResponsePath: "message.content.parts.0",
			StreamMode:   "sse",
			ModelMapping: `{
			  "gpt-4o": "gpt-4o",
			  "gpt-4": "text-davinci-002-render-sha",
			  "gpt-4-turbo": "gpt-4-turbo-2024-04-09",
			  "gpt-3.5-turbo": "text-davinci-002-render-sha",
			  "dall-e-3": "dall-e-3",
			  "*": "gpt-4o"
			}`,
			Status:    1,
			Priority:  100,
			IsBuiltin: true,
			CreatedTime: now,
			UpdatedTime: now,
		},
		{
			Provider:  "claude",
			Version:   3,
			Label:     "Claude (claude.ai/api)",
			EndpointURL: "https://claude.ai/api/organizations/{{org_id}}/chat_conversations",
			AuthType:    "cookie",
			AuthKeyName: "sessionKey",
			RequestTemplate: `{
			  "chatbot_id": "{{model}}",
			  "messages": [
			    {
			      "role": "{{role}}",
			      "content": [
			        {
			          "type": "text",
			          "text": "{{content}}"
			        }
			      ]
			    }
			  ],
			  "stream": {{stream}},
			  "rendering_mode": "messages"
			}`,
			HeadersTemplate: `{
			  "Content-Type": "application/json",
			  "Accept": "text/event-stream",
			  "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
			}`,
			ResponsePath: "content.0.text",
			StreamMode:   "sse",
			ModelMapping: `{
			  "claude-3-opus": "claude-3-opus-20240229",
			  "claude-3-sonnet": "claude-3-sonnet-20240229",
			  "claude-3-haiku": "claude-3-haiku-20240307",
			  "*": "claude-3-sonnet"
			}`,
			Status:    1,
			Priority:  100,
			IsBuiltin: true,
			CreatedTime: now,
			UpdatedTime: now,
		},
		{
			Provider:  "gemini",
			Version:   1,
			Label:     "Gemini (gemini.google.com)",
			EndpointURL: "https://gemini.google.com/_/BardChatUi/data/assistant.lamda.BardFrontendService/StreamGenerate",
			AuthType:    "cookie",
			AuthKeyName: "__Secure-1PSID",
			RequestTemplate: `{
			  "question": "{{content}}",
			  "model": "{{model}}",
			  "conversation_id": "{{uuid}}",
			  "stream": {{stream}}
			}`,
			HeadersTemplate: `{
			  "Content-Type": "application/json",
			  "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
			}`,
			ResponsePath: "content",
			StreamMode:   "sse",
			ModelMapping: `{
			  "gemini-2.0-flash": "gemini-2.0-flash",
			  "gemini-2.0-pro": "gemini-2.0-pro",
			  "*": "gemini-2.0-flash"
			}`,
			Status:    2,
			Priority:  50,
			IsBuiltin: true,
			CreatedTime: now,
			UpdatedTime: now,
		},
		{
			Provider:  "deepseek",
			Version:   1,
			Label:     "DeepSeek Chat (免费)",
			EndpointURL: "https://chat.deepseek.com/api/v0/chat/completions",
			AuthType:    "bearer",
			AuthKeyName: "Authorization",
			RequestTemplate: `{
			  "model": "{{model}}",
			  "messages": [{"role": "{{role}}", "content": "{{content}}"}],
			  "stream": {{stream}},
			  "max_tokens": {{max_tokens}}
			}`,
			HeadersTemplate: `{
			  "Content-Type": "application/json",
			  "User-Agent": "Mozilla/5.0",
			  "Accept": "text/event-stream"
			}`,
			ResponsePath: "choices.0.delta.content",
			StreamMode:   "sse",
			ModelMapping: `{
			  "deepseek-chat": "deepseek-chat",
			  "deepseek-reasoner": "deepseek-reasoner",
			  "*": "deepseek-chat"
			}`,
			Status:    2,
			Priority:  50,
			IsBuiltin: true,
			CreatedTime: now,
			UpdatedTime: now,
		},
		{
			Provider:  "grok",
			Version:   1,
			Label:     "Grok (X Premium)",
			EndpointURL: "https://api.x.com/2/grok/conversation",
			AuthType:    "cookie",
			AuthKeyName: "auth_token",
			RequestTemplate: `{
			  "message": {"role": "{{role}}", "content": "{{content}}"},
			  "model": "{{model}}",
			  "stream": {{stream}}
			}`,
			HeadersTemplate: `{
			  "Content-Type": "application/json",
			  "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			  "X-Csrf-Token": "",
			  "Authorization": "Bearer {{token}}"
			}`,
			ResponsePath: "result.response.text",
			StreamMode:   "sse",
			ModelMapping: `{
			  "grok-2": "grok-2",
			  "grok-3": "grok-3",
			  "*": "grok-2"
			}`,
			Status:    2,
			Priority:  50,
			IsBuiltin: true,
			CreatedTime: now,
			UpdatedTime: now,
		},
	}
}

// ── CRUD ──

func CreateSub2APISchema(s *Sub2APISchema) error {
	s.CreatedTime = time.Now().UnixMilli()
	s.UpdatedTime = s.CreatedTime
	return DB.Create(s).Error
}

func GetSub2APISchema(id int) (*Sub2APISchema, error) {
	var s Sub2APISchema
	err := DB.First(&s, id).Error
	return &s, err
}

func ListSub2APISchemas() ([]Sub2APISchema, error) {
	var schemas []Sub2APISchema
	err := DB.Order("provider asc, priority desc, version desc").Find(&schemas).Error
	return schemas, err
}

func ListActiveSchemasByProvider(provider string) ([]Sub2APISchema, error) {
	var schemas []Sub2APISchema
	err := DB.Where("provider = ? AND status = 1", provider).
		Order("priority desc, version desc").Find(&schemas).Error
	return schemas, err
}

func GetActiveSchema(provider string) (*Sub2APISchema, error) {
	var s Sub2APISchema
	err := DB.Where("provider = ? AND status = 1", provider).
		Order("priority desc, version desc").First(&s).Error
	return &s, err
}

func UpdateSub2APISchema(s *Sub2APISchema) error {
	s.UpdatedTime = time.Now().UnixMilli()
	return DB.Select("*").Omit("created_time", "is_builtin").Updates(s).Error
}

func DeleteSub2APISchema(id int) error {
	return DB.Delete(&Sub2APISchema{}, id).Error
}

// CountActiveSchemas returns count of active schemas per provider.
func CountActiveSchemas() ([]struct {
	Provider string `json:"provider"`
	Count    int    `json:"count"`
}, error) {
	var results []struct {
		Provider string `json:"provider"`
		Count    int    `json:"count"`
	}
	err := DB.Model(&Sub2APISchema{}).
		Select("provider, COUNT(*) as count").
		Where("status = 1").
		Group("provider").
		Scan(&results).Error
	return results, err
}

// AttemptSchemaFallback tries the next priority schema when the current one fails.
func AttemptSchemaFallback(provider string, currentVersion int) (*Sub2APISchema, error) {
	var s Sub2APISchema
	err := DB.Where("provider = ? AND status = 1 AND version != ?", provider, currentVersion).
		Order("priority desc, version desc").First(&s).Error
	return &s, err
}
