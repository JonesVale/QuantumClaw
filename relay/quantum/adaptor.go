package quantum

import "context"

// QuantumAdaptor — 量子算力平台适配器接口
type QuantumAdaptor interface {
	RunTask(ctx context.Context, req *QuantumTaskRequest) (*QuantumTaskResult, error)
	QueryTask(ctx context.Context, taskID string) (*QuantumTaskResult, error)
	CancelTask(ctx context.Context, taskID string) error
	ListBackends(ctx context.Context) ([]string, error)
	ProviderName() string
}
