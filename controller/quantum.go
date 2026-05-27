package controller

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/relay/channeltype"
	"github.com/quantumclaw/quantumclaw/relay/quantum"
	"github.com/quantumclaw/quantumclaw/relay/quantum/ionq"
	"github.com/quantumclaw/quantumclaw/relay/quantum/ibmq"
	"github.com/quantumclaw/quantumclaw/relay/quantum/rigetti"
	"github.com/quantumclaw/quantumclaw/relay/quantum/azure"
	"github.com/quantumclaw/quantumclaw/relay/quantum/braket"
)

// QuantumBackend — 量子后端信息
type QuantumBackend struct {
	Provider     string `json:"provider"`
	ProviderID   int    `json:"provider_id"`
	BackendName  string `json:"backend_name"`
	Status       string `json:"status"` // online / offline / maintenance
	QueueDepth   int    `json:"queue_depth"`
}

// GetQuantumBackends — 获取所有可用量子后端
func GetQuantumBackends(c *gin.Context) {
	channels, err := model.GetAllChannels(0, 0, "all")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	typeNames := channeltype.ChannelTypeNames
	var backends []QuantumBackend
	for _, ch := range channels {
		if ch.Type < 100 || ch.Type >= channeltype.QuantumDummy {
			continue
		}
		if ch.Key == "" {
			continue
		}
		providerName := ""
		if name, ok := typeNames[ch.Type]; ok {
			providerName = name
		}
		// 从 channel Models 字段解析 backend 名称
		modelNames := splitModels(ch.Models)
		if len(modelNames) == 0 {
			// 使用默认后端名
			modelNames = []string{providerName + "-default"}
		}
		for _, m := range modelNames {
			backends = append(backends, QuantumBackend{
				Provider:    providerName,
				ProviderID:  ch.Type,
				BackendName: m,
				Status:      "online",  // 默认 online
				QueueDepth:  0,
			})
		}
	}

	if backends == nil {
		backends = []QuantumBackend{}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    backends,
	})
}

// GetQuantumProviders — 获取量子供应商统计
func GetQuantumProviders(c *gin.Context) {
	channels, err := model.GetAllChannels(0, 0, "all")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	typeNames := channeltype.ChannelTypeNames
	type ProviderStat struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		Backends    int    `json:"backends"`
		Configured  bool   `json:"configured"`
	}

	// 所有量子 provider
	allQuantum := []struct {
		ID   int
		Name string
	}{
		{channeltype.IonQ, "IonQ"},
		{channeltype.IBMQ, "IBM Q"},
		{channeltype.Rigetti, "Rigetti"},
		{channeltype.AWSBraket, "AWS Braket"},
		{channeltype.AzureQuantum, "Azure Quantum"},
		{channeltype.GoogleQuantum, "Google Quantum"},
	}

	configuredMap := make(map[int]bool)
	backendCount := make(map[int]int)
	for _, ch := range channels {
		if ch.Type >= 100 && ch.Type < channeltype.QuantumDummy {
			if ch.Key != "" && ch.Key != "PUT_YOUR_API_KEY_HERE" {
				configuredMap[ch.Type] = true
			}
			models := splitModels(ch.Models)
			if len(models) > 0 {
				backendCount[ch.Type] = len(models)
			} else {
				backendCount[ch.Type] = 1
			}
		}
	}

	var result []ProviderStat
	for _, q := range allQuantum {
		displayName := q.Name
		if name, ok := typeNames[q.ID]; ok {
			displayName = name
		}
		result = append(result, ProviderStat{
			ID:          q.ID,
			Name:        displayName,
			DisplayName: displayName,
			Backends:    backendCount[q.ID],
			Configured:  configuredMap[q.ID],
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// getQuantumAdaptor 根据 provider 名称返回对应的量子适配器
func getQuantumAdaptor(channelType int, apiKey string, baseURL string) (quantum.QuantumAdaptor, error) {
	if baseURL == "" {
		baseURL = channeltype.ChannelBaseURLs[channelType]
	}

	switch channelType {
	case channeltype.IonQ:
		return &ionq.Adaptor{APIKey: apiKey, BaseURL: baseURL}, nil
	case channeltype.IBMQ:
		return &ibmq.Adaptor{APIKey: apiKey, BaseURL: baseURL}, nil
	case channeltype.Rigetti:
		return &rigetti.Adaptor{APIKey: apiKey, BaseURL: baseURL}, nil
	case channeltype.AzureQuantum:
		return &azure.Adaptor{APIKey: apiKey, BaseURL: baseURL}, nil
	case channeltype.AWSBraket:
		return &braket.Adaptor{APIKey: apiKey, BaseURL: baseURL}, nil
	default:
		return nil, fmt.Errorf("unknown quantum channel type: %d", channelType)
	}
}

// SubmitQuantumTask — 提交量子任务（使用真实量子适配器）
func SubmitQuantumTask(c *gin.Context) {
	var req struct {
		Provider string `json:"provider"`
		Backend  string `json:"backend"`
		Qasm     string `json:"qasm"`
		Shots    int    `json:"shots"`
		Wait     bool   `json:"wait"` // true=同步等待完成, false=异步返回task id
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	if req.Qasm == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "qasm is required"})
		return
	}
	if req.Shots <= 0 {
		req.Shots = 1024
	}

	// 从 context 获取 channel 信息（由 Distribute 中间件设置）
	channelType := c.GetInt("channel")
	apiKey := strings.TrimPrefix(c.Request.Header.Get("Authorization"), "Bearer ")
	baseURL := c.GetString("base_url")

	// 创建适配器
	adaptorImpl, err := getQuantumAdaptor(channelType, apiKey, baseURL)
	if err != nil {
		logger.Errorf(c.Request.Context(), "quantum: create adaptor: %v", err)
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	// 构建量子任务请求
	qReq := &quantum.QuantumTaskRequest{
		Backend:           req.Backend,
		Shots:             req.Shots,
		OptimizationLevel: 1,
	}
	qReq.Circuit.QASM = req.Qasm
	qReq.Circuit.Qubits = countQubitsFromQASM(req.Qasm)

	// 提交任务
	result, err := adaptorImpl.RunTask(c.Request.Context(), qReq)
	if err != nil {
		logger.Errorf(c.Request.Context(), "quantum: run task: %v", err)
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	// 如果 wait=true，轮询直到完成
	if req.Wait && result.Status != "completed" && result.Status != "failed" {
		result = quantumPollUntilComplete(c.Request.Context(), adaptorImpl, result.TaskID, 2*time.Minute)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"task_id":       result.TaskID,
			"status":        result.Status,
			"provider":      adaptorImpl.ProviderName(),
			"backend":       result.Backend,
			"exec_time_ms":  result.ExecTimeMs,
			"results":       result.Results,
			"error":         result.Error,
		},
	})
}

func quantumPollUntilComplete(ctx context.Context, impl quantum.QuantumAdaptor, taskID string, timeout time.Duration) *quantum.QuantumTaskResult {
	deadline := time.Now().Add(timeout)
	pollInterval := 2 * time.Second

	for time.Now().Before(deadline) {
		result, err := impl.QueryTask(ctx, taskID)
		if err == nil && (result.Status == "completed" || result.Status == "failed" || result.Status == "cancelled") {
			result.ExecTimeMs = int64(time.Since(time.Now().Add(-time.Second * time.Duration(result.ExecTimeMs/1000)))) // 修正时间
			return result
		}

		select {
		case <-ctx.Done():
			return &quantum.QuantumTaskResult{
				TaskID: taskID,
				Status: "cancelled",
				Error:  "context cancelled",
			}
		case <-time.After(pollInterval):
		}
	}

	return &quantum.QuantumTaskResult{
		TaskID: taskID,
		Status: "timeout",
		Error:  fmt.Sprintf("task did not complete within %v", timeout),
	}
}

// countQubitsFromQASM 从 QASM 中提取量子比特数
func countQubitsFromQASM(qasm string) int {
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

func splitModels(models string) []string {
	if models == "" {
		return nil
	}
	var result []string
	for _, m := range strings.Split(models, ",") {
		m = strings.TrimSpace(m)
		if m != "" {
			result = append(result, m)
		}
	}
	return result
}
