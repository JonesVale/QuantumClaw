package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/relay/channeltype"
)

// ── Free chat: reads from channels table ──
// No more env vars. Admins/resellers configure channels at /channels.
// Free chat auto-discovers eligible channels (type in ETypeSet, enabled, key not empty).

// freeChatEligibleTypes — which channel types are eligible for free chat
var freeChatEligibleTypes = map[int]bool{
	channeltype.Groq:        true,
	channeltype.DeepSeek:    true,
	channeltype.Gemini:      true,
	channeltype.SiliconFlow: true,
	channeltype.Mistral:     true,
}

type freeChatProvider struct {
	Type      int                  `json:"type"`
	Name      string               `json:"name"`
	ChannelID int                  `json:"channel_id"`
	Endpoint  string               `json:"endpoint"`
	APIKey    string               `json:"-"`
	Models    []map[string]string  `json:"models"`
}

// resolveFreeChatProviders reads from channels table, filters eligible ones
func resolveFreeChatProviders() []freeChatProvider {
	var channels []model.Channel
	model.DB.Where("status = ? AND key != '' AND key NOT LIKE ? AND type IN (?)",
		model.ChannelStatusEnabled, "PUT_YOUR%", []int{channeltype.Groq, channeltype.DeepSeek, channeltype.Gemini, channeltype.SiliconFlow, channeltype.Mistral}).Find(&channels)

	typeNames := channeltype.ChannelTypeNames
	providerMap := make(map[int]*freeChatProvider)

	for _, ch := range channels {
		if ch.Key == "" {
			continue
		}

		pName := ch.Name
		if name, ok := typeNames[ch.Type]; ok {
			pName = name
		}

		if _, exists := providerMap[ch.Type]; !exists {
			endpoint := ""
			if ch.BaseURL != nil {
				endpoint = *ch.BaseURL
			}
			if endpoint == "" {
				endpoint = getDefaultEndpoint(ch.Type)
			}

			providerMap[ch.Type] = &freeChatProvider{
				Type:      ch.Type,
				Name:      pName,
				ChannelID: ch.Id,
				Endpoint:  endpoint,
				APIKey:    ch.Key,
				Models:    parseModels(ch.Models),
			}
		}
	}

	result := make([]freeChatProvider, 0, len(providerMap))
	for _, p := range providerMap {
		result = append(result, *p)
	}
	return result
}

func getDefaultEndpoint(chType int) string {
	switch chType {
	case 1:
		return "https://api.openai.com/v1/chat/completions"
	case 14:
		return "https://api.anthropic.com/v1/messages"
	case 24:
		return "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions"
	case 28:
		return "https://api.mistral.ai/v1/chat/completions"
	case 29:
		return "https://api.groq.com/openai/v1/chat/completions"
	case 36:
		return "https://api.deepseek.com/v1/chat/completions"
	case 44:
		return "https://api.siliconflow.cn/v1/chat/completions"
	default:
		return "https://api.openai.com/v1/chat/completions"
	}
}

func parseModels(modelsStr string) []map[string]string {
	if modelsStr == "" {
		return nil
	}
	var result []map[string]string
	for _, m := range strings.Split(modelsStr, ",") {
		m = strings.TrimSpace(m)
		if m != "" {
			result = append(result, map[string]string{"id": m, "name": m})
		}
	}
	return result
}

// GetFreeChatProviders returns available free chat providers
func GetFreeChatProviders(c *gin.Context) {
	providers := resolveFreeChatProviders()
	type providerResp struct {
		Type       int                  `json:"type"`
		Name       string               `json:"name"`
		ChannelID  int                  `json:"channel_id"`
		Configured bool                 `json:"configured"`
		Models     []map[string]string  `json:"models"`
	}
	resp := make([]providerResp, 0, len(providers))
	for _, p := range providers {
		resp = append(resp, providerResp{
			Type:       p.Type,
			Name:       p.Name,
			ChannelID:  p.ChannelID,
			Configured: p.APIKey != "",
			Models:     p.Models,
		})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

// Chat handles free chat requests
func Chat(c *gin.Context) {
	provider := c.Query("provider")
	modelId := c.Query("model")
	content := c.Query("content")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider is required"})
		return
	}
	if modelId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model is required"})
		return
	}
	if content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content is required"})
		return
	}

	providers := resolveFreeChatProviders()
	var fp *freeChatProvider
	for _, p := range providers {
		if strings.EqualFold(p.Name, provider) {
			fp = &p
			break
		}
	}
	if fp == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown provider: " + provider})
		return
	}

	if fp.APIKey == "" {
		logger.SysError(fmt.Sprintf("free chat: %s has no API key configured", provider))
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("Free chat: %s not configured. Set up a channel for this provider in Channels page.", provider),
		})
		return
	}

	// Validate model
	validModel := false
	for _, m := range fp.Models {
		if m["id"] == modelId {
			validModel = true
			break
		}
	}
	if !validModel {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model '" + modelId + "' is not available for provider '" + provider + "'"})
		return
	}

	// Build OpenAI-compatible request
	payload := map[string]interface{}{
		"model": modelId,
		"messages": []map[string]string{
			{"role": "user", "content": content},
		},
		"max_tokens": 2048,
	}
	payloadBytes, _ := json.Marshal(payload)

	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequest("POST", fp.Endpoint, strings.NewReader(string(payloadBytes)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+fp.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		logger.SysError(fmt.Sprintf("chat request failed: %s %s %v", provider, modelId, err))
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("Chat request failed: %v", err),
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("Upstream returned %d: %s", resp.StatusCode, string(body)),
		})
		return
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func RegisterFreeChatRoutes(r *gin.RouterGroup) {
	r.GET("/free-chat/providers", GetFreeChatProviders)
	r.GET("/free-chat/chat", Chat)
}
