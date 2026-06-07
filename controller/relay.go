package controller

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/ctxkey"
	"github.com/quantumclaw/quantumclaw/common/helper"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/middleware"
	dbmodel "github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/monitor"
	"github.com/quantumclaw/quantumclaw/relay/controller"
	"github.com/quantumclaw/quantumclaw/relay/model"
	"github.com/quantumclaw/quantumclaw/relay/relaymode"
)

// https://platform.openai.com/docs/api-reference/chat

func relayHelper(c *gin.Context, relayMode int) *model.ErrorWithStatusCode {
	var err *model.ErrorWithStatusCode
	switch relayMode {
	case relaymode.ImagesGenerations:
		err = controller.RelayImageHelper(c, relayMode)
	case relaymode.AudioSpeech:
		fallthrough
	case relaymode.AudioTranslation:
		fallthrough
	case relaymode.AudioTranscription:
		err = controller.RelayAudioHelper(c, relayMode)
	case relaymode.Proxy:
		err = controller.RelayProxyHelper(c, relayMode)
	case relaymode.Assistants, relaymode.AssistantsThreads:
		err = RelayAssistantsHelper(c)
	case relaymode.AssistantsFiles:
		err = RelayFilesHelper(c)
	case relaymode.Files:
		err = RelayFilesHelper(c)
	case relaymode.FineTuning:
		err = RelayFineTuningHelper(c)
	// 异步任务：Midjourney / 视频生成 / Suno 音乐
	case relaymode.Midjourney:
		fallthrough
	case relaymode.VideoGeneration:
		fallthrough
	case relaymode.Suno:
		controller.RelayAsyncTask(c)
		// 异步任务不直接返回错误给客户端，已在函数内处理响应
		return nil
	default:
		err = controller.RelayTextHelper(c)
	}
	return err
}

func Relay(c *gin.Context) {
	ctx := c.Request.Context()
	relayMode := relaymode.GetByPath(c.Request.URL.Path)
	if config.DebugEnabled {
		requestBody, _ := common.GetRequestBody(c)
		logger.Debugf(ctx, "request body: %s", string(requestBody))
	}
	channelId := c.GetInt(ctxkey.ChannelId)
	userId := c.GetInt(ctxkey.Id)
	bizErr := relayHelper(c, relayMode)
	if bizErr == nil {
		monitor.Emit(channelId, true)
		return
	}
	lastFailedChannelId := channelId
	channelName := c.GetString(ctxkey.ChannelName)
	group := c.GetString(ctxkey.Group)
	originalModel := c.GetString(ctxkey.OriginalModel)
	// 传值复制，避免 goroutine 闭包捕获指针
	{
		firstBizErr := *bizErr
		go processChannelRelayError(ctx, userId, channelId, channelName, firstBizErr)
	}
	requestId := c.GetString(helper.RequestIdKey)
	retryTimes := config.RetryTimes
	if !shouldRetry(c, bizErr.StatusCode) {
		logger.Errorf(ctx, "relay error happen, status code is %d, won't retry in this case", bizErr.StatusCode)
		retryTimes = 0
	}
	// 检查用户是否同意使用平台资源池
	if retryTimes > 0 {
		usePool := dbmodel.IsConsentValid(userId)
		if !usePool {
			retryTimes = 0
		}
	}
	for i := retryTimes; i > 0; i-- {
		// 重试时按价格排序取最便宜（排除已失败的渠道）
		channel, err := dbmodel.GetCheapestSatisfiedChannel(group, originalModel, 0, "")
		if err != nil {
			logger.Errorf(ctx, "GetCheapestSatisfiedChannel failed: %+v", err)
			break
		}
		logger.Infof(ctx, "using channel #%d to retry (remain times %d)", channel.Id, i)
		if channel.Id == lastFailedChannelId {
			continue
		}
		middleware.SetupContextForSelectedChannel(c, channel, originalModel)
		requestBody, err := common.GetRequestBody(c)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		retryErr := relayHelper(c, relayMode)
		if retryErr == nil {
			// 重试成功 → 把之前失败的渠道加入冷期
			monitor.PenalizeModel(lastFailedChannelId, originalModel)
			return
		}
		channelId := c.GetInt(ctxkey.ChannelId)
		lastFailedChannelId = channelId
		channelName := c.GetString(ctxkey.ChannelName)
		// 传值复制，避免 goroutine 闭包捕获指针
		{
			errCopy := *retryErr
			go processChannelRelayError(ctx, userId, channelId, channelName, errCopy)
		}
		bizErr = retryErr
	}
	if bizErr != nil {
		if bizErr.StatusCode == http.StatusTooManyRequests {
			bizErr.Error.Message = "当前分组上游负载已饱和，请稍后再试"
		}

		bizErr.Error.Message = helper.MessageWithRequestId(bizErr.Error.Message, requestId)
		c.JSON(bizErr.StatusCode, gin.H{
			"error": bizErr.Error,
		})
	}
}

func shouldRetry(c *gin.Context, statusCode int) bool {
	if _, ok := c.Get(ctxkey.SpecificChannelId); ok {
		return false
	}
	if statusCode == http.StatusTooManyRequests {
		return true
	}
	if statusCode/100 == 5 {
		return true
	}
	if statusCode == http.StatusBadRequest {
		return false
	}
	if statusCode/100 == 2 {
		return false
	}
	return true
}

func processChannelRelayError(ctx context.Context, userId int, channelId int, channelName string, err model.ErrorWithStatusCode) {
	logger.Errorf(ctx, "relay error (channel id %d, user id: %d): %s", channelId, userId, err.Message)
	// https://platform.openai.com/docs/guides/error-codes/api-errors
	if monitor.ShouldDisableChannel(&err.Error, err.StatusCode) {
		monitor.DisableChannel(channelId, channelName, err.Message)
	} else {
		monitor.Emit(channelId, false)
	}
}

func RelayNotImplemented(c *gin.Context) {
	err := model.Error{
		Message: "API not implemented",
		Type:    "quantumclaw_error",
		Param:   "",
		Code:    "api_not_implemented",
	}
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": err,
	})
}

func RelayNotFound(c *gin.Context) {
	err := model.Error{
		Message: fmt.Sprintf("Invalid URL (%s %s)", c.Request.Method, c.Request.URL.Path),
		Type:    "invalid_request_error",
		Param:   "",
		Code:    "",
	}
	c.JSON(http.StatusNotFound, gin.H{
		"error": err,
	})
}
