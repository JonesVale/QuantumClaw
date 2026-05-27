package apitype

const (
	OpenAI = iota
	Anthropic
	PaLM
	Baidu
	Zhipu
	Ali
	Xunfei
	AIProxyLibrary
	Tencent
	Gemini
	Ollama
	AwsClaude
	Coze
	Cohere
	Cloudflare
	DeepL
	VertexAI
	Proxy
	Replicate
	Dify

	Dummy // this one is only for AI count, do not add AI channels after this

	// ==================== 量子算力 API 类型 ====================
	IONQ
	IBMQ
	RIGETTI
	AWS_BRAKET
	AZURE_QUANTUM
	GOOGLE_QUANTUM
	QUANTUM_DUMMY

	// ==================== 网页模型 API 类型 ====================
	Sub2API = iota + 200
)
