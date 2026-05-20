package channeltype

import "fmt"

var ChannelBaseURLs = []string{
	"",                              // 0  Unknown
	"https://api.openai.com",        // 1  OpenAI
	"https://oa.api2d.net",          // 2  API2D
	"",                              // 3  Azure
	"https://api.closeai-proxy.xyz", // 4  CloseAI
	"https://api.openai-sb.com",     // 5  OpenAISB
	"https://api.openaimax.com",     // 6  OpenAIMax
	"https://api.ohmygpt.com",       // 7  OhMyGPT
	"",                              // 8  Custom
	"https://api.caipacity.com",     // 9  Ails
	"https://api.aiproxy.io",        // 10 AIProxy
	"https://generativelanguage.googleapis.com", // 11 PaLM
	"https://api.api2gpt.com",                   // 12 API2GPT
	"https://api.aigc2d.com",                    // 13 AIGC2D
	"https://api.anthropic.com",                 // 14 Anthropic
	"https://aip.baidubce.com",                  // 15 Baidu
	"https://open.bigmodel.cn",                  // 16 Zhipu
	"https://dashscope.aliyuncs.com",            // 17 Ali
	"",                                          // 18 Xunfei
	"https://ai.360.cn",                         // 19 AI360
	"https://openrouter.ai/api",                 // 20 OpenRouter
	"https://api.aiproxy.io",                    // 21 AIProxyLibrary
	"https://fastgpt.run/api/openapi",           // 22 FastGPT
	"https://hunyuan.tencentcloudapi.com",       // 23 Tencent
	"https://generativelanguage.googleapis.com", // 24 Gemini
	"https://api.moonshot.cn",                   // 25 Moonshot
	"https://api.baichuan-ai.com",               // 26 Baichuan
	"https://api.minimax.chat",                  // 27 Minimax
	"https://api.mistral.ai",                    // 28 Mistral
	"https://api.groq.com/openai",               // 29 Groq
	"http://localhost:11434",                    // 30 Ollama
	"https://api.lingyiwanwu.com",               // 31 LingYiWanWu
	"https://api.stepfun.com",                   // 32 StepFun
	"",                                          // 33 AwsClaude
	"https://api.coze.com",                      // 34 Coze
	"https://api.cohere.ai",                     // 35 Cohere
	"https://api.deepseek.com",                  // 36 DeepSeek
	"https://api.cloudflare.com",                // 37 Cloudflare
	"https://api-free.deepl.com",                // 38 DeepL
	"https://api.together.xyz",                  // 39 TogetherAI
	"https://ark.cn-beijing.volces.com",         // 40 Doubao
	"https://api.novita.ai/v3/openai",           // 41 Novita
	"",                                          // 42 VertextAI
	"",                                          // 43 Proxy
	"https://api.siliconflow.cn",                // 44 SiliconFlow
	"https://api.x.ai",                          // 45 XAI
	"https://api.replicate.com/v1/models/",      // 46 Replicate
	"https://qianfan.baidubce.com",              // 47 BaiduV2
	"https://spark-api-open.xf-yun.com",         // 48 XunfeiV2
	"https://dashscope.aliyuncs.com",            // 49 AliBailian
	"",                                          // 50 OpenAICompatible
	"https://generativelanguage.googleapis.com/v1beta/openai/", // 51 GeminiOpenAICompatible
	"",                                          // 52 Codex
	"",                                          // 53 Dify
	"",                                          // 54 Jimeng
	"",                                          // 55 Jina
	"",                                          // 56 MokaAI
	"",                                          // 57 Submodel
	"https://ark.cn-beijing.volces.com",         // 58 VolcEngine
	"",                                          // 59 Xinference
	"https://open.bigmodel.cn",                  // 60 ZhipuV4
	"",                                          // 61 Dummy

	// ==================== 量子算力 Base URLs (type >= 100, 空缺62-99) ====================
	"", "", "", "", "", "", "", "", "", "", // 62-71
	"", "", "", "", "", "", "", "", "", "", // 72-81
	"", "", "", "", "", "", "", "", "", "", // 82-91
	"", "", "", "", "", "", "", "",          // 92-99
	"",                                          // 100 IonQ
	"",                                          // 101 IBMQ
	"",                                          // 102 Rigetti
	"",                                          // 103 AWSBraket
	"",                                          // 104 AzureQuantum
	"",                                          // 105 GoogleQuantum
}

// ValidateChannelBaseURLs 验证 ChannelBaseURLs 长度是否与 QuantumDummy 常量匹配
// 若不一致返回错误，避免运行时 panic
func ValidateChannelBaseURLs() error {
	if len(ChannelBaseURLs) != QuantumDummy {
		return fmt.Errorf("channel base urls length %d does not match QuantumDummy %d", len(ChannelBaseURLs), QuantumDummy)
	}
	return nil
}
