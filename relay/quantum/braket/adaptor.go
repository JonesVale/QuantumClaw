package braket

import (
	"context"
	"fmt"

	"github.com/quantumclaw/quantumclaw/relay/quantum"
)

// Adaptor — AWS Braket 量子平台适配器（桩实现）
type Adaptor struct {
	APIKey  string
	BaseURL string
}

func (a *Adaptor) ProviderName() string   { return "aws_braket" }
func (a *Adaptor) RunTask(ctx context.Context, req *quantum.QuantumTaskRequest) (*quantum.QuantumTaskResult, error) {
	return nil, fmt.Errorf("aws_braket: not yet implemented")
}
func (a *Adaptor) QueryTask(ctx context.Context, taskID string) (*quantum.QuantumTaskResult, error) {
	return nil, fmt.Errorf("aws_braket: not yet implemented")
}
func (a *Adaptor) CancelTask(ctx context.Context, taskID string) error {
	return fmt.Errorf("aws_braket: not yet implemented")
}
func (a *Adaptor) ListBackends(ctx context.Context) ([]string, error) {
	return []string{"braket_sv1", "braket_dm1", "braket_rigetti"}, nil
}
