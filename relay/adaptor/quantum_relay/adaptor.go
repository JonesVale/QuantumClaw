package quantumrelay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/openai"
	"github.com/quantumclaw/quantumclaw/relay/meta"
	"github.com/quantumclaw/quantumclaw/relay/model"
	"github.com/quantumclaw/quantumclaw/relay/quantum"
	"github.com/quantumclaw/quantumclaw/relay/quantum/ionq"
	"github.com/quantumclaw/quantumclaw/relay/quantum/ibmq"
	"github.com/quantumclaw/quantumclaw/relay/quantum/rigetti"
	"github.com/quantumclaw/quantumclaw/relay/quantum/azure"
	"github.com/quantumclaw/quantumclaw/relay/quantum/braket"
)

// Adaptor — 量子算力 relay.Adaptor 桥接
// 将 quantum.QuantumAdaptor（异步任务模式）包装为 relay 标准 Adaptor（同步请求-响应模式）
type Adaptor struct {
	meta        *meta.Meta
	provider    string
	quantumImpl quantum.QuantumAdaptor
}

func (a *Adaptor) Init(meta *meta.Meta) {
	a.meta = meta
}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	return meta.BaseURL, nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	req.Header.Set("Authorization", "Bearer "+meta.APIKey)
	req.Header.Set("Content-Type", "application/json")
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	return request, nil
}

func (a *Adaptor) ConvertImageRequest(request *model.ImageRequest) (any, error) {
	return nil, fmt.Errorf("quantum: image requests not supported")
}

// DoRequest 将 OpenAI ChatCompletion 请求转换为量子任务并执行
// 同步模式：提交任务 → 轮询直到完成 → 返回结果作为 HTTP 响应
func (a *Adaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	// 1. 解析请求
	body, err := io.ReadAll(requestBody)
	if err != nil {
		return nil, fmt.Errorf("quantum: read request body: %w", err)
	}

	var req model.GeneralOpenAIRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("quantum: unmarshal request: %w", err)
	}

	// 2. 从请求中提取量子电路参数
	qReq := buildQuantumRequest(&req)

	// 3. 创建对应 provider 的适配器
	adaptorImpl, err := newQuantumProvider(meta)
	if err != nil {
		return nil, fmt.Errorf("quantum: create provider: %w", err)
	}
	a.quantumImpl = adaptorImpl
	a.provider = adaptorImpl.ProviderName()

	// 4. 提交任务
	result, err := adaptorImpl.RunTask(c.Request.Context(), qReq)
	if err != nil {
		return nil, fmt.Errorf("quantum: run task: %w", err)
	}

	// 5. 如果是同步模式，轮询直到完成
	usage := &model.Usage{
		PromptTokens:     estimateTokens(&req),
		CompletionTokens: 0,
		TotalTokens:      0,
	}

	if !req.Stream {
		// 非流式：轮询直到完成
		result = pollUntilComplete(c, adaptorImpl, result.TaskID, 60*time.Second)
		usage.CompletionTokens = int(result.ExecTimeMs / 10) // 近似
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}

	// 6. 构建 OpenAI 兼容响应
	openaiResp := buildOpenAIResponse(req.Model, result, usage)

	respBytes, _ := json.Marshal(openaiResp)
	logger.Debugf(c.Request.Context(), "quantum: provider=%s task=%s status=%s",
		a.provider, result.TaskID, result.Status)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body:          io.NopCloser(bytes.NewReader(respBytes)),
		ContentLength: int64(len(respBytes)),
	}, nil
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (*model.Usage, *model.ErrorWithStatusCode) {
	// 直接传递响应，使用 openai adaptor 的 DoResponse
	oa := &openai.Adaptor{}
	return oa.DoResponse(c, resp, meta)
}

func (a *Adaptor) GetModelList() []string {
	return []string{
		"ionq_harmony", "ionq_aria",
		"ibm_sherbrooke", "ibm_kyiv", "ibm_brisbane",
		"rigetti_aspem", "rigetti_ankaa",
		"azure_quantinium",
		"braket_sv1", "braket_dm1",
	}
}

func (a *Adaptor) GetChannelName() string {
	if a.provider != "" {
		return "quantum_" + a.provider
	}
	return "quantum"
}

// ==================== 工具函数 ====================

func newQuantumProvider(meta *meta.Meta) (quantum.QuantumAdaptor, error) {
	apiKey := meta.APIKey
	baseURL := meta.BaseURL

	switch meta.APIType {
	case 100: // apitype.IONQ
		return &ionq.Adaptor{APIKey: apiKey, BaseURL: baseURL}, nil
	case 101: // apitype.IBMQ
		return &ibmq.Adaptor{APIKey: apiKey, BaseURL: baseURL}, nil
	case 102: // apitype.RIGETTI
		return &rigetti.Adaptor{APIKey: apiKey, BaseURL: baseURL}, nil
	case 104: // apitype.AZURE_QUANTUM
		return &azure.Adaptor{APIKey: apiKey, BaseURL: baseURL}, nil
	case 103: // apitype.AWS_BRAKET
		return &braket.Adaptor{APIKey: apiKey, BaseURL: baseURL}, nil
	default:
		return nil, fmt.Errorf("unknown quantum api type: %d", meta.APIType)
	}
}

func buildQuantumRequest(req *model.GeneralOpenAIRequest) *quantum.QuantumTaskRequest {
	qReq := &quantum.QuantumTaskRequest{
		Shots:             1024,
		OptimizationLevel: 1,
	}

	// 从模型名提取 backend
	if strings.Contains(req.Model, "ibm") || strings.Contains(req.Model, "ibmq") {
		qReq.Backend = req.Model
		qReq.Provider = "ibmq"
	} else if strings.Contains(req.Model, "ionq") {
		qReq.Backend = req.Model
		qReq.Provider = "ionq"
	} else if strings.Contains(req.Model, "rigetti") {
		qReq.Backend = req.Model
		qReq.Provider = "rigetti"
	} else if strings.Contains(req.Model, "azure") {
		qReq.Backend = req.Model
		qReq.Provider = "azure"
	} else if strings.Contains(req.Model, "braket") {
		qReq.Backend = req.Model
		qReq.Provider = "braket"
	}

	// 从用户消息提取 QASM（如果存在）
	for _, msg := range req.Messages {
		if content, ok := msg.Content.(string); ok {
			// 检查是否包含 QASM 代码块
			if strings.Contains(content, "OPENQASM") || strings.Contains(content, "qreg") {
				extracted := extractQASM(content)
				if extracted != "" {
					qReq.Circuit.QASM = extracted
					qReq.Circuit.Qubits = countQubits(extracted)
				}
			}
		}
	}

	// 如果没有从消息中提取到 QASM，使用默认 Bell 态电路
	if qReq.Circuit.QASM == "" {
		qReq.Circuit.QASM = fmt.Sprintf(`OPENQASM 2.0;
include "qelib1.inc";
qreg q[2];
creg c[2];
h q[0];
cx q[0], q[1];
measure q[0] -> c[0];
measure q[1] -> c[1];`)
		qReq.Circuit.Qubits = 2
	}

	return qReq
}

func extractQASM(content string) string {
	// 提取 ```qasm ... ``` 或 ``` ... ``` 或裸 QASM
	start := strings.Index(content, "OPENQASM")
	if start < 0 {
		start = strings.Index(content, "qreg")
	}
	if start < 0 {
		return ""
	}
	end := strings.LastIndex(content, "c[") // 找到测量指令作为结束标记
	if end < 0 {
		end = len(content)
	} else {
		end = strings.Index(content[end:], ";") + end + 1
	}
	if end > len(content) {
		end = len(content)
	}
	return content[start:end]
}

func countQubits(qasm string) int {
	// qreg q[N] 或 q[N]
	idx := strings.Index(qasm, "qreg")
	if idx < 0 {
		return 1
	}
	rest := qasm[idx:]
	start := strings.Index(rest, "[")
	if start < 0 {
		return 1
	}
	end := strings.Index(rest[start:], "]")
	if end < 0 {
		return 1
	}
	var n int
	fmt.Sscanf(rest[start+1:start+end], "%d", &n)
	if n <= 0 {
		return 1
	}
	return n
}

func pollUntilComplete(c *gin.Context, impl quantum.QuantumAdaptor, taskID string, timeout time.Duration) *quantum.QuantumTaskResult {
	deadline := time.Now().Add(timeout)
	pollInterval := 2 * time.Second

	for time.Now().Before(deadline) {
		result, err := impl.QueryTask(c.Request.Context(), taskID)
		if err == nil && (result.Status == "completed" || result.Status == "failed" || result.Status == "cancelled") {
			return result
		}
		time.Sleep(pollInterval)
	}

	return &quantum.QuantumTaskResult{
		TaskID: taskID,
		Status: "timeout",
		Error:  fmt.Sprintf("task did not complete within %v", timeout),
	}
}

func estimateTokens(req *model.GeneralOpenAIRequest) int {
	tokens := 0
	for _, msg := range req.Messages {
		if content, ok := msg.Content.(string); ok {
			tokens += len(content) / 4 // 简单估计
		}
	}
	if tokens == 0 {
		tokens = 100
	}
	return tokens
}

// buildOpenAIResponse 将量子结果包装为 OpenAI ChatCompletion 响应格式
func buildOpenAIResponse(modelName string, result *quantum.QuantumTaskResult, usage *model.Usage) map[string]any {
	// 将量子计算结果格式化为文本
	content := formatQuantumResult(result)

	resp := map[string]any{
		"id":      "q-" + result.TaskID,
		"object":  "chat.completion",
		"created": time.Now().Unix(),
		"model":   modelName,
		"choices": []map[string]any{
			{
				"index": 0,
				"message": map[string]any{
					"role":    "assistant",
					"content": content,
				},
				"finish_reason": "stop",
			},
		},
		"usage": map[string]any{
			"prompt_tokens":     usage.PromptTokens,
			"completion_tokens": usage.CompletionTokens,
			"total_tokens":      usage.TotalTokens,
		},
	}

	// 添加量子特有元数据
	quantumMeta := map[string]any{
		"provider":  result.Provider,
		"backend":   result.Backend,
		"task_id":   result.TaskID,
		"status":    result.Status,
		"exec_time_ms": result.ExecTimeMs,
	}
	if result.Results != nil {
		quantumMeta["results"] = map[string]any{
			"counts":        result.Results.Counts,
			"probabilities": result.Results.Probabilities,
			"shots":         result.Results.Shots,
		}
	}
	if result.Error != "" {
		quantumMeta["error"] = result.Error
	}

	resp["quantum_metadata"] = quantumMeta

	return resp
}

func formatQuantumResult(result *quantum.QuantumTaskResult) string {
	if result.Status == "failed" || result.Status == "timeout" {
		if result.Error != "" {
			return fmt.Sprintf("量子计算任务失败: %s", result.Error)
		}
		return fmt.Sprintf("量子计算任务状态: %s", result.Status)
	}

	if result.Results != nil {
		// 格式化结果：概率分布
		text := fmt.Sprintf("量子计算完成 (provider: %s, backend: %s, 耗时: %dms)\n\n", 
			result.Provider, result.Backend, result.ExecTimeMs)
		text += "=== 测量结果分布 ===\n"
		for state, count := range result.Results.Counts {
			total := result.Results.Shots
			if total == 0 {
				total = 1
			}
			pct := float64(count) / float64(total) * 100
			text += fmt.Sprintf("|%s⟩: %d 次 (%.1f%%)\n", state, count, pct)
		}
		text += fmt.Sprintf("\n总采样次数: %d", result.Results.Shots)
		return text
	}

	return fmt.Sprintf("量子任务已提交 (task: %s, status: %s)", result.TaskID, result.Status)
}
