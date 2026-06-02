package service

import (
	"context"
	"fmt"
	"math"

	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/helper"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
	billingratio "github.com/quantumclaw/quantumclaw/relay/billing/ratio"
	"github.com/quantumclaw/quantumclaw/relay/meta"
	relaymodel "github.com/quantumclaw/quantumclaw/relay/model"
)

// ==================== 现金计费 — 预检查 ====================

// PreConsumeBalance 消费前检查余额是否足够
// 只检查不预扣，简化回退逻辑
// 返回：estimatedPrice（预估价，分），error
func PreConsumeBalance(ctx context.Context, meta *meta.Meta, promptTokens int, ratio float64) (int64, error) {
	// 读取渠道信息，获取 CostPerUnit + SellPriceRate
	channel, err := model.GetChannelById(meta.ChannelId, true)
	if err != nil {
		return 0, fmt.Errorf("get channel: %w", err)
	}

	// 预估价 = 按 prompt 估算的配额 × CostPerUnit × SellPriceRate
	estimatedQuota := getEstimatedQuota(promptTokens, ratio)
	estimatedPrice := quotaToPrice(estimatedQuota, channel.CostPerUnit, channel.SellPriceRate)

	// 查用户现金余额
	balance, err := model.GetUserCashBalance(meta.UserId)
	if err != nil {
		return 0, fmt.Errorf("get user balance: %w", err)
	}

	// 余额必须 ≥ 预估费用
	if balance < 1 {
		return estimatedPrice, fmt.Errorf("余额不足，当前余额 %d 分，需要至少 1 分", balance)
	}
	if balance < estimatedPrice {
		return estimatedPrice, fmt.Errorf("余额不足，当前 %d 分，预估需要 %d 分", balance, estimatedPrice)
	}

	return estimatedPrice, nil
}

// ==================== 现金计费 — 消费后扣款 ====================

// PostConsumeDeduct 消费成功后扣款 + 分账
// usage: 实际用量（含 prompt_tokens, completion_tokens）
// ratio: modelRatio * groupRatio（沿用现有倍率体系）
func PostConsumeDeduct(ctx context.Context, meta *meta.Meta, usage *relaymodel.Usage, textRequest *relaymodel.GeneralOpenAIRequest,
	ratio float64, modelRatio float64, groupRatio float64, preConsumedQuota int64, systemPromptReset bool) error {

	if usage == nil {
		return fmt.Errorf("usage is nil, cannot deduct")
	}

	// 1. 读取渠道
	channel, err := model.GetChannelById(meta.ChannelId, true)
	if err != nil {
		return fmt.Errorf("get channel %d: %w", meta.ChannelId, err)
	}

	// 2. 计算实际配额（沿用现有逻辑）
	quota := calculateQuota(usage, textRequest, ratio, meta.ChannelType, meta.Config.CacheBillingRatio)
	if quota <= 0 {
		return fmt.Errorf("quota is %d, skip deduction", quota)
	}

	// 3. 配额 → 现金价格（分）
	priceCents := quotaToPrice(quota, channel.CostPerUnit, channel.SellPriceRate)
	if priceCents < 1 {
		priceCents = 1
	}

	// 4. 扣消费者余额（三级优先：现金→佣金→配额，仍不足则记 Debt）
	balance, err := model.GetUserCashBalance(meta.UserId)
	if err != nil {
		return fmt.Errorf("get user balance: %w", err)
	}
	remaining := priceCents

	// Tier 1: 现金余额
	if balance > 0 {
		deduct := balance
		if deduct > remaining {
			deduct = remaining
		}
		model.MinusUserCashBalance(meta.UserId, deduct)
		remaining -= deduct
	}

	// Tier 2: 佣金余额
	if remaining > 0 {
		commBalance, err := model.GetUserCommissionBalance(meta.UserId)
		if err == nil && commBalance > 0 {
			deduct := commBalance
			if deduct > remaining {
				deduct = remaining
			}
			model.MinusUserCommissionBalance(meta.UserId, deduct)
			remaining -= deduct
		}
	}

	// Tier 3: 配额回退（只扣可用配额，避免负数）
	if remaining > 0 {
		userQuota, qErr := model.CacheGetUserQuota(ctx, meta.UserId)
		if qErr == nil && userQuota > 0 {
			deduct := userQuota
			if deduct > remaining {
				deduct = remaining
			}
			model.DecreaseUserQuota(meta.UserId, deduct)
			model.PreConsumeTokenQuota(meta.TokenId, deduct)
			remaining -= deduct
		}
	}

	// Tier 4: 仍不足 → 记追偿挂账（Debt），请求已转发完成必须记账
	if remaining > 0 {
		model.DB.Model(&model.User{}).Where("id = ?", meta.UserId).
			UpdateColumn("debt", gorm.Expr("COALESCE(debt,0) + ?", remaining))
		// 欠费超过阈值 → 自动禁用账号
		var totalDebt int64
		model.DB.Model(&model.User{}).Where("id = ?", meta.UserId).Select("COALESCE(debt,0)").Find(&totalDebt)
		model.RecordLog(ctx, meta.UserId, model.LogTypeSystem,
			fmt.Sprintf("消费追偿挂账 %d 分，累计欠费 %d 分", remaining, totalDebt))
		logger.Warnf(ctx, "user %d: debt %d added, total debt %d", meta.UserId, remaining, totalDebt)
		if config.DebtDisableThreshold > 0 && totalDebt >= config.DebtDisableThreshold {
			model.DB.Model(&model.User{}).Where("id = ?", meta.UserId).
				Update("status", model.UserStatusDisabled)
			model.RecordLog(ctx, meta.UserId, model.LogTypeSystem,
				fmt.Sprintf("欠费 %d 分超过阈值 %d 分，账号已自动禁用", totalDebt, config.DebtDisableThreshold))
		}
	}

	// 5. 记余额流水
	actualDeducted := priceCents - remaining
	remark := fmt.Sprintf("模型:%s 配额:%d 价格:%d分 实扣:%d分 挂账:%d分",
		textRequest.Model, quota, priceCents, actualDeducted, remaining)
	if err := model.CreateBalanceLog(meta.UserId, model.BalanceLogTypeConsume, -actualDeducted, balance-actualDeducted, meta.ChannelId, remark); err != nil {
		logger.Error(ctx, fmt.Sprintf("create balance log: %v", err))
	}

	// 6. 分账
	if channel.UserId > 0 {
		// 使用渠道级分账比例
		split := channel.ProfitSplit
		if split <= 0 || split > 1 {
			split = 0.85 // 默认渠道商 85%
		}
		commissionAmount := int64(float64(priceCents) * (1.0 - split)) // 平台抽成
		netAmount := priceCents - commissionAmount                     // 渠道商净得
		if netAmount < 0 {
			netAmount = 0
		}
		if err := model.CreateProviderEarning(
			int64(channel.UserId),
			int64(meta.ChannelId),
			int64(meta.UserId),
			priceCents,
			commissionAmount,
			netAmount,
			helper.GetCurrentMonth(),
			model.EarningStatusSettled,
		); err != nil {
			logger.Error(ctx, fmt.Sprintf("create provider earning: %v", err))
		}
	}

	// 7. 保留日志 + 渠道用量统计
	logContent := fmt.Sprintf("倍率：%.2f × %.2f | 现金扣款：%d分", modelRatio, groupRatio, priceCents)
	model.RecordConsumeLog(ctx, &model.Log{
		UserId:            meta.UserId,
		ChannelId:         meta.ChannelId,
		PromptTokens:      usage.PromptTokens,
		CompletionTokens:  usage.CompletionTokens,
		ModelName:         textRequest.Model,
		TokenName:         meta.TokenName,
		Quota:             int(quota),
		Content:           logContent,
		IsStream:          meta.IsStream,
		ElapsedTime:       helper.CalcElapsedTime(meta.StartTime),
		SystemPromptReset: systemPromptReset,
		PromoterId:        meta.PromoterId,
		ChannelOwnerId:    meta.ChannelOwnerId,
		IsFallback:        meta.IsFallback,
	})
	model.UpdateUserUsedQuotaAndRequestCount(meta.UserId, quota)
	model.UpdateChannelUsedQuota(meta.ChannelId, quota)
	return nil
}

// ==================== 辅助函数 ====================

// getEstimatedQuota 估算请求配额（仅 prompt 部分）
func getEstimatedQuota(promptTokens int, ratio float64) int64 {
	return int64(float64(config.PreConsumedQuota+int64(promptTokens)) * ratio)
}

// calculateQuota 计算实际消耗配额（沿用现有逻辑）
func calculateQuota(usage *relaymodel.Usage, textRequest *relaymodel.GeneralOpenAIRequest,
	ratio float64, channelType int, cacheBillingRatio float64) int64 {

	completionRatio := billingratio.GetCompletionRatio(textRequest.Model, channelType)
	promptTokens := usage.PromptTokens
	completionTokens := usage.CompletionTokens

	if cacheBillingRatio > 0 && usage.PromptTokensDetails != nil && usage.PromptTokensDetails.CachedTokens > 0 {
		cachedTokens := usage.PromptTokensDetails.CachedTokens
		if cachedTokens > promptTokens {
			cachedTokens = promptTokens
		}
		nonCachedTokens := promptTokens - cachedTokens
		promptTokens = nonCachedTokens + int(math.Ceil(float64(cachedTokens)*cacheBillingRatio))
		if promptTokens < 0 {
			promptTokens = 0
		}
	}

	quota := int64(math.Ceil((float64(promptTokens) + float64(completionTokens)*completionRatio) * ratio))
	if ratio != 0 && quota <= 0 {
		quota = 1
	}
	totalTokens := promptTokens + completionTokens
	if totalTokens == 0 {
		quota = 0
	}
	return quota
}

// quotaToPrice 配额 → 现金价格（分）
// price = quota × (CostPerUnit / 1000000) × SellPriceRate × 100（USD→分）
func quotaToPrice(quota int64, costPerUnit, sellPriceRate float64) int64 {
	if quota <= 0 {
		return 0
	}
	costUsd := float64(quota) * costPerUnit / 1000000.0
	priceUsd := costUsd * sellPriceRate
	priceCents := int64(math.Ceil(priceUsd * 100.0))
	if priceCents < 1 {
		priceCents = 1
	}
	return priceCents
}

// PostConsumeQuantumDeduct 量子任务扣款
func PostConsumeQuantumDeduct(userId, channelId int, costQuota int64) error {
	if costQuota <= 0 {
		return nil
	}
	// 读取渠道获取 CostPerUnit + SellPriceRate
	channel, err := model.GetChannelById(channelId, true)
	if err != nil {
		return fmt.Errorf("get channel %d: %w", channelId, err)
	}
	// 配额 → 现金价格（分）
	priceCents := quotaToPrice(costQuota, channel.CostPerUnit, channel.SellPriceRate)
	// 扣消费者余额
	balance, err := model.GetUserCashBalance(userId)
	if err != nil {
		return fmt.Errorf("get user balance: %w", err)
	}
	if balance < priceCents {
		return fmt.Errorf("user %d balance %d < price %d", userId, balance, priceCents)
	}
	if err := model.MinusUserCashBalance(userId, priceCents); err != nil {
		return fmt.Errorf("deduct user %d balance: %w", userId, err)
	}
	// 余额流水
	newBalance := balance - priceCents
	_ = model.CreateBalanceLog(userId, model.BalanceLogTypeConsume, -priceCents, newBalance, channelId, "quantum task")
	// 分账
	if channel.UserId > 0 {
		netAmount := priceCents
		if netAmount < 0 {
			netAmount = 0
		}
		_ = model.CreateProviderEarning(int64(channel.UserId), int64(channelId), int64(userId),
			priceCents, 0, netAmount, helper.GetCurrentMonth(), model.EarningStatusSettled)
	}
	return nil
}
