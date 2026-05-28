package openai

import (
	"github.com/quantumclaw/quantumclaw/relay/adaptor/ai360"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/alibailian"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/baichuan"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/baiduv2"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/deepseek"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/doubao"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/geminiv2"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/groq"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/lingyiwanwu"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/minimax"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/mistral"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/moonshot"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/novita"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/quantumclaw"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/siliconflow"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/stepfun"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/togetherai"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/xai"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/xunfeiv2"
	"github.com/quantumclaw/quantumclaw/relay/channeltype"
)

// OpenAICompatibleChannels 扩展列表（包含 OpenAICompatible 通用渠道类型）
var OpenAICompatibleChannels = []int{
	channeltype.OpenAICompatible,
	channeltype.GeminiOpenAICompatible,
	channeltype.QuantumClaw,
	channeltype.AliBailian,
	channeltype.BaiduV2,
}

var CompatibleChannels = []int{
	channeltype.Azure,
	channeltype.AI360,
	channeltype.Moonshot,
	channeltype.Baichuan,
	channeltype.Minimax,
	channeltype.Doubao,
	channeltype.Mistral,
	channeltype.Groq,
	channeltype.LingYiWanWu,
	channeltype.StepFun,
	channeltype.DeepSeek,
	channeltype.TogetherAI,
	channeltype.Novita,
	channeltype.SiliconFlow,
	channeltype.XAI,
	channeltype.BaiduV2,
	channeltype.XunfeiV2,
}

func GetCompatibleChannelMeta(channelType int) (string, []string) {
	switch channelType {
	case channeltype.Azure:
		return "azure", ModelList
	case channeltype.AI360:
		return "360", ai360.ModelList
	case channeltype.Moonshot:
		return "moonshot", moonshot.ModelList
	case channeltype.Baichuan:
		return "baichuan", baichuan.ModelList
	case channeltype.Minimax:
		return "minimax", minimax.ModelList
	case channeltype.Mistral:
		return "mistralai", mistral.ModelList
	case channeltype.Groq:
		return "groq", groq.ModelList
	case channeltype.LingYiWanWu:
		return "lingyiwanwu", lingyiwanwu.ModelList
	case channeltype.StepFun:
		return "stepfun", stepfun.ModelList
	case channeltype.DeepSeek:
		return "deepseek", deepseek.ModelList
	case channeltype.TogetherAI:
		return "together.ai", togetherai.ModelList
	case channeltype.Doubao:
		return "doubao", doubao.ModelList
	case channeltype.Novita:
		return "novita", novita.ModelList
	case channeltype.SiliconFlow:
		return "siliconflow", siliconflow.ModelList
	case channeltype.XAI:
		return "xai", xai.ModelList
	case channeltype.BaiduV2:
		return "baiduv2", baiduv2.ModelList
	case channeltype.XunfeiV2:
		return "xunfeiv2", xunfeiv2.ModelList
	case channeltype.QuantumClaw:
		return "quantumclaw", quantumclaw.ModelList
	case channeltype.AliBailian:
		return "alibailian", alibailian.ModelList
	case channeltype.GeminiOpenAICompatible:
		return "geminiv2", geminiv2.ModelList
	default:
		return "openai", ModelList
	}
}
