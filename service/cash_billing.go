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
	"gorm.io/gorm"
)

// ==================== 现金计费 — 消费后扣款 ====================

// PostConsumeDeduct 消费成功后扣款 + 分账
// billingSource 来自 PreConsumeBilling 的优先级链结果：
//
//	"subscription" — 订阅已预扣，跳过现金/佣金/配额扣款
//	"cash"         — 现金未预扣，正常走现金→佣金→配额→挂账链
//	"commission"   — 佣金未预扣，跳过现金直接扣佣金→配额→挂账
//	"quota"        — 配额已预扣，跳过现金/佣金，只做配额对齐
//	""             — 兼容旧调用方（audio/image），默认走全链
func PostConsumeDeduct(ctx context.Context, meta *meta.Meta, usage *relaymodel.Usage, textRequest *relaymodel.GeneralOpenAIRequest,
	ratio float64, modelRatio float64, groupRatio float64, preConsumedQuota int64, systemPromptReset bool, billingSource string) error {

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

	// 4. 扣消费者余额（根据 billingSource 决定扣款策略）
	remaining := priceCents
	var deductedFromCash int64   // 实际从现金扣了多少（用于余额流水）
	initialCashBalance, _ := model.GetUserCashBalance(meta.UserId)

	switch billingSource {
	case "subscription":
		// 订阅已预扣，不再重复扣现金/佣金/配额
		remaining = 0

	case "commission":
		// PreConsumeBilling 仅检查了佣金余额，未实际扣款
		// 直接扣佣金
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
		// 佣金不足 → 走配额兜底
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
		// 仍不足 → 记挂账
		if remaining > 0 {
			recordDebt(ctx, meta.UserId, remaining)
			remaining = 0
		}

	case "quota":
		// PreConsumeBilling 已预扣配额，此处只对齐余额
		// 如果还有剩余，记挂账
		if remaining > 0 {
			recordDebt(ctx, meta.UserId, remaining)
			remaining = 0
		}

	default:
		// "cash" 或 "" — 正常走全链：现金 → 佣金 → 配额 → 挂账
		// Tier 1: 现金余额
		if remaining > 0 {
			cashBalance, err := model.GetUserCashBalance(meta.UserId)
			if err == nil && cashBalance > 0 {
				deduct := cashBalance
				if deduct > remaining {
					deduct = remaining
				}
				model.MinusUserCashBalance(meta.UserId, deduct)
				deductedFromCash = deduct
				remaining -= deduct
			}
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

		// Tier 3: 配额回退
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

		// Tier 4: 仍不足 → 记追偿挂账
		if remaining > 0 {
			recordDebt(ctx, meta.UserId, remaining)
			remaining = 0
		}
	}

	// 5. 记余额流水（仅记录现金余额变动）
	actualDeducted := priceCents - remaining
	newCashBalance := initialCashBalance - deductedFromCash
	if newCashBalance < 0 {
		newCashBalance = 0
	}
	remark := fmt.Sprintf("模型:%s 配额:%d 价格:%d分 来源:%s 实扣:%d分 挂账:%d分",
		textRequest.Model, quota, priceCents, billingSource, actualDeducted, remaining)
	if err := model.CreateBalanceLog(meta.UserId, model.BalanceLogTypeConsume, -deductedFromCash, newCashBalance, meta.ChannelId, remark); err != nil {
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
	logContent := fmt.Sprintf("倍率：%.2f × %.2f | 扣款：%d分 来源:%s", modelRatio, groupRatio, priceCents, billingSource)
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
// 保留用于单元测试兼容
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

// recordDebt 记录追偿挂账（仅记账，不封号）
// 用户余额不足时记录欠费用于追踪，不影响登录和使用
// 欠费用户在下次请求时会直接在 API 层面被拒绝，但依然可以登录和充值
func recordDebt(ctx context.Context, userId int, amount int64) {
	model.DB.Model(&model.User{}).Where("id = ?", userId).
		UpdateColumn("debt", gorm.Expr("COALESCE(debt,0) + ?", amount))
	var totalDebt int64
	model.DB.Model(&model.User{}).Where("id = ?", userId).Select("COALESCE(debt,0)").Find(&totalDebt)
	model.RecordLog(ctx, userId, model.LogTypeSystem,
		fmt.Sprintf("消费追偿挂账 %d 分，累计欠费 %d 分", amount, totalDebt))
	logger.Warnf(ctx, "user %d: debt %d added, total debt %d", userId, amount, totalDebt)
}

// PostConsumeQuantumDeduct 量子任务扣款
// 使用与 PostConsumeDeduct 一致的三级扣款链：现金→佣金→配额
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

	remaining := priceCents

	// Tier 1: 现金余额
	if remaining > 0 {
		cashBalance, err := model.GetUserCashBalance(userId)
		if err == nil && cashBalance > 0 {
			deduct := cashBalance
			if deduct > remaining {
				deduct = remaining
			}
			if err := model.MinusUserCashBalance(userId, deduct); err != nil {
				return fmt.Errorf("deduct user %d cash: %w", userId, err)
			}
			remaining -= deduct
		}
	}

	// Tier 2: 佣金余额
	if remaining > 0 {
		commBalance, err := model.GetUserCommissionBalance(userId)
		if err == nil && commBalance > 0 {
			deduct := commBalance
			if deduct > remaining {
				deduct = remaining
			}
			if err := model.MinusUserCommissionBalance(userId, deduct); err != nil {
				return fmt.Errorf("deduct user %d commission: %w", userId, err)
			}
			remaining -= deduct
		}
	}

	// Tier 3: 不足 → 返回错误（量子任务不支持挂账）
	if remaining > 0 {
		return fmt.Errorf("user %d balance + commission %d < price %d", userId, priceCents-remaining, priceCents)
	}

	// 余额流水
	actualDeducted := priceCents - remaining
	_ = model.CreateBalanceLog(userId, model.BalanceLogTypeConsume, -actualDeducted, remaining, channelId, "quantum task")

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
