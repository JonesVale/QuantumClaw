package channeltype

import (
	"testing"

	"github.com/quantumclaw/quantumclaw/relay/apitype"
	"github.com/stretchr/testify/assert"
)

func TestToAPIType_KnownAPIs(t *testing.T) {
	tests := []struct {
		name       string
		channelType int
		wantAPIType int
	}{
		// 已知独立 API 类型
		{"Anthropic -> Anthropic", Anthropic, apitype.Anthropic},
		{"Baidu -> Baidu", Baidu, apitype.Baidu},
		{"PaLM -> PaLM", PaLM, apitype.PaLM},
		{"Zhipu -> Zhipu", Zhipu, apitype.Zhipu},
		{"Ali -> Ali", Ali, apitype.Ali},
		{"Xunfei -> Xunfei", Xunfei, apitype.Xunfei},
		{"AIProxyLibrary -> AIProxyLibrary", AIProxyLibrary, apitype.AIProxyLibrary},
		{"Tencent -> Tencent", Tencent, apitype.Tencent},
		{"Gemini -> Gemini", Gemini, apitype.Gemini},
		{"Ollama -> Ollama", Ollama, apitype.Ollama},
		{"AwsClaude -> AwsClaude", AwsClaude, apitype.AwsClaude},
		{"Coze -> Coze", Coze, apitype.Coze},
		{"Cohere -> Cohere", Cohere, apitype.Cohere},
		{"Cloudflare -> Cloudflare", Cloudflare, apitype.Cloudflare},
		{"DeepL -> DeepL", DeepL, apitype.DeepL},
		{"VertextAI -> VertexAI", VertextAI, apitype.VertexAI},
		{"Replicate -> Replicate", Replicate, apitype.Replicate},
		{"Proxy -> Proxy", Proxy, apitype.Proxy},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToAPIType(tt.channelType)
			assert.Equal(t, tt.wantAPIType, got, "ToAPIType(%d) = %d, want %d", tt.channelType, got, tt.wantAPIType)
		})
	}
}

func TestToAPIType_OpenAICompatible(t *testing.T) {
	// 所有 OpenAI-Compatible 渠道必须映射到 apitype.OpenAI
	// 这是本次修复的核心：修复前这些渠道映射到 0（默认值），导致 relay 路由失败
	openAICompatibleChannels := []struct {
		name        string
		channelType int
	}{
		{"DeepSeek", DeepSeek},
		{"Doubao", Doubao},
		{"Minimax", Minimax},
		{"Groq", Groq},
		{"Mistral", Mistral},
		{"Novita", Novita},
		{"Baichuan", Baichuan},
		{"Moonshot", Moonshot},
		{"LingYiWanWu", LingYiWanWu},
		{"StepFun", StepFun},
		{"TogetherAI", TogetherAI},
		{"SiliconFlow", SiliconFlow},
		{"XAI", XAI},
		{"BaiduV2", BaiduV2},
		{"XunfeiV2", XunfeiV2},
		{"AliBailian", AliBailian},
		{"OpenRouter", OpenRouter},
		{"GeminiOpenAICompatible", GeminiOpenAICompatible},
		{"OpenAICompatible", OpenAICompatible},
	}

	for _, tt := range openAICompatibleChannels {
		t.Run(tt.name, func(t *testing.T) {
			got := ToAPIType(tt.channelType)
			assert.Equal(t, apitype.OpenAI, got,
				"渠道 %s (channelType=%d) 应映射到 apitype.OpenAI(%d)，实际得到 %d",
				tt.name, tt.channelType, apitype.OpenAI, got)
		})
	}
}

func TestToAPIType_UnknownDefaultsToOpenAI(t *testing.T) {
	// 未知 channelType 应默认返回 apitype.OpenAI（向后兼容）
	// apitype.OpenAI = 0（iota 第一个值），所以返回值是 0 是正确的
	dummy := Dummy
	for i := 1; i <= 50; i++ { // 从 Dummy+1 开始，排除 Dummy（Dummy 是最后一个有效值）
		channelType := dummy + i
		got := ToAPIType(channelType)
		assert.Equal(t, apitype.OpenAI, got,
			"ToAPIType(%d) 应返回 apitype.OpenAI(0)，实际得到 %d", channelType, got)
	}
}

func TestToAPIType_DummyIsLastValid(t *testing.T) {
	// Dummy 是最后一个有效 channelType，未在 switch 中显式处理
	// 默认返回 apitype.OpenAI（向后兼容）
	got := ToAPIType(Dummy)
	assert.Equal(t, apitype.OpenAI, got, "Dummy 默认返回 OpenAI（向后兼容）")
}
