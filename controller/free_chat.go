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

// ── 免费聊天：从 channels 表读取可用提供商 ───────────────
// 不再需要环境变量。管理员或代理商在 /channels 配置渠道，
// 免费聊天自动发现 type 在免费聊天范围内、status=enabled、key 不为空的渠道。

// freeChatEligibleTypes — 哪些 channel type 可用于免费聊天
var freeChatEligibleTypes = map[int]bool{
	channeltype.Groq:        true,
	channeltype.DeepSeek:    true,
	channeltype.Gemini:      true,
	channeltype.SiliconFlow: true,
	channeltype.Mistral:     true,
}

type freeChatProvider struct {
	Type       int                `json:"type"`
	Name       string             `json:"name"`
	ChannelID  int                `json:"channel_id"`
	Endpoint   string             `json:"endpoint"`
	APIKey     string             `json:"-"`
	Models     []map[string]string `json:"models"`
}

// resolveFreeChatProviders 从 channels 表读取，过滤出可用的免费聊天 provider
func resolveFreeChatProviders() []freeChatProvider {
	var channels []model.Channel
	model.DB.Where("status = ? AND key != '' AND key NOT LIKE ? AND type IN (?)",
		model.ChannelStatusEnabled, "PUT_YOUR%", []int{29, 36, 24, 44, 28}).Find(&channels)

	typeNames := channeltype.ChannelTypeNames
	providerMap := make(map[int]*freeChatProvider)

	for _, ch := range channels {
		if !freeChatEligibleTypes[ch.Type] {
			continue
		}
		// Decrypt key
		key := ch.Key
		if model.CryptoSecret != "" {
			if dec, err := decryptWithSecret(key); err == nil {
				key = dec
			}
		}
		if key == "" {
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
			// Fallback to known endpoints
			if endpoint == "" {
				endpoint = getDefaultEndpoint(ch.Type)
			}

			providerMap[ch.Type] = &freeChatProvider{
				Type:      ch.Type,
				Name:      pName,
				ChannelID: ch.Id,
				Endpoint:  endpoint,
				APIKey:    key,
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

func decryptWithSecret(encrypted string) (string, error) {
	// Reuse the existing encrypt package logic
	// If CryptoSecret is set, the key was encrypted at insert time
	// For now, we assume if key starts with a certain prefix or was encrypted
	return encrypted, nil
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

// GetFreeChatProviders 返回当前可用的免费聊天提供商列表
func GetFreeChatProviders(c *gin.Context) {
	providers := resolveFreeChatProviders()
	type providerResp struct {
		Type       int                `json:"type"`
		Name       string             `json:"name"`
		ChannelID  int                `json:"channel_id"`
		Configured bool               `json:"configured"`
		Models     []map[string]string `json:"models"`
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
	model := c.Query("model")
	content := c.Query("content")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider is required"})
		return
	}
	if model == "" {
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
		n := strings.ToLower(p.Name)
		if n == strings.ToLower(provider) {
			fp = &p
			break
		}
	}
	if fp == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown provider: " + provider})
		return
	}

	if fp.APIKey == "" {
		logger.SysError(fmt.Sprintf("免费聊天: %s 未配置 API Key", provider))
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("免费聊天: %s 未配置 API Key，请管理员在渠道设置中配置", provider),
		})
		return
	}

	// Validate model
	validModel := false
	for _, m := range fp.Models {
		if m["id"] == model {
			validModel = true
			break
		}
	}
	if !validModel {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model '" + model + "' is not available for provider '" + provider + "'"})
		return
	}

	// Build OpenAI-compatible request
	payload := map[string]interface{}{
		"model": model,
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
		logger.SysError(fmt.Sprintf("聊天请求失败: %s %s %v", provider, model, err))
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("聊天请求失败: %v", err),
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("上游返回 %d: %s", resp.StatusCode, string(body)),
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
