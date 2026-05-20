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
)

// ── 免费模型提供商配置 ─────────────────────────────────────

var (
	freeChatGroqKey        = os.Getenv("GROQ_API_KEY")
	freeChatDeepseekKey    = os.Getenv("DEEPSEEK_API_KEY")
	freeChatGeminiKey      = os.Getenv("GEMINI_API_KEY")
	freeChatSiliconFlowKey = os.Getenv("SILICONFLOW_API_KEY")
	freeChatMistralKey     = os.Getenv("MISTRAL_API_KEY")
)

type freeProvider struct {
	Name     string
	Endpoint string
	EnvKey   string
	APIKey   string
	Models   []map[string]string
}

var freeProviders = []freeProvider{
	{
		Name:     "groq",
		Endpoint: "https://api.groq.com/openai/v1/chat/completions",
		EnvKey:   "GROQ_API_KEY",
		APIKey:   freeChatGroqKey,
		Models: []map[string]string{
			{"id": "llama-3.3-70b-versatile", "name": "Llama 3.3 70B"},
			{"id": "llama-3.1-8b-instant", "name": "Llama 3.1 8B"},
			{"id": "mixtral-8x7b-32768", "name": "Mixtral 8x7B"},
			{"id": "gemma2-9b-it", "name": "Gemma 2 9B"},
		},
	},
	{
		Name:     "deepseek",
		Endpoint: "https://api.deepseek.com/v1/chat/completions",
		EnvKey:   "DEEPSEEK_API_KEY",
		APIKey:   freeChatDeepseekKey,
		Models: []map[string]string{
			{"id": "deepseek-chat", "name": "DeepSeek V3"},
			{"id": "deepseek-reasoner", "name": "DeepSeek R1"},
		},
	},
	{
		Name:     "gemini",
		Endpoint: "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions",
		EnvKey:   "GEMINI_API_KEY",
		APIKey:   freeChatGeminiKey,
		Models: []map[string]string{
			{"id": "gemini-2.5-flash-preview-04-17", "name": "Gemini 2.5 Flash"},
			{"id": "gemini-2.0-flash", "name": "Gemini 2.0 Flash"},
			{"id": "gemini-2.0-flash-lite", "name": "Gemini 2.0 Flash Lite"},
		},
	},
	{
		Name:     "siliconflow",
		Endpoint: "https://api.siliconflow.cn/v1/chat/completions",
		EnvKey:   "SILICONFLOW_API_KEY",
		APIKey:   freeChatSiliconFlowKey,
		Models: []map[string]string{
			{"id": "Qwen/Qwen2.5-72B-Instruct", "name": "Qwen 2.5 72B"},
			{"id": "Qwen/Qwen2.5-32B-Instruct", "name": "Qwen 2.5 32B"},
			{"id": "Qwen/Qwen2.5-7B-Instruct", "name": "Qwen 2.5 7B"},
			{"id": "deepseek-ai/DeepSeek-V3", "name": "DeepSeek V3"},
			{"id": "Pro/Qwen/Qwen2.5-7B-Instruct", "name": "Qwen 2.5 7B Pro"},
		},
	},
	{
		Name:     "mistral",
		Endpoint: "https://api.mistral.ai/v1/chat/completions",
		EnvKey:   "MISTRAL_API_KEY",
		APIKey:   freeChatMistralKey,
		Models: []map[string]string{
			{"id": "mistral-small-latest", "name": "Mistral Small"},
			{"id": "mistral-nemo-latest", "name": "Mistral Nemo"},
			{"id": "open-mistral-nemo", "name": "Open Mistral Nemo"},
		},
	},
}

// FreeChatRequest 免费聊天请求
type FreeChatRequest struct {
	Provider string `json:"provider"` // groq | deepseek
	Model    string `json:"model"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
}

// FreeChatResponse 响应
type FreeChatResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func init() {
	for i, p := range freeProviders {
		if p.APIKey == "" {
			logger.SysError(fmt.Sprintf("免费聊天: %s 未配置 (设置 %s 环境变量)", p.Name, p.EnvKey))
		}
		freeProviders[i].APIKey = p.APIKey
	}
}

func getProvider(name string) *freeProvider {
	for _, p := range freeProviders {
		if p.Name == name {
			return &p
		}
	}
	return nil
}

// FreeChat 免费聊天代理
func FreeChat(c *gin.Context) {
	var req FreeChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, FreeChatResponse{Success: false, Message: "请求参数错误"})
		return
	}
	if len(req.Messages) == 0 {
		c.JSON(http.StatusBadRequest, FreeChatResponse{Success: false, Message: "消息不能为空"})
		return
	}

	// 确定提供商
	provider := getProvider(req.Provider)
	if provider == nil {
		provider = getProvider("groq") // 默认 Groq
	}
	if provider.APIKey == "" {
		c.JSON(http.StatusServiceUnavailable, FreeChatResponse{
			Success: false,
			Message: fmt.Sprintf("%s 未配置，请联系管理员设置 %s", provider.Name, provider.EnvKey),
		})
		return
	}

	model := req.Model
	if model == "" {
		model = provider.Models[0]["id"]
	}

	// 构建请求
	body := map[string]interface{}{
		"model":       model,
		"messages":    req.Messages,
		"stream":      true,
		"max_tokens":  4096,
		"temperature": 0.7,
	}
	bodyJSON, _ := json.Marshal(body)

	httpReq, err := http.NewRequest(http.MethodPost, provider.Endpoint, strings.NewReader(string(bodyJSON)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, FreeChatResponse{Success: false, Message: "创建请求失败"})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+provider.APIKey)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		logger.SysError(fmt.Sprintf("%s proxy error: %s", provider.Name, err.Error()))
		c.JSON(http.StatusBadGateway, FreeChatResponse{Success: false, Message: fmt.Sprintf("%s 暂时不可用", provider.Name)})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		c.JSON(http.StatusBadGateway, FreeChatResponse{
			Success: false,
			Message: fmt.Sprintf("%s 返回错误: %s", provider.Name, string(errBody)),
		})
		return
	}

	// 转发流式
	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := c.Writer.Write(buf[:n]); writeErr != nil {
				break
			}
			c.Writer.Flush()
		}
		if err != nil {
			break
		}
	}
}

// FreeChatModels 获取所有可用免费模型
func FreeChatModels(c *gin.Context) {
	type providerResp struct {
		Name   string              `json:"name"`
		Label  string              `json:"label"`
		Ready  bool                `json:"ready"`
		Models []map[string]string `json:"models"`
	}
	var list []providerResp
	for _, p := range freeProviders {
		list = append(list, providerResp{
			Name:   p.Name,
			Label:  p.Name,
			Ready:  p.APIKey != "",
			Models: p.Models,
		})
	}
	c.JSON(http.StatusOK, FreeChatResponse{Success: true, Data: list})
}

// FreeChatStatus 检测连通性
func FreeChatStatus(c *gin.Context) {
	provider := c.Query("provider")
	if provider == "" {
		provider = "groq"
	}
	p := getProvider(provider)
	if p == nil {
		c.JSON(http.StatusOK, FreeChatResponse{Success: false, Message: "未知提供商"})
		return
	}
	if p.APIKey == "" {
		c.JSON(http.StatusOK, FreeChatResponse{Success: false, Message: fmt.Sprintf("%s 未配置", p.Name)})
		return
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(p.Endpoint)
	if err != nil {
		c.JSON(http.StatusOK, FreeChatResponse{Success: false, Message: fmt.Sprintf("%s 不可达", p.Name)})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusOK {
		c.JSON(http.StatusOK, FreeChatResponse{Success: true, Message: fmt.Sprintf("%s 正常", p.Name)})
	} else {
		c.JSON(http.StatusOK, FreeChatResponse{Success: false, Message: fmt.Sprintf("%s 异常", p.Name)})
	}
}

// RegisterFreeChatRoutes 注册免费聊天路由
func RegisterFreeChatRoutes(r *gin.RouterGroup) {
	freeGroup := r.Group("/free-chat")
	freeGroup.Use(middleware.GlobalAPIRateLimit())
	{
		freeGroup.POST("/completions", FreeChat)
		freeGroup.GET("/models", FreeChatModels)
		freeGroup.GET("/status", FreeChatStatus)
	}
}
