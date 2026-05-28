package channeltype

import "github.com/quantumclaw/quantumclaw/relay/apitype"

func ToAPIType(channelType int) int {
	apiType := apitype.OpenAI
	switch channelType {
	case Anthropic:
		apiType = apitype.Anthropic
	case Baidu:
		apiType = apitype.Baidu
	case PaLM:
		apiType = apitype.PaLM
	case Zhipu:
		apiType = apitype.Zhipu
	case Ali:
		apiType = apitype.Ali
	case Xunfei:
		apiType = apitype.Xunfei
	case AIProxyLibrary:
		apiType = apitype.AIProxyLibrary
	case Tencent:
		apiType = apitype.Tencent
	case Gemini:
		apiType = apitype.Gemini
	case Ollama:
		apiType = apitype.Ollama
	case AwsClaude:
		apiType = apitype.AwsClaude
	case Coze:
		apiType = apitype.Coze
	case Cohere:
		apiType = apitype.Cohere
	case Cloudflare:
		apiType = apitype.Cloudflare
	case DeepL:
		apiType = apitype.DeepL
	case VertextAI:
		apiType = apitype.VertexAI
	case Replicate:
		apiType = apitype.Replicate
	case Proxy:
		apiType = apitype.Proxy
	// ==================== QuantumClaw 新增：OpenAI-Compatible 渠道 ====================
	// 以下渠道均为 OpenAI 兼容协议，共用 openai.Adaptor
	case DeepSeek, Doubao, Minimax, Groq, Mistral, Novita,
		Baichuan, Moonshot, LingYiWanWu, StepFun,
		TogetherAI, SiliconFlow, XAI,
		BaiduV2, XunfeiV2, AliBailian,
		QuantumClaw, GeminiOpenAICompatible, OpenAICompatible,
		// ==================== Phase 2: 新增缺失 Provider ====================
		Codex, Jimeng, Jina, MokaAI,
		Submodel, VolcEngine, Xinference, ZhipuV4:
		apiType = apitype.OpenAI

	case Dify:
		apiType = apitype.Dify

	case Sub2API:
		apiType = apitype.Sub2API

	case VLLM, SGLang:
		apiType = apitype.OpenAI

	// ==================== 量子算力渠道映射 ====================
	case IonQ:
		apiType = apitype.IONQ
	case IBMQ:
		apiType = apitype.IBMQ
	case Rigetti:
		apiType = apitype.RIGETTI
	case AWSBraket:
		apiType = apitype.AWS_BRAKET
	case AzureQuantum:
		apiType = apitype.AZURE_QUANTUM
	case GoogleQuantum:
		apiType = apitype.GOOGLE_QUANTUM
	}

	return apiType
}

// IsDomesticModel 判断渠道类型是否为国内模型
// 国内: Baidu, Ali, Zhipu, Xunfei, Tencent, DeepSeek, Moonshot, Baichuan, Minimax,
//       LingYiWanWu, StepFun, Doubao, SiliconFlow, MokaAI, VolcEngine, ZhipuV4,
//       Jimeng, Xinference, Dify, 360, Submodel
// 国外/其他: 剩余所有（OpenAI, Anthropic, Gemini, AWS, Groq, Mistral, Cohere 等）
func IsDomesticModel(channelType int) bool {
	switch channelType {
	case Baidu, Zhipu, Ali, Xunfei, Tencent, DeepSeek,
		Moonshot, Baichuan, Minimax, LingYiWanWu, StepFun,
		Doubao, SiliconFlow, BaiduV2, XunfeiV2, AliBailian,
		ZhipuV4, MokaAI, VolcEngine, Xinference,
		AI360, Jimeng, Submodel, Dify:
		return true
	default:
		return false
	}
}
