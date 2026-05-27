package model

import (
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/relay/channeltype"
)

// DefaultChannelModels — 每个渠道类型常见的默认模型列表
var DefaultChannelModels = map[int]string{
	channeltype.OpenAI:        "gpt-4o,gpt-4o-mini,gpt-4-turbo,gpt-4,gpt-3.5-turbo,o1,o1-mini,o3-mini,dall-e-3,tts-1,whisper-1,text-embedding-3-small,text-embedding-3-large",
	channeltype.Anthropic:     "claude-3-5-sonnet-20241022,claude-3-5-haiku-20241022,claude-3-opus,claude-3-sonnet,claude-3-haiku,claude-opus-4-20250514,claude-sonnet-4-20250514",
	channeltype.Gemini:        "gemini-1.5-pro,gemini-1.5-flash,gemini-2.0-flash,gemini-2.0-flash-lite,gemini-2.5-pro-preview-05-07,gemini-2.5-flash-preview-04-17",
	channeltype.DeepSeek:      "deepseek-chat,deepseek-v3,deepseek-v3-0324,deepseek-reasoner,deepseek-r1,deepseek-r1-0528,deepseek-r1-distill-qwen-1.5b,deepseek-r1-distill-qwen-7b,deepseek-r1-distill-qwen-14b,deepseek-r1-distill-qwen-32b,deepseek-r1-distill-llama-8b,deepseek-r1-distill-llama-70b",
	channeltype.Groq:          "mixtral-8x7b-32768,llama-3.3-70b-versatile,llama-3.1-8b-instant",
	channeltype.SiliconFlow:   "Pro/deepseek-ai/DeepSeek-V3,Pro/Qwen/Qwen2.5-72B-Instruct",
	channeltype.Moonshot:      "kimi-latest,moonshot-v1-8k,moonshot-v1-32k",
	channeltype.Ali:           "qwen-turbo,qwen-plus,qwen-max,qwen-turbo-latest,qwen-plus-latest,qwen-max-latest,qwen2.5-72b-instruct,qwen2.5-32b-instruct,qwen2.5-14b-instruct,qwen2.5-7b-instruct,qwen2.5-coder-32b-instruct,qwen2.5-vl-72b-instruct,qwen2.5-vl-32b-instruct,qwen-vl-plus",
	channeltype.Baidu:         "ERNIE-4.5-Turbo-8K,ERNIE-3.5-8K",
	channeltype.Zhipu:         "glm-4-plus,glm-4,glm-4-flash",
	channeltype.Tencent:       "hunyuan-turbo,hunyuan-lite",
	channeltype.Minimax:       "minimax-text-01,abab6.5s-chat",
	channeltype.Mistral:       "mistral-large-latest,open-mistral-nemo",
	channeltype.QuantumClaw:    "quantumclaw/auto",
	channeltype.Doubao:        "doubao-pro-4k,doubao-lite-32k",
	channeltype.StepFun:       "step-2-16k",
	channeltype.AwsClaude:     "claude-3-5-sonnet-20241022",
	channeltype.Cohere:        "command-r-plus,command-r",
	channeltype.Cloudflare:    "@cf/meta/llama-3.1-8b-instruct",
	channeltype.DeepL:         "default",
	channeltype.TogetherAI:    "togethercomputer/stripedhyena-nous-7b",
	channeltype.XAI:           "grok-beta,grok-2-1212",
}

// SeedDefaultChannels — 检测 channels 表为空时，自动插入预设渠道
func SeedDefaultChannels() {
	var count int64
	DB.Model(&Channel{}).Count(&count)
	if count > 0 {
		return // 已有渠道，跳过
	}

	now := time.Now().Unix()

	// AI 大模型渠道 (type 1~60, 按 ChannelTypeNames 中有名称的)
	aiChannels := []int{
		channeltype.OpenAI, channeltype.Anthropic, channeltype.Gemini,
		channeltype.DeepSeek, channeltype.Groq, channeltype.SiliconFlow,
		channeltype.Moonshot, channeltype.Ali, channeltype.Baidu,
		channeltype.Zhipu, channeltype.Tencent, channeltype.Minimax,
		channeltype.Mistral, channeltype.QuantumClaw, channeltype.Doubao,
		channeltype.StepFun, channeltype.Cohere,
		channeltype.Cloudflare, channeltype.DeepL, channeltype.TogetherAI,
		channeltype.XAI, channeltype.Ollama,
		channeltype.ZhipuV4, channeltype.AwsClaude,
	}

	// 量子算力渠道 (type >= 100)
	quantumChannels := []int{
		channeltype.IonQ, channeltype.IBMQ, channeltype.Rigetti,
		channeltype.AWSBraket, channeltype.AzureQuantum, channeltype.GoogleQuantum,
	}

	// 插入 AI 渠道
	for _, t := range aiChannels {
		name := channeltype.ChannelTypeNames[t]
		if name == "" {
			continue
		}
		baseURL := ""
		if t >= 0 && t < len(channeltype.ChannelBaseURLs) {
			baseURL = channeltype.ChannelBaseURLs[t]
		}
		models := DefaultChannelModels[t]

		ch := &Channel{
			Type:        t,
			Name:        name,
			Key:         "PUT_YOUR_API_KEY_HERE",
			BaseURL:     StrPtr(baseURL),
			Models:      models,
			Group:       "default",
			Status:      ChannelStatusEnabled,
			Weight:      UintPtr(1),
			CreatedTime: now,
			Category:    "paid",
		}
		if err := DB.Create(ch).Error; err != nil {
			logger.SysError("failed to seed channel: " + name + " - " + err.Error())
			continue
		}
		// 同步更新 Ability 表，否则 Distribute 找不到渠道
		if err := ch.AddAbilities(); err != nil {
			logger.SysError("failed to seed abilities for: " + name + " - " + err.Error())
		}
	}

	// 插入量子算力渠道
	for _, t := range quantumChannels {
		name := channeltype.ChannelTypeNames[t]
		if name == "" {
			continue
		}
		ch := &Channel{
			Type:        t,
			Name:        name,
			Key:         "PUT_YOUR_API_KEY_HERE",
			BaseURL:     StrPtr(""),
			Models:      "",
			Group:       "quantum",
			Status:      ChannelStatusEnabled,
			Weight:      UintPtr(1),
			CreatedTime: now,
			Category:    "paid",
		}
		if err := DB.Create(ch).Error; err != nil {
			logger.SysError("failed to seed quantum channel: " + name + " - " + err.Error())
			continue
		}
		if err := ch.AddAbilities(); err != nil {
			logger.SysError("failed to seed abilities for: " + name + " - " + err.Error())
		}
	}

	logger.SysLog("default channels seeded: " + formatInt(len(aiChannels)) + " AI + " + formatInt(len(quantumChannels)) + " Quantum = " + formatInt(len(aiChannels)+len(quantumChannels)))
}

func StrPtr(s string) *string { return &s }
func UintPtr(u uint) *uint    { return &u }
func formatInt(n int) string {
	if n == 0 {
		return "0"
	}
	return itoa(n)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
