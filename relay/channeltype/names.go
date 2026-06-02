package channeltype

// ChannelTypeNames — 所有渠道类型的名称映射
// key = channelType int, value = 显示名称（英文技术标识符）
// 前端通过 GET /api/channel/types 动态获取，避免硬编码
var ChannelTypeNames = map[int]string{
	// ==================== AI 大模型渠道 (0~61) ====================
	Unknown:              "Unknown",
	OpenAI:               "OpenAI",
	API2D:                "API2D",
	Azure:                "Azure",
	CloseAI:              "CloseAI",
	OpenAISB:             "OpenAI SB",
	OpenAIMax:            "OpenAI Max",
	OhMyGPT:              "OhMyGPT",
	Custom:               "Custom",
	Ails:                 "Ails",
	AIProxy:              "AI Proxy",
	PaLM:                 "Google PaLM",
	API2GPT:              "API2GPT",
	AIGC2D:               "AIGC2D",
	Anthropic:            "Anthropic",
	Baidu:                "Baidu",
	Zhipu:                "Zhipu (GLM)",
	Ali:                  "Ali (Qwen)",
	Xunfei:               "Xunfei (Spark)",
	AI360:                "360 AI",
	QuantumClaw:           "API Relay",
	AIProxyLibrary:       "AIProxy Library",
	FastGPT:              "FastGPT",
	Tencent:              "Tencent (Hunyuan)",
	Gemini:               "Google Gemini",
	Moonshot:             "Moonshot (Kimi)",
	Baichuan:             "Baichuan",
	Minimax:              "MiniMax",
	Mistral:              "Mistral AI",
	Groq:                 "Groq",
	Ollama:               "Ollama",
	LingYiWanWu:          "LingYi (01.AI)",
	StepFun:              "StepFun (Step-2)",
	AwsClaude:            "AWS Claude",
	Coze:                 "Coze",
	Cohere:               "Cohere",
	DeepSeek:             "DeepSeek",
	Cloudflare:           "Cloudflare",
	DeepL:                "DeepL",
	TogetherAI:           "Together AI",
	Doubao:               "Doubao (VolcEngine)",
	Novita:               "Novita AI",
	VertextAI:            "Google Vertex AI",
	Proxy:                "Proxy",
	SiliconFlow:          "SiliconFlow",
	XAI:                  "xAI (Grok)",
	Replicate:            "Replicate",
	BaiduV2:              "Baidu V2",
	XunfeiV2:             "Xunfei V2",
	AliBailian:           "Ali Bailian",
	OpenAICompatible:     "OpenAI Compatible",
	GeminiOpenAICompatible: "Gemini OpenAI Compatible",
	Codex:                "Codex",
	Dify:                 "Dify",
	Jimeng:               "Jimeng",
	Jina:                 "Jina AI",
	MokaAI:               "Moka AI",
	Submodel:             "Submodel",
	VolcEngine:           "VolcEngine (Ark)",
	Xinference:           "Xinference",
	ZhipuV4:              "Zhipu V4",

	// ==================== 量子算力渠道 (type >= 100) ====================
	IonQ:         "IonQ",
	IBMQ:         "IBM Q",
	Rigetti:      "Rigetti",
	AWSBraket:    "AWS Braket",
	AzureQuantum: "Azure Quantum",
	GoogleQuantum: "Google Quantum",
}

const (
	// 区域常量
	RegionChina    = "china"
	RegionOverseas = "overseas"
)

// ChannelTypeRegion 渠道类型 → 区域映射（国内/国外）
// 用于 routing 时按 region 过滤，确保国内资源不自动切到国外
var ChannelTypeRegion = map[int]string{
	// ── 国内 AI 渠道 ──
	Baidu:        RegionChina,
	Zhipu:        RegionChina,
	Ali:          RegionChina,
	Xunfei:       RegionChina,
	AI360:        RegionChina,
	Tencent:      RegionChina,
	Baichuan:     RegionChina,
	Minimax:      RegionChina,
	Moonshot:     RegionChina,
	LingYiWanWu:  RegionChina,
	StepFun:      RegionChina,
	Doubao:       RegionChina,
	XunfeiV2:     RegionChina,
	BaiduV2:      RegionChina,
	AliBailian:   RegionChina,
	ZhipuV4:      RegionChina,
	Jimeng:       RegionChina,
	VolcEngine:   RegionChina,
	MokaAI:       RegionChina,
	DeepSeek:     RegionChina,

	// ── 国内量子算力 ──
	// (预留)

	// ── 国外 AI 渠道 ──
	OpenAI:                RegionOverseas,
	Anthropic:             RegionOverseas,
	Gemini:                RegionOverseas,
	PaLM:                  RegionOverseas,
	Mistral:               RegionOverseas,
	AwsClaude:             RegionOverseas,
	Azure:                 RegionOverseas,
	Cohere:                RegionOverseas,
	Groq:                  RegionOverseas,
	DeepL:                 RegionOverseas,
	TogetherAI:            RegionOverseas,
	Cloudflare:            RegionOverseas,
	Novita:                RegionOverseas,
	VertextAI:             RegionOverseas,
	SiliconFlow:           RegionOverseas,
	XAI:                   RegionOverseas,
	Replicate:             RegionOverseas,
	OpenAICompatible:      RegionOverseas,
	GeminiOpenAICompatible: RegionOverseas,
	Codex:                 RegionOverseas,
	Proxy:                 RegionOverseas,
	Coze:                  RegionOverseas,
	Jina:                  RegionOverseas,
	Dify:                  RegionOverseas,
	Submodel:              RegionOverseas,
	Xinference:            RegionOverseas,
	Sub2API:               RegionOverseas,
	VLLM:                  RegionOverseas,
	SGLang:                RegionOverseas,
	QuantumClaw:            RegionOverseas,

	// ── 国外量子算力 ──
	IonQ:         RegionOverseas,
	IBMQ:         RegionOverseas,
	Rigetti:      RegionOverseas,
	AWSBraket:    RegionOverseas,
	AzureQuantum: RegionOverseas,
	GoogleQuantum: RegionOverseas,
}

// ResolveChannelRegion 根据渠道类型判定区域
// 优先使用 channel 上已设置的 Region，其次按类型自动判定
func ResolveChannelRegion(channelType int, existingRegion string) string {
	if existingRegion != "" {
		return existingRegion
	}
	if r, ok := ChannelTypeRegion[channelType]; ok {
		return r
	}
	return RegionOverseas // 未识别的默认走国外
}

// ChannelTypeNames uses display-friendly names (e.g., "Google Gemini", "Ali (Qwen)")
// while model_metadata.Provider uses slug/tech names (e.g., "Google", "Alibaba").
// This mapping bridges the two so GetChannelTypes can auto-populate model lists.
var ChannelTypeNameToProvider = map[string]string{
	"Unknown":                 "",
	"OpenAI":                  "OpenAI",
	"API2D":                   "",
	"Azure":                   "",
	"CloseAI":                 "OpenAI",
	"OpenAI SB":               "OpenAI",
	"OpenAI Max":              "OpenAI",
	"OhMyGPT":                 "OpenAI",
	"Custom":                  "",
	"Ails":                    "",
	"AI Proxy":                "",
	"Google PaLM":             "Google",
	"API2GPT":                 "",
	"AIGC2D":                  "",
	"Anthropic":               "Anthropic",
	"Baidu":                   "Baidu",
	"Zhipu (GLM)":             "Zhipu",
	"Ali (Qwen)":              "Alibaba",
	"Xunfei (Spark)":          "Xunfei",
	"360 AI":                  "",
	"API Relay":               "",
	"AIProxy Library":         "",
	"FastGPT":                 "",
	"Tencent (Hunyuan)":       "Tencent",
	"Google Gemini":           "Google",
	"Moonshot (Kimi)":         "Moonshot",
	"Baichuan":                "Baichuan",
	"MiniMax":                 "MiniMax",
	"Mistral AI":              "Mistral",
	"Groq":                    "Groq",
	"Ollama":                  "Ollama",
	"LingYi (01.AI)":          "LingYi",
	"StepFun (Step-2)":        "StepFun",
	"AWS Claude":              "Anthropic",
	"Coze":                    "Coze",
	"Cohere":                  "Cohere",
	"DeepSeek":                "DeepSeek",
	"Cloudflare":              "Cloudflare",
	"DeepL":                   "DeepL",
	"Together AI":             "Together AI",
	"Doubao (VolcEngine)":     "Doubao",
	"Novita AI":               "Novita",
	"Google Vertex AI":        "Google",
	"Proxy":                   "",
	"SiliconFlow":             "SiliconFlow",
	"xAI (Grok)":              "xAI",
	"Replicate":               "Replicate",
	"Baidu V2":                "Baidu",
	"Xunfei V2":               "Xunfei",
	"Ali Bailian":             "Alibaba",
	"OpenAI Compatible":       "",
	"Gemini OpenAI Compatible": "Google",
	"Codex":                   "",
	"Dify":                    "",
	"Jimeng":                  "Jimeng",
	"Jina AI":                 "Jina",
	"Moka AI":                 "",
	"Submodel":                "",
	"VolcEngine (Ark)":        "VolcEngine",
	"Xinference":              "",
	"Zhipu V4":                "Zhipu",
	"IonQ":                    "IonQ",
	"IBM Q":                   "IBM",
	"Rigetti":                 "Rigetti",
	"AWS Braket":              "",
	"Azure Quantum":           "",
	"Google Quantum":          "Google",
}
