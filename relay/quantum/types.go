package quantum

// QuantumTaskRequest — 统一量子任务请求入参
type QuantumTaskRequest struct {
	Provider string `json:"provider,omitempty"` // "auto" | "ionq" | "ibmq"
	Backend  string `json:"backend,omitempty"`  // "auto" | "ionq_harmony" | "ibm_sherbrooke"
	Circuit  Circuit `json:"circuit,omitempty"`
	Shots    int    `json:"shots"`
	OptimizationLevel int `json:"optimization_level,omitempty"`
}

// Circuit — 统一量子电路描述
type Circuit struct {
	Qubits int    `json:"qubits"`
	Gates  []Gate `json:"gates,omitempty"`  // 结构化门序列
	QASM   string `json:"qasm,omitempty"`   // 原始 QASM 字符串（替代 Gates）
}

// Gate — 量子门操作
type Gate struct {
	Name    string    `json:"name"`
	Targets []int     `json:"targets"`
	Control []int     `json:"controls,omitempty"`
	Params  []float64 `json:"params,omitempty"`
}

// QuantumTaskResult — 统一量子任务响应出参
type QuantumTaskResult struct {
	TaskID     string       `json:"task_id"`
	Status     string       `json:"status"`               // "queued" | "running" | "completed" | "failed"
	Provider   string       `json:"provider"`
	Backend    string       `json:"backend"`
	Results    *TaskResults `json:"results,omitempty"`
	Error      string       `json:"error,omitempty"`
	ExecTimeMs int64        `json:"execution_time_ms"`
	CostQuota  int64        `json:"cost_quota"`
}

// TaskResults — 量子计算结果
type TaskResults struct {
	Counts        map[string]int     `json:"counts"`
	Probabilities map[string]float64 `json:"probabilities,omitempty"`
	Shots         int                `json:"shots"`
}
