package middleware

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/ctxkey"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/monitor"
	"github.com/quantumclaw/quantumclaw/relay/channeltype"
)

type ModelRequest struct {
	Model string `json:"model" form:"model"`
}

// Distribute 链式路由：三级路由原则
//
// 原则 1: 用户指定渠道商优先
// 原则 2: 同品牌同模型 → 最低价优先
// 原则 3: 国内资源不自动切到国外
//
// 路由顺序:
//   Step 0: 请求指定了 channel_id → 直接路由
//   Step 1: 用户设置了 PreferredProviderId → 查首选供应商的渠道
//   Step 2: 用自己的推广人的渠道（promoterId）
//   Step 3: 全资源池兜底（同 region，选最便宜）
//   Step 4: 全部失败 → 报错
func Distribute() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		userId := c.GetInt(ctxkey.Id)
		userGroup, _ := model.CacheGetUserGroup(userId)
		c.Set(ctxkey.Group, userGroup)

		var requestModel string
		var channel *model.Channel

		// ── 余额预检阈值（从配置读取，分）──
		minBalance := config.MinChannelOwnerBalanceCents

		// ── Step 0: 请求指定 channel_id（调试/定向路由）──
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

		// 查推广人
		promoterId := model.GetPromoterId(userId)
		c.Set(ctxkey.PromoterId, promoterId)

		// 获取用户首选供应商
		user, _ := model.GetUserById(userId, false)

		// ── Step 1: 首选供应商 ──
		if user != nil && user.PreferredProviderId > 0 {
			ch, err := selectChannelWithBalance(userGroup, requestModel, user.PreferredProviderId, "", minBalance)
			if err == nil && !monitor.IsModelPenalized(ch.Id, requestModel) {
				logger.Debugf(ctx, "user %d, model %s: using PREFERRED provider channel #%d (owner=%d)",
					userId, requestModel, ch.Id, ch.UserId)
				setupAndContinue(c, ch, requestModel, promoterId, false)
				return
			}
			logger.Debugf(ctx, "user %d preferred provider #%d has no channel for %s, fallback",
				userId, user.PreferredProviderId, requestModel)
		}

	// ── Step 2: 推广人渠道（最便宜优先）──
	if promoterId >= 0 {
		ch, err := selectChannelWithBalance(userGroup, requestModel, promoterId, "", minBalance)
		if err == nil && !monitor.IsModelPenalized(ch.Id, requestModel) {
				logger.Debugf(ctx, "user %d, model %s: using PROMOTER channel #%d (owner=%d)",
					userId, requestModel, ch.Id, promoterId)
				setupAndContinue(c, ch, requestModel, promoterId, false)
				return
			}
		}

		// ── Step 3: 全资源池兜底（原则 3：国内 → 仅限国内）──
		// 先尝试国内渠道
		poolChannel, err := selectChannelWithBalance(userGroup, requestModel, 0, channeltype.RegionChina, minBalance)
		if err == nil && !monitor.IsModelPenalized(poolChannel.Id, requestModel) {
			logger.Debugf(ctx, "user %d, model %s: FALLBACK china channel #%d (owner=%d)",
				userId, requestModel, poolChannel.Id, poolChannel.UserId)
			setupAndContinue(c, poolChannel, requestModel, promoterId, true)
			return
		}

		// 国内没有 → 尝试国外渠道（原则 3：只有国内渠道不可用时才切国外）
		poolChannel, err = selectChannelWithBalance(userGroup, requestModel, 0, channeltype.RegionOverseas, minBalance)
		if err == nil && !monitor.IsModelPenalized(poolChannel.Id, requestModel) {
			logger.Debugf(ctx, "user %d, model %s: FALLBACK overseas channel #%d (owner=%d, from=china_pool)",
				userId, requestModel, poolChannel.Id, poolChannel.UserId)
			setupAndContinue(c, poolChannel, requestModel, promoterId, true)
			return
		}

		// ── Step 4: 全部失败 ──
		message := fmt.Sprintf("当前分组 %s 下对于模型 %s 无可用渠道", userGroup, requestModel)
		if channel != nil {
			logger.SysError(fmt.Sprintf("渠道不存在：%d", channel.Id))
			message = "数据库一致性已被破坏，请联系管理员"
		}
		abortWithMessage(c, http.StatusServiceUnavailable, message)
	}
}

// setupAndContinue 设置上下文并继续请求链
func setupAndContinue(c *gin.Context, channel *model.Channel, modelName string, promoterId int, isFallback bool) {
	c.Set(ctxkey.ChannelOwner, channel.UserId)
	c.Set(ctxkey.IsFallback, isFallback)
	c.Set(ctxkey.FallbackFrom, promoterId)
	SetupContextForSelectedChannel(c, channel, modelName)
	c.Next()
}

// SetupContextForSelectedChannel 设置 channel 上下文
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
		}
	}
	if cfg.APIVersion != "" {
		c.Set(ctxkey.Config, cfg)
	}
}

// selectChannelWithBalance 根据配置选择渠道选择策略
// 当 MinChannelOwnerBalanceCents > 0 时使用带余额预检的版本，否则使用原版（向后兼容）
func selectChannelWithBalance(group string, modelName string, ownerId int, region string, minBalance int64) (*model.Channel, error) {
	if minBalance > 0 {
		return model.GetCheapestSatisfiedChannelWithBalance(group, modelName, ownerId, region, minBalance)
	}
	return model.GetCheapestSatisfiedChannel(group, modelName, ownerId, region)
}
