package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/ctxkey"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/relay"
	"github.com/quantumclaw/quantumclaw/relay/quantum"
	"github.com/quantumclaw/quantumclaw/relay/quantum/ibmq"
	"github.com/quantumclaw/quantumclaw/relay/quantum/ionq"
	"github.com/quantumclaw/quantumclaw/relay/relaymode"
	"github.com/quantumclaw/quantumclaw/service"
)

// QuantumRelay 统一量子控制器 — 提交/查询/取消/列举
func QuantumRelay(c *gin.Context) {
	relayMode := relaymode.GetByPath(c.Request.URL.Path)

	switch relayMode {
	case relaymode.QuantumRun:
		quantumRunHandler(c)
	case relaymode.QuantumStatus:
		quantumStatusHandler(c)
	case relaymode.QuantumCancel:
		quantumCancelHandler(c)
	case relaymode.QuantumBackends:
		quantumBackendsHandler(c)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown quantum relay mode"})
	}
}

func quantumRunHandler(c *gin.Context) {
	channelType := c.GetInt(ctxkey.Channel)

	qAdaptor, err := relay.GetQuantumAdaptor(channelType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req quantum.QuantumTaskRequest
	if err := common.UnmarshalBodyReusable(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid request: %v", err)})
		return
	}

	setupAdaptorAuth(c, qAdaptor)

	result, err := qAdaptor.RunTask(c.Request.Context(), &req)
	if err != nil {
		logger.Errorf(c.Request.Context(), "quantum run failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.CostQuota > 0 {
		recordQuantumConsumption(c, result)
	}

	c.JSON(http.StatusOK, result)
}

func quantumStatusHandler(c *gin.Context) {
	channelType := c.GetInt(ctxkey.Channel)
	taskID := c.Param("task_id")

	qAdaptor, err := relay.GetQuantumAdaptor(channelType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	setupAdaptorAuth(c, qAdaptor)

	result, err := qAdaptor.QueryTask(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func quantumCancelHandler(c *gin.Context) {
	channelType := c.GetInt(ctxkey.Channel)
	taskID := c.Param("task_id")

	qAdaptor, err := relay.GetQuantumAdaptor(channelType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	setupAdaptorAuth(c, qAdaptor)

	if err := qAdaptor.CancelTask(c.Request.Context(), taskID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "cancelled"})
}

func quantumBackendsHandler(c *gin.Context) {
	channelType := c.GetInt(ctxkey.Channel)

	qAdaptor, err := relay.GetQuantumAdaptor(channelType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	setupAdaptorAuth(c, qAdaptor)

	backends, err := qAdaptor.ListBackends(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"backends": backends})
}

func setupAdaptorAuth(c *gin.Context, qa quantum.QuantumAdaptor) {
	apiKey := extractAPIKey(c)

	switch a := qa.(type) {
	case *ionq.Adaptor:
		a.APIKey = apiKey
		baseURL := c.GetString(ctxkey.BaseURL)
		if baseURL == "" {
			baseURL = "https://api.ionq.co"
		}
		a.BaseURL = baseURL
	case *ibmq.Adaptor:
		a.APIKey = apiKey
		baseURL := c.GetString(ctxkey.BaseURL)
		if baseURL == "" {
			baseURL = "https://api.quantum.ibm.com"
		}
		a.BaseURL = baseURL
	}
}

func extractAPIKey(c *gin.Context) string {
	auth := c.Request.Header.Get("Authorization")
	if len(auth) > 7 && auth[:7] == "Bearer " {
		return auth[7:]
	}
	return auth
}

func recordQuantumConsumption(c *gin.Context, result *quantum.QuantumTaskResult) {
	userID := c.GetInt(ctxkey.Id)
	tokenID := c.GetInt(ctxkey.TokenId)
	tokenName := c.GetString(ctxkey.TokenName)
	channelID := c.GetInt(ctxkey.ChannelId)

	model.RecordConsumeLog(c, &model.Log{
		UserId:    userID,
		ChannelId: channelID,
		ModelName: result.Backend,
		TokenName: tokenName,
		Quota:     int(result.CostQuota),
		Content:   fmt.Sprintf("quantum task %s on %s", result.TaskID, result.Provider),
	})
	service.UpdateUserUsedQuotaAndRequestCount(userID, result.CostQuota)
	service.UpdateChannelUsedQuota(channelID, result.CostQuota)

	if tokenID > 0 {
		_ = model.PostConsumeTokenQuota(tokenID, result.CostQuota)
		_ = model.CacheUpdateUserQuota(c.Request.Context(), userID)
	}
}
