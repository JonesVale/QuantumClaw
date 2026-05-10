package aiproxy

import "github.com/quantumclaw/quantumclaw/relay/adaptor/openai"

var ModelList = []string{""}

func init() {
	ModelList = openai.ModelList
}
