package ionq

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/quantumclaw/quantumclaw/relay/quantum"
)

// Adaptor — IonQ 量子平台适配器，实现 quantum.QuantumAdaptor 接口
type Adaptor struct {
	APIKey  string
	BaseURL string
}

// ProviderName 返回适配器名称
func (a *Adaptor) ProviderName() string {
	return "ionq"
}

// RunTask 向 IonQ 提交量子任务
func (a *Adaptor) RunTask(ctx context.Context, req *quantum.QuantumTaskRequest) (*quantum.QuantumTaskResult, error) {
	ionQReq, err := ConvertToIonQRequest(req)
	if err != nil {
		return nil, fmt.Errorf("ionq: convert request: %w", err)
	}

	body, err := json.Marshal(ionQReq)
	if err != nil {
		return nil, fmt.Errorf("ionq: marshal request: %w", err)
	}

	apiURL := a.BaseURL + "/v1/jobs"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ionq: create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+a.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ionq: do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ionq: read response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("ionq: api error (status %d): %s", resp.StatusCode, string(respBody))
	}

	result, err := ParseIonQResponse(respBody)
	if err != nil {
		return nil, fmt.Errorf("ionq: parse response: %w", err)
	}
	result.Provider = "ionq"
	return result, nil
}

// QueryTask 查询 IonQ 任务状态
func (a *Adaptor) QueryTask(ctx context.Context, taskID string) (*quantum.QuantumTaskResult, error) {
	apiURL := a.BaseURL + "/v1/jobs/" + taskID
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("ionq: create query: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+a.APIKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ionq: query request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ionq: read query response: %w", err)
	}

	result, err := ParseIonQResponse(respBody)
	if err != nil {
		return nil, fmt.Errorf("ionq: parse query response: %w", err)
	}
	result.Provider = "ionq"
	return result, nil
}

// CancelTask 取消 IonQ 任务
func (a *Adaptor) CancelTask(ctx context.Context, taskID string) error {
	apiURL := a.BaseURL + "/v1/jobs/" + taskID + "/cancel"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, apiURL, nil)
	if err != nil {
		return fmt.Errorf("ionq: create cancel request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+a.APIKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("ionq: cancel request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ionq: cancel error (status %d): %s", resp.StatusCode, string(body))
	}
	return nil
}

// ListBackends 列出 IonQ 可用后端
func (a *Adaptor) ListBackends(ctx context.Context) ([]string, error) {
	return []string{"ionq_harmony", "ionq_aria"}, nil
}
