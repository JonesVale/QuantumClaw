package middleware

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/quantumclaw/quantumclaw/common/ctxkey"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/relay/channeltype"
)

type ModelRequest struct {
	Model string `json:"model" form:"model"`
}

// Distribute 链式路由：所有人走同一套逻辑
//
// 所有用户都有 promoterId（0=平台自己）：
//
//	Step 1: 先用 promoter 自己的 channel（UserId=promoterId）
//	Step 2: 不够/全挂 → 全资源池兜底（不限 UserId）
//	Step 3: 全失败 → 报错
//
// 结算在请求完成后异步写入 token_transaction 表。
func Distribute() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		userId := c.GetInt(ctxkey.Id)
		userGroup, _ := model.CacheGetUserGroup(userId)
		c.Set(ctxkey.Group, userGroup)

		var requestModel string
		var channel *model.Channel

		// ── 如果请求指定了 channel_id，直接用它（调试/定向路由）──
		channelId, ok := c.Get(ctxkey.SpecificChannelId)
		if ok {
			id, err := strconv.Atoi(channelId.(string))
			if err != nil {
				abortWithMessage(c, http.StatusBadRequest, "无效的渠道 Id")
				return
			}
			channel, err = model.GetChannelById(id, true)
			if err != nil {
				abortWithMessage(c, http.StatusBadRequest, "无效的渠道 Id")
				return
			}
			if channel.Status != model.ChannelStatusEnabled {
				abortWithMessage(c, http.StatusForbidden, "该渠道已被禁用")
				return
			}
			setupAndContinue(c, channel, "", 0, false)
			return
		}

		// ── 正常路由 ──
		requestModel = c.GetString(ctxkey.RequestModel)

		// Step 1: 查推广人（0=平台自己）
		promoterId := model.GetPromoterId(userId)
		c.Set(ctxkey.PromoterId, promoterId)

		// Step 2: 第一层——用推广人自己的 channel
		if promoterId >= 0 {
			ch, err := model.CacheGetRandomSatisfiedChannelByOwner(userGroup, requestModel, promoterId)
			if err == nil {
				logger.Debugf(ctx, "user %d, model %s: using promoter channel #%d (owner=%d)",
					userId, requestModel, ch.Id, promoterId)
				setupAndContinue(c, ch, requestModel, promoterId, false)
				return
			}
		}

		// Step 3: 第二层——全资源池兜底（所有渠道商 Key 混用）
		poolChannel, err := model.CacheGetRandomSatisfiedChannelAnyOwner(userGroup, requestModel, promoterId)
		if err == nil {
			logger.Debugf(ctx, "user %d, model %s: FALLBACK pool channel #%d (owner=%d, fallback_from=%d)",
				userId, requestModel, poolChannel.Id, poolChannel.UserId, promoterId)
			setupAndContinue(c, poolChannel, requestModel, promoterId, true)
			return
		}

		// Step 4: 全部失败
		message := fmt.Sprintf("当前分组 %s 下对于模型 %s 无可用渠道", userGroup, requestModel)
		if channel != nil {
			logger.SysError(fmt.Sprintf("渠道不存在：%d", channel.Id))
			message = "数据库一致性已被破坏，请联系管理员"
		}
		abortWithMessage(c, http.StatusServiceUnavailable, message)
	}
}

// setupAndContinue 设置上下文并继续请求链
// 注意：交易流水不在 middleware 写入，由 RecordConsumeLog 中的 createTransactionFromLog 负责
// 该函数使用精确的 prompt_tokens + completion_tokens 计算金额
func setupAndContinue(c *gin.Context, channel *model.Channel, modelName string, promoterId int, isFallback bool) {
	c.Set(ctxkey.ChannelOwner, channel.UserId)
	c.Set(ctxkey.IsFallback, isFallback)
	c.Set(ctxkey.FallbackFrom, promoterId)
	SetupContextForSelectedChannel(c, channel, modelName)
	c.Next()
}

// SetupContextForSelectedChannel 设置 channel 上下文（已有逻辑不动）
func SetupContextForSelectedChannel(c *gin.Context, channel *model.Channel, modelName string) {
	c.Set(ctxkey.Channel, channel.Type)
	c.Set(ctxkey.ChannelId, channel.Id)
	c.Set(ctxkey.ChannelName, channel.Name)
	if channel.SystemPrompt != nil && *channel.SystemPrompt != "" {
		c.Set(ctxkey.SystemPrompt, *channel.SystemPrompt)
	}
	c.Set(ctxkey.ModelMapping, channel.GetModelMapping())
	c.Set(ctxkey.OriginalModel, modelName)
	c.Request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", channel.Key))
	c.Set(ctxkey.BaseURL, channel.GetBaseURL())
	cfg, _ := channel.LoadConfig()
	if channel.Other != nil {
		switch channel.Type {
		case channeltype.Azure:
			if cfg.APIVersion == "" {
				cfg.APIVersion = *channel.Other
			}
		case channeltype.Xunfei:
			if cfg.APIVersion == "" {
				cfg.APIVersion = *channel.Other
			}
		case channeltype.Gemini:
			if cfg.APIVersion == "" {
				cfg.APIVersion = *channel.Other
			}
		case channeltype.AIProxyLibrary:
			if cfg.LibraryID == "" {
				cfg.LibraryID = *channel.Other
			}
		case channeltype.Ali:
			if cfg.Plugin == "" {
				cfg.Plugin = *channel.Other
			}
		}
	}
	c.Set(ctxkey.Config, cfg)
}
