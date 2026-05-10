package minimax

import (
	"fmt"
	"github.com/quantumclaw/quantumclaw/relay/meta"
	"github.com/quantumclaw/quantumclaw/relay/relaymode"
)

func GetRequestURL(meta *meta.Meta) (string, error) {
	switch meta.Mode {
	case relaymode.ChatCompletions:
		return fmt.Sprintf("%s/v1/text/chatcompletion_v2", meta.BaseURL), nil
	case relaymode.Embeddings:
		return fmt.Sprintf("%s/v1/embeddings", meta.BaseURL), nil
	default:
	}
	// fallback: use relay request path
	return fmt.Sprintf("%s%s", meta.BaseURL, meta.RequestURLPath), nil
}
