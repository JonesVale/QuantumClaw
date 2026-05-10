package relay

import (
	"github.com/quantumclaw/quantumclaw/relay/adaptor"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/aiproxy"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/ali"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/anthropic"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/aws"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/baidu"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/cloudflare"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/cohere"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/coze"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/deepl"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/gemini"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/ollama"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/openai"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/palm"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/proxy"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/replicate"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/tencent"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/vertexai"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/xunfei"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/zhipu"
	"github.com/quantumclaw/quantumclaw/relay/apitype"
)

func GetAdaptor(apiType int) adaptor.Adaptor {
	switch apiType {
	case apitype.AIProxyLibrary:
		return &aiproxy.Adaptor{}
	case apitype.Ali:
		return &ali.Adaptor{}
	case apitype.Anthropic:
		return &anthropic.Adaptor{}
	case apitype.AwsClaude:
		return &aws.Adaptor{}
	case apitype.Baidu:
		return &baidu.Adaptor{}
	case apitype.Gemini:
		return &gemini.Adaptor{}
	case apitype.OpenAI:
		return &openai.Adaptor{}
	case apitype.PaLM:
		return &palm.Adaptor{}
	case apitype.Tencent:
		return &tencent.Adaptor{}
	case apitype.Xunfei:
		return &xunfei.Adaptor{}
	case apitype.Zhipu:
		return &zhipu.Adaptor{}
	case apitype.Ollama:
		return &ollama.Adaptor{}
	case apitype.Coze:
		return &coze.Adaptor{}
	case apitype.Cohere:
		return &cohere.Adaptor{}
	case apitype.Cloudflare:
		return &cloudflare.Adaptor{}
	case apitype.DeepL:
		return &deepl.Adaptor{}
	case apitype.VertexAI:
		return &vertexai.Adaptor{}
	case apitype.Proxy:
		return &proxy.Adaptor{}
	case apitype.Replicate:
		return &replicate.Adaptor{}
	}
	return nil
}
