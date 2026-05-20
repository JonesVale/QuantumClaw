package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/ctxkey"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/middleware"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/relay"
	"github.com/quantumclaw/quantumclaw/relay/quantum"
	"github.com/quantumclaw/quantumclaw/relay/quantum/azure"
	"github.com/quantumclaw/quantumclaw/relay/quantum/braket"
	"github.com/quantumclaw/quantumclaw/relay/quantum/ibmq"
	"github.com/quantumclaw/quantumclaw/relay/quantum/ionq"
	"github.com/quantumclaw/quantumclaw/relay/quantum/rigetti"
	"github.com/quantumclaw/quantumclaw/relay/relaymode"
	"github.com/quantumclaw/quantumclaw/service"
)

// QuantumRelay — 统一量子路由分发
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
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "unknown quantum relay mode"})
	}
}

// quantumRunHandler — 提交量子任务，带自动重试
func quantumRunHandler(c *gin.Context) {
	channelType := c.GetInt(ctxkey.Channel)
	channelID := c.GetInt(ctxkey.ChannelId)

	result, apiErr := doQuantumRun(c, channelType, channelID)
	if apiErr == nil {
		if result.CostQuota > 0 {
			recordQuantumConsumption(c, result)
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
		return
	}

	group := c.GetString(ctxkey.Group)
	originalModel := c.GetString(ctxkey.OriginalModel)
	retryTimes := config.RetryTimes
	if retryTimes <= 0 {
		retryTimes = 3
	}

	lastFailedChannelID := channelID
	for i := retryTimes; i > 0; i-- {
		channel, err := model.CacheGetRandomSatisfiedChannel(group, originalModel, i != retryTimes)
		if err != nil {
			logger.Errorf(c.Request.Context(), "quantum retry: channel select failed: %v", err)
			break
		}
		if channel.Id == lastFailedChannelID {
			continue
		}
		middleware.SetupContextForSelectedChannel(c, channel, originalModel)

		channelType = c.GetInt(ctxkey.Channel)
		channelID = c.GetInt(ctxkey.ChannelId)
		result, apiErr = doQuantumRun(c, channelType, channelID)
		if apiErr == nil {
			if result.CostQuota > 0 {
				recordQuantumConsumption(c, result)
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
			return
		}
		lastFailedChannelID = channelID
	}

	c.JSON(http.StatusServiceUnavailable, gin.H{
		"success": false, "message": "all quantum channels failed, please try again later",
	})
}

func doQuantumRun(c *gin.Context, channelType, channelID int) (*quantum.QuantumTaskResult, error) {
	qAdaptor, err := relay.GetQuantumAdaptor(channelType)
	if err != nil {
		return nil, err
	}

	var req quantum.QuantumTaskRequest
	if err := common.UnmarshalBodyReusable(c, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	setupAdaptorAuth(c, qAdaptor)

	result, err := qAdaptor.RunTask(c.Request.Context(), &req)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func quantumStatusHandler(c *gin.Context) {
	channelType := c.GetInt(ctxkey.Channel)
	taskID := c.Param("task_id")

	qAdaptor, err := relay.GetQuantumAdaptor(channelType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	setupAdaptorAuth(c, qAdaptor)

	result, err := qAdaptor.QueryTask(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func quantumCancelHandler(c *gin.Context) {
	channelType := c.GetInt(ctxkey.Channel)
	taskID := c.Param("task_id")

	qAdaptor, err := relay.GetQuantumAdaptor(channelType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	setupAdaptorAuth(c, qAdaptor)

	if err := qAdaptor.CancelTask(c.Request.Context(), taskID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "cancelled"})
}

func quantumBackendsHandler(c *gin.Context) {
	channelType := c.GetInt(ctxkey.Channel)

	qAdaptor, err := relay.GetQuantumAdaptor(channelType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	setupAdaptorAuth(c, qAdaptor)

	backends, err := qAdaptor.ListBackends(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": backends})
}

func setupAdaptorAuth(c *gin.Context, qa quantum.QuantumAdaptor) {
	apiKey := extractAPIKey(c)
	baseURL := c.GetString(ctxkey.BaseURL)

	switch a := qa.(type) {
	case *ionq.Adaptor:
		a.APIKey = apiKey
		if baseURL == "" {
			baseURL = "https://api.ionq.co"
		}
		a.BaseURL = baseURL
	case *ibmq.Adaptor:
		a.APIKey = apiKey
		if baseURL == "" {
			baseURL = "https://api.quantum.ibm.com"
		}
		a.BaseURL = baseURL
	case *rigetti.Adaptor:
		a.APIKey = apiKey
		if baseURL == "" {
			baseURL = "https://api.qcs.rigetti.com"
		}
		a.BaseURL = baseURL
	case *braket.Adaptor:
		a.APIKey = apiKey
		if baseURL == "" {
			baseURL = "https://braket.us-west-1.amazonaws.com"
		}
		a.BaseURL = baseURL
	case *azure.Adaptor:
		a.APIKey = apiKey
		if baseURL == "" {
			baseURL = "https://quantum.azure.com/api"
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
