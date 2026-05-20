package relay

import (
	"testing"

	"github.com/quantumclaw/quantumclaw/relay/channeltype"
	"github.com/quantumclaw/quantumclaw/relay/quantum"
	"github.com/quantumclaw/quantumclaw/relay/quantum/azure"
	"github.com/quantumclaw/quantumclaw/relay/quantum/braket"
	"github.com/quantumclaw/quantumclaw/relay/quantum/ibmq"
	"github.com/quantumclaw/quantumclaw/relay/quantum/ionq"
	"github.com/quantumclaw/quantumclaw/relay/quantum/rigetti"
)

func TestGetQuantumAdaptor_KnownTypes(t *testing.T) {
	tests := []struct {
		name        string
		channelType int
		wantType    string
	}{
		{"IonQ", channeltype.IonQ, "*ionq.Adaptor"},
		{"IBMQ", channeltype.IBMQ, "*ibmq.Adaptor"},
		{"Rigetti", channeltype.Rigetti, "*rigetti.Adaptor"},
		{"AWSBraket", channeltype.AWSBraket, "*braket.Adaptor"},
		{"AzureQuantum", channeltype.AzureQuantum, "*azure.Adaptor"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adaptor, err := GetQuantumAdaptor(tt.channelType)
			if err != nil {
				t.Fatalf("GetQuantumAdaptor(%d) returned error: %v", tt.channelType, err)
			}
			if adaptor == nil {
				t.Fatal("GetQuantumAdaptor returned nil")
			}

			// Verify correct adaptor type
			switch adaptor.(type) {
			case *ionq.Adaptor:
				if tt.channelType != channeltype.IonQ {
					t.Errorf("got IonQ adaptor for channel type %d", tt.channelType)
				}
			case *ibmq.Adaptor:
				if tt.channelType != channeltype.IBMQ {
					t.Errorf("got IBMQ adaptor for channel type %d", tt.channelType)
				}
			case *rigetti.Adaptor:
				if tt.channelType != channeltype.Rigetti {
					t.Errorf("got Rigetti adaptor for channel type %d", tt.channelType)
				}
			case *braket.Adaptor:
				if tt.channelType != channeltype.AWSBraket {
					t.Errorf("got Braket adaptor for channel type %d", tt.channelType)
				}
			case *azure.Adaptor:
				if tt.channelType != channeltype.AzureQuantum {
					t.Errorf("got Azure adaptor for channel type %d", tt.channelType)
				}
			default:
				t.Errorf("unexpected adaptor type: %T", adaptor)
			}
		})
	}
}

func TestGetQuantumAdaptor_UnknownType(t *testing.T) {
	// Unknown quantum channel type should return error
	_, err := GetQuantumAdaptor(999)
	if err == nil {
		t.Error("GetQuantumAdaptor(999) should return error for unknown type")
	}
}

func TestGetQuantumAdaptor_AIChannelReturnsError(t *testing.T) {
	// AI channel types (0-61) should return error for GetQuantumAdaptor
	_, err := GetQuantumAdaptor(channeltype.OpenAI)
	if err == nil {
		t.Error("GetQuantumAdaptor for OpenAI channel should return error")
	}
}

func TestGetQuantumAdaptor_AllImplementQuantumAdaptor(t *testing.T) {
	types := []int{
		channeltype.IonQ,
		channeltype.IBMQ,
		channeltype.Rigetti,
		channeltype.AWSBraket,
		channeltype.AzureQuantum,
	}

	for _, ct := range types {
		adaptor, err := GetQuantumAdaptor(ct)
		if err != nil {
			t.Errorf("GetQuantumAdaptor(%d) failed: %v", ct, err)
			continue
		}
		// All adaptors must implement the QuantumAdaptor interface
		var _ quantum.QuantumAdaptor = adaptor
	}

	_ = types
}

func TestGetQuantumAdaptor_ProviderNameNotEmpty(t *testing.T) {
	types := []int{
		channeltype.IonQ,
		channeltype.IBMQ,
		channeltype.Rigetti,
		channeltype.AWSBraket,
		channeltype.AzureQuantum,
	}

	for _, ct := range types {
		adaptor, err := GetQuantumAdaptor(ct)
		if err != nil {
			continue
		}
		if adaptor.ProviderName() == "" {
			t.Errorf("ProviderName() for channel type %d should not be empty", ct)
		}
	}
}

func TestGetQuantumAdaptor_ListBackendsNotEmpty(t *testing.T) {
	types := []int{
		channeltype.IonQ,
		channeltype.IBMQ,
		channeltype.Rigetti,
		channeltype.AWSBraket,
		channeltype.AzureQuantum,
	}

	for _, ct := range types {
		adaptor, err := GetQuantumAdaptor(ct)
		if err != nil {
			continue
		}
		backends, err := adaptor.ListBackends(nil)
		if err != nil {
			t.Errorf("ListBackends() for channel type %d failed: %v", ct, err)
			continue
		}
		if len(backends) == 0 {
			t.Errorf("ListBackends() for channel type %d returned empty list", ct)
		}
	}
}
