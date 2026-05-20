package relay

import (
	"fmt"

	"github.com/quantumclaw/quantumclaw/relay/channeltype"
	"github.com/quantumclaw/quantumclaw/relay/quantum"
	"github.com/quantumclaw/quantumclaw/relay/quantum/azure"
	"github.com/quantumclaw/quantumclaw/relay/quantum/braket"
	"github.com/quantumclaw/quantumclaw/relay/quantum/ibmq"
	"github.com/quantumclaw/quantumclaw/relay/quantum/ionq"
	"github.com/quantumclaw/quantumclaw/relay/quantum/rigetti"
)

// GetQuantumAdaptor 根据 channel type 返回对应的量子适配器
func GetQuantumAdaptor(channelType int) (quantum.QuantumAdaptor, error) {
	switch channelType {
	case channeltype.IonQ:
		return &ionq.Adaptor{}, nil
	case channeltype.IBMQ:
		return &ibmq.Adaptor{}, nil
	case channeltype.Rigetti:
		return &rigetti.Adaptor{}, nil
	case channeltype.AWSBraket:
		return &braket.Adaptor{}, nil
	case channeltype.AzureQuantum:
		return &azure.Adaptor{}, nil
	default:
		return nil, fmt.Errorf("unsupported quantum channel type: %d (implemented: IonQ, IBMQ, Rigetti, AWSBraket, AzureQuantum)", channelType)
	}
}
