package novita

import (
	"fmt"

	"github.com/quantumclaw/quantumclaw/relay/meta"
	"github.com/quantumclaw/quantumclaw/relay/relaymode"
)

func GetRequestURL(meta *meta.Meta) (string, error) {
	switch meta.Mode {
	case relaymode.ChatCompletions:
		return fmt.Sprintf("%s/chat/completions", meta.BaseURL), nil
	case relaymode.Embeddings:
		return fmt.Sprintf("%s/embeddings", meta.BaseURL), nil
	case relaymode.ImagesGenerations:
		return fmt.Sprintf("%s/images/generations", meta.BaseURL), nil
	default:
	}
	// fallback: use relay request path
	return fmt.Sprintf("%s%s", meta.BaseURL, meta.RequestURLPath), nil
}
