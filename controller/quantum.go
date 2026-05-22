package controller

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/helper"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/relay/channeltype"
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

// SubmitQuantumTask — 提交量子任务 (模拟)
func SubmitQuantumTask(c *gin.Context) {
	var req struct {
		Provider string `json:"provider"`
		Backend  string `json:"backend"`
		Qasm     string `json:"qasm"`
		Shots    int    `json:"shots"`
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

	// 返回模拟任务 ID
	taskID := "Q-" + strconv.FormatInt(helper.GetTimestamp(), 10)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"task_id": taskID,
			"status":  "queued",
			"provider": req.Provider,
			"backend":  req.Backend,
		},
	})
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
