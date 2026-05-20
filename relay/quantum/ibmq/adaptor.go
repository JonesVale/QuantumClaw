package ibmq

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

// Adaptor — IBM Q 量子平台适配器
type Adaptor struct {
	APIKey  string
	BaseURL string
}

func (a *Adaptor) ProviderName() string { return "ibmq" }

// RunTask 向 IBM Q 提交量子任务
func (a *Adaptor) RunTask(ctx context.Context, req *quantum.QuantumTaskRequest) (*quantum.QuantumTaskResult, error) {
	// IBM Q 使用 QASM 格式
	qasm, err := CircuitToQASM(&req.Circuit)
	if err != nil {
		return nil, fmt.Errorf("ibmq: circuit to qasm: %w", err)
	}

	// 构建 IBM Q API 请求
	ibmReq := map[string]interface{}{
		"qasm": qasm,
		"backend": req.Backend,
		"shots": req.Shots,
	}

	body, _ := json.Marshal(ibmReq)
	apiURL := a.BaseURL + "/api/jobs"
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	httpReq.Header.Set("Authorization", "Bearer "+a.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ibmq: do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("ibmq: api error (status %d): %s", resp.StatusCode, string(respBody))
	}

	result, err := ParseIBMQResponse(respBody)
	if err != nil {
		return nil, fmt.Errorf("ibmq: parse response: %w", err)
	}
	result.Provider = "ibmq"
	return result, nil
}

func (a *Adaptor) QueryTask(ctx context.Context, taskID string) (*quantum.QuantumTaskResult, error) {
	apiURL := a.BaseURL + "/api/jobs/" + taskID
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	httpReq.Header.Set("Authorization", "Bearer "+a.APIKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ibmq: query request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	return ParseIBMQResponse(respBody)
}

func (a *Adaptor) CancelTask(ctx context.Context, taskID string) error {
	apiURL := a.BaseURL + "/api/jobs/" + taskID + "/cancel"
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPut, apiURL, nil)
	httpReq.Header.Set("Authorization", "Bearer "+a.APIKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("ibmq: cancel request: %w", err)
	}
	defer resp.Body.Close()
	return nil
}

func (a *Adaptor) ListBackends(ctx context.Context) ([]string, error) {
	return []string{"ibm_sherbrooke", "ibm_kyiv", "ibm_brisbane"}, nil
}

// ==================== 格式转换 ====================

// CircuitToQASM 将统一 Circuit 转换为 QASM 字符串
func CircuitToQASM(circuit *quantum.Circuit) (string, error) {
	if circuit == nil {
		return "", fmt.Errorf("nil circuit")
	}
	if circuit.QASM != "" {
		return circuit.QASM, nil
	}
	if len(circuit.Gates) == 0 {
		return "", fmt.Errorf("empty circuit: no gates or qasm provided")
	}

	qasm := fmt.Sprintf("OPENQASM 2.0;\ninclude \"qelib1.inc\";\nqreg q[%d];\ncreg c[%d];\n", circuit.Qubits, circuit.Qubits)

	gateNameMap := map[string]string{
		"h": "h", "x": "x", "y": "y", "z": "z",
		"cx": "cx", "cz": "cz", "ccx": "ccx",
		"rx": "rx", "ry": "ry", "rz": "rz",
		"t": "t", "s": "s", "sdg": "sdg", "tdg": "tdg",
	}

	for _, g := range circuit.Gates {
		gateName, ok := gateNameMap[g.Name]
		if !ok {
			gateName = g.Name
		}

		if len(g.Params) > 0 {
			// 参数量子门: rx(pi/2) q[0];
			qasm += fmt.Sprintf("%s(%v) q[%d];\n", gateName, g.Params[0], g.Targets[0])
			continue
		}
		if len(g.Control) > 0 && len(g.Targets) > 0 {
			// 受控门: cx q[0], q[1];
			qasm += fmt.Sprintf("%s q[%d], q[%d];\n", gateName, g.Control[0], g.Targets[0])
			continue
		}
		if len(g.Targets) > 0 {
			qasm += fmt.Sprintf("%s q[%d];\n", gateName, g.Targets[0])
		}
	}

	// 添加量子比特测量
	for i := 0; i < circuit.Qubits; i++ {
		qasm += fmt.Sprintf("measure q[%d] -> c[%d];\n", i, i)
	}

	return qasm, nil
}

// IBMQJobResponse — IBM Q 任务响应格式
type IBMQJobResponse struct {
	ID         string            `json:"id"`
	Status     string            `json:"status"`
	Result     map[string]int    `json:"result,omitempty"`
	Error      string            `json:"error,omitempty"`
	ExecTimeMs int64             `json:"execution_time_ms,omitempty"`
}

// ParseIBMQResponse 解析 IBM Q 响应为统一格式
func ParseIBMQResponse(data []byte) (*quantum.QuantumTaskResult, error) {
	var ibmResp IBMQJobResponse
	if err := json.Unmarshal(data, &ibmResp); err != nil {
		return nil, fmt.Errorf("unmarshal ibmq response: %w", err)
	}

	result := &quantum.QuantumTaskResult{
		TaskID:     ibmResp.ID,
		ExecTimeMs: ibmResp.ExecTimeMs,
		Backend:    "ibmq",
	}

	switch ibmResp.Status {
	case "COMPLETED":
		result.Status = "completed"
	case "RUNNING":
		result.Status = "running"
	case "QUEUED":
		result.Status = "queued"
	case "FAILED":
		result.Status = "failed"
		result.Error = ibmResp.Error
	default:
		result.Status = "queued"
	}

	if len(ibmResp.Result) > 0 {
		totalShots := 0
		for _, count := range ibmResp.Result {
			totalShots += count
		}
		result.Results = &quantum.TaskResults{
			Counts: ibmResp.Result,
			Shots:  totalShots,
		}
	}

	return result, nil
}
