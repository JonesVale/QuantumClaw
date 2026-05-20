package ionq

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/quantumclaw/quantumclaw/relay/quantum"
)

// IonQJobRequest — IonQ API 的任务请求格式
type IonQJobRequest struct {
	Target    string      `json:"target"`
	Circuit   interface{} `json:"circuit"`
	Shots     int         `json:"shots"`
	Name      string      `json:"name,omitempty"`
}

// IonQJobResponse — IonQ API 的任务响应格式
type IonQJobResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Name   string `json:"name"`
	Circuit   interface{}   `json:"circuit,omitempty"`
	Result    *IonQResult   `json:"result,omitempty"`
	Failure   *IonQFailure  `json:"failure,omitempty"`
	QueuePosition int       `json:"queue_position,omitempty"`
	Cost     *IonQCost   `json:"cost,omitempty"`
}

// IonQResult — IonQ 计算结果
type IonQResult struct {
	Histogram map[string]float64 `json:"histogram"`
}

// IonQFailure — IonQ 任务失败信息
type IonQFailure struct {
	Code    string `json:"code"`
	Message string `json:"error"`
}

// IonQCost — IonQ 费用信息
type IonQCost struct {
	AlgorithmFamily string  `json:"algorithm_family"`
	ExecTimeMs      int64   `json:"exec_time_ms"`
	Cost            float64 `json:"cost"`
}

// ==================== 格式转换 ====================

// ConvertToIonQRequest 将统一 QuantumTaskRequest 转换为 IonQ API 格式
func ConvertToIonQRequest(req *quantum.QuantumTaskRequest) (*IonQJobRequest, error) {
	target := req.Backend
	if target == "" || target == "auto" {
		target = "qpu.harmony" // 默认后端
	}

	shots := req.Shots
	if shots <= 0 {
		shots = 1000
	}

	// 构建 IonQ 电路格式
	circuit := buildIonQCircuit(req)
	if circuit == nil {
		return nil, fmt.Errorf("empty circuit: no gates or qasm provided")
	}

	return &IonQJobRequest{
		Target:  target,
		Circuit: circuit,
		Shots:   shots,
	}, nil
}

// buildIonQCircuit 将统一 Circuit 转换为 IonQ 电路格式
func buildIonQCircuit(req *quantum.QuantumTaskRequest) interface{} {
	circuit := req.Circuit

	// 优先使用 QASM
	if circuit.QASM != "" {
		return map[string]interface{}{
			"gateSet": "qis",
			"qubits":  circuit.Qubits,
			"circuit": circuit.QASM,
		}
	}

	// 使用结构化门序列
	if len(circuit.Gates) == 0 {
		return nil
	}

	var gates []map[string]interface{}
	for _, g := range circuit.Gates {
		gate := map[string]interface{}{
			"gate":    g.Name,
			"targets": g.Targets,
		}
		if len(g.Control) > 0 {
			gate["controls"] = g.Control
		}
		if len(g.Params) > 0 {
			gate["params"] = g.Params
		}
		gates = append(gates, gate)
	}

	return map[string]interface{}{
		"gateSet": "qis",
		"qubits":  circuit.Qubits,
		"gates":   gates,
	}
}

// ParseIonQResponse 解析 IonQ API 响应为统一格式
func ParseIonQResponse(data []byte) (*quantum.QuantumTaskResult, error) {
	var ionQResp IonQJobResponse
	if err := json.Unmarshal(data, &ionQResp); err != nil {
		return nil, fmt.Errorf("unmarshal ionq response: %w", err)
	}

	result := &quantum.QuantumTaskResult{
		TaskID: ionQResp.ID,
		Backend: ionQResp.Name,
	}

	// 映射状态
	switch ionQResp.Status {
	case "ready", "running":
		result.Status = "running"
	case "completed", "finished":
		result.Status = "completed"
	case "failed":
		result.Status = "failed"
		if ionQResp.Failure != nil {
			result.Error = ionQResp.Failure.Message
		}
	default:
		result.Status = "queued"
	}

	// 解析计算结果
	if ionQResp.Result != nil && len(ionQResp.Result.Histogram) > 0 {
		counts := make(map[string]int)
		probs := make(map[string]float64)
		totalShots := 0

		// IonQ 返回概率分布，需要映射到 counts
		for state, prob := range ionQResp.Result.Histogram {
			probs[state] = prob
			counts[state] = int(math.Round(prob * 1000)) // 估算 counts
			totalShots += counts[state]
		}

		result.Results = &quantum.TaskResults{
			Counts:        counts,
			Probabilities: probs,
			Shots:         totalShots,
		}
	}

	// 费用信息
	if ionQResp.Cost != nil {
		result.ExecTimeMs = ionQResp.Cost.ExecTimeMs
		result.CostQuota = int64(ionQResp.Cost.Cost * 1000000) // 转换为微额度
	}

	return result, nil
}
