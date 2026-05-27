package vllm

import (
	"github.com/quantumclaw/quantumclaw/relay/adaptor"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/openai"
)

type Adaptor struct {
	openai.Adaptor
}

func (a *Adaptor) GetModelList() []string {
	return []string{}
}

var _ adaptor.Adaptor = (*Adaptor)(nil)
