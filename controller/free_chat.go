package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/middleware"
	"github.com/quantumclaw/quantumclaw/model"
)

// ── 免费模型提供商配置 ─────────────────────────────────────
// API keys 可从 PlatformConfig 表读取（通过管理页面配置）
// 如果 DB 中没有，则回退到环境变量

const (
	configKeyGroqKey        = "free_chat_groq_key"
	configKeyDeepseekKey    = "free_chat_deepseek_key"
	configKeyGeminiKey      = "free_chat_gemini_key"
	configKeySiliconFlowKey = "free_chat_siliconflow_key"
	configKeyMistralKey     = "free_chat_mistral_key"
)

// getFreeChatAPIKey 从 DB PlatformConfig 获取 API key，不存在则回退到 env
func getFreeChatAPIKey(configKey, envKey, envFallback string) string {
	// 先查 DB
	if model.DB != nil {
		var cfg model.PlatformConfig
		if err := model.DB.Where("`key` = ?", configKey).First(&cfg).Error; err == nil && cfg.Value != "" {
			return cfg.Value
		}
	}
	// 回退到包级别的 env 值
	if envFallback != "" {
		return envFallback
	}
	return os.Getenv(envKey)
}

var envGroqKey        = os.Getenv("GROQ_API_KEY")
var envDeepseekKey    = os.Getenv("DEEPSEEK_API_KEY")
var envGeminiKey      = os.Getenv("GEMINI_API_KEY")
var envSiliconFlowKey = os.Getenv("SILICONFLOW_API_KEY")
var envMistralKey     = os.Getenv("MISTRAL_API_KEY")

type freeProvider struct {
	Name     string
	Endpoint string
	EnvKey   string
	ConfigKey string
	Models   []map[string]string
}

// resolveProviders 每次调用时从 DB 读取最新的 API key
func resolveProviders() []freeProvider {
	base := []freeProvider{
		{
			Name:      "groq",
			Endpoint:  "https://api.groq.com/openai/v1/chat/completions",
			EnvKey:    "GROQ_API_KEY",
			ConfigKey: configKeyGroqKey,
			Models: []map[string]string{
				{"id": "llama-3.3-70b-versatile", "name": "Llama 3.3 70B"},
				{"id": "llama-3.1-8b-instant", "name": "Llama 3.1 8B"},
				{"id": "mixtral-8x7b-32768", "name": "Mixtral 8x7B"},
				{"id": "gemma2-9b-it", "name": "Gemma 2 9B"},
			},
		},
		{
			Name:      "deepseek",
			Endpoint:  "https://api.deepseek.com/v1/chat/completions",
			EnvKey:    "DEEPSEEK_API_KEY",
			ConfigKey: configKeyDeepseekKey,
			Models: []map[string]string{
				{"id": "deepseek-chat", "name": "DeepSeek V3"},
				{"id": "deepseek-reasoner", "name": "DeepSeek R1"},
			},
		},
		{
			Name:      "gemini",
			Endpoint:  "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions",
			EnvKey:    "GEMINI_API_KEY",
			ConfigKey: configKeyGeminiKey,
			Models: []map[string]string{
				{"id": "gemini-2.5-flash-preview-04-17", "name": "Gemini 2.5 Flash"},
				{"id": "gemini-2.0-flash", "name": "Gemini 2.0 Flash"},
				{"id": "gemini-2.0-flash-lite", "name": "Gemini 2.0 Flash Lite"},
			},
		},
		{
			Name:      "siliconflow",
			Endpoint:  "https://api.siliconflow.cn/v1/chat/completions",
			EnvKey:    "SILICONFLOW_API_KEY",
			ConfigKey: configKeySiliconFlowKey,
			Models: []map[string]string{
				{"id": "Qwen/Qwen2.5-72B-Instruct", "name": "Qwen 2.5 72B"},
				{"id": "Qwen/Qwen2.5-32B-Instruct", "name": "Qwen 2.5 32B"},
				{"id": "Qwen/Qwen2.5-7B-Instruct", "name": "Qwen 2.5 7B"},
				{"id": "deepseek-ai/DeepSeek-V3", "name": "DeepSeek V3"},
			},
		},
		{
			Name:      "mistral",
			Endpoint:  "https://api.mistral.ai/v1/chat/completions",
			EnvKey:    "MISTRAL_API_KEY",
			ConfigKey: configKeyMistralKey,
			Models: []map[string]string{
				{"id": "mistral-small-latest", "name": "Mistral Small"},
				{"id": "open-mistral-nemo", "name": "Mistral Nemo"},
			},
		},
	}
	for i := range base {
		base[i].APIKey = getFreeChatAPIKey(base[i].ConfigKey, base[i].EnvKey, "")
	}
	return base
}

// GetFreeChatProviders 返回当前配置的免费聊天提供商列表（用于前端显示）
func GetFreeChatProviders(c *gin.Context) {
	providers := resolveProviders()
	type providerResp struct {
		Name     string              `json:"name"`
		EnvKey   string              `json:"env_key"`
		ConfigKey string             `json:"config_key"`
		Configured bool              `json:"configured"`
		Models   []map[string]string `json:"models"`
	}
	resp := make([]providerResp, 0, len(providers))
	for _, p := range providers {
		resp = append(resp, providerResp{
			Name:       p.Name,
			EnvKey:     p.EnvKey,
			ConfigKey:  p.ConfigKey,
			Configured: p.APIKey != "",
			Models:     p.Models,
		})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

// ── 聊天接口 ────────────────────────────────────────────

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

	// 每次请求实时解析 providers（DB 配置变更立即生效）
	resolved := resolveProviders()

	var fp *freeProvider
	for _, p := range resolved {
		if p.Name == provider {
			fp = &p
			break
		}
	}
	if fp == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown provider: " + provider})
		return
	}

	if fp.APIKey == "" {
		logger.SysError(fmt.Sprintf("免费聊天: %s 未配置 (设置 %s 环境变量或通过管理页面填写)", provider, fp.EnvKey))
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("免费聊天: %s 未配置 (设置 %s 环境变量或通过管理页面填写)", provider, fp.EnvKey),
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

	// 构建 OpenAI 兼容请求
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
