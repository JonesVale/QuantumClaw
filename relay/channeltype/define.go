package channeltype

const (
	Unknown = iota
	OpenAI
	API2D
	Azure
	CloseAI
	OpenAISB
	OpenAIMax
	OhMyGPT
	Custom
	Ails
	AIProxy
	PaLM
	API2GPT
	AIGC2D
	Anthropic
	Baidu
	Zhipu
	Ali
	Xunfei
	AI360
	QuantumClaw
	AIProxyLibrary
	FastGPT
	Tencent
	Gemini
	Moonshot
	Baichuan
	Minimax
	Mistral
	Groq
	Ollama
	LingYiWanWu
	StepFun
	AwsClaude
	Coze
	Cohere
	DeepSeek
	Cloudflare
	DeepL
	TogetherAI
	Doubao
	Novita
	VertextAI
	Proxy
	SiliconFlow
	XAI
	Replicate
	BaiduV2
	XunfeiV2
	AliBailian
	OpenAICompatible
	GeminiOpenAICompatible
	Codex
	Dify
	Jimeng
	Jina
	MokaAI
	Submodel
	VolcEngine
	Xinference
	ZhipuV4
	Dummy
)

// ==================== 量子算力渠道 (type >= 100) ====================
const (
	IonQ         = 100
	IBMQ         = 101
	Rigetti      = 102
	AWSBraket    = 103
	AzureQuantum = 104
	GoogleQuantum = 105
	QuantumDummy = 106 // 用于 ChannelBaseURLs 长度检查
)
