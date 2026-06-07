package common_handler

import (
	"context"
	"fmt"

	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/relay/common"
)

// ==================== 计费优先级链 ====================
// 优先级：订阅 → CashBalance → CommissionBalance → Quota
//
// 返佣（CommissionBalance）是推广奖励积分，可消费也可提现。

// PreConsumeBilling 统一预扣入口
// billingSource: "subscription" | "cash" | "commission" | "quota"
func PreConsumeBilling(ctx context.Context, meta *common.RelayInfo, promptTokens int, ratio float64) (preConsumedQuota int64, billingSource string, err error) {
	preConsumedQuota = getPreConsumedQuota(promptTokens, 0, ratio)

	// ── 审计日志：入口 ──
	logger.Info(ctx, fmt.Sprintf("[BILLING_AUDIT] PreConsumeBilling 入口 | user=%d preQuota=%d promptTokens=%d ratio=%.4f",
		meta.UserID, preConsumedQuota, promptTokens, ratio))
	// ── 审计日志 END ──

	// ── Tier 1: 订阅 ──
	subRemain := getActiveSubscriptionRemaining(meta.UserID)
	if subRemain > 0 {
		if subRemain >= preConsumedQuota {
			_, err := model.PreConsumeUserSubscription(
				fmt.Sprintf("pc-%d-%d", meta.TokenID, promptTokens),
				meta.UserID, preConsumedQuota)
			if err == nil {
				// ── 审计日志：订阅扣款成功 ──
				logger.Info(ctx, fmt.Sprintf("[BILLING_AUDIT] PreConsumeBilling → subscription | user=%d subRemain=%d preQuota=%d",
					meta.UserID, subRemain, preConsumedQuota))
				// ── 审计日志 END ──
				return preConsumedQuota, "subscription", nil
			}
			logger.SysWarn(fmt.Sprintf("subscription pre-consume failed: %v", err))
		} else {
			_, err := model.PreConsumeUserSubscription(
				fmt.Sprintf("pc-sub-%d", meta.TokenID),
				meta.UserID, subRemain)
			if err == nil {
				preConsumedQuota -= subRemain
				logger.Info(ctx, fmt.Sprintf("[BILLING_AUDIT] PreConsumeBilling → subscription(partial) | user=%d subUsed=%d remainPreQuota=%d",
					meta.UserID, subRemain, preConsumedQuota))
			}
		}
	}

	// ── Tier 2: 现金余额 ──
	balance, err := model.GetUserCashBalance(meta.UserID)
	if err == nil && balance >= preConsumedQuota {
		// ── 审计日志：现金检查通过 ──
		logger.Info(ctx, fmt.Sprintf("[BILLING_AUDIT] PreConsumeBilling → cash | user=%d balance=%d preQuota=%d",
			meta.UserID, balance, preConsumedQuota))
		// ── 审计日志 END ──
		return preConsumedQuota, "cash", nil
	} else {
		logger.Warn(ctx, fmt.Sprintf("[BILLING_AUDIT] PreConsumeBilling Tier2跳过 | user=%d balance=%d preQuota=%d err=%v",
			meta.UserID, balance, preConsumedQuota, err))
	}

	// ── Tier 3: 佣金余额 ──
	user, err := model.GetUserById(meta.UserID, false)
	if err == nil && user.CommissionBalance >= preConsumedQuota {
		// ── 审计日志：佣金检查通过 ──
		logger.Info(ctx, fmt.Sprintf("[BILLING_AUDIT] PreConsumeBilling → commission | user=%d commission=%d preQuota=%d",
			meta.UserID, user.CommissionBalance, preConsumedQuota))
		// ── 审计日志 END ──
		return preConsumedQuota, "commission", nil
	} else {
		logger.Warn(ctx, fmt.Sprintf("[BILLING_AUDIT] PreConsumeBilling Tier3跳过 | user=%d commission=%d preQuota=%d err=%v",
			meta.UserID, user.CommissionBalance, preConsumedQuota, err))
	}

	// ── Tier 4: 配额回退 ──
	userQuota, err := model.CacheGetUserQuota(ctx, meta.UserID)
	if err != nil {
		logger.Warn(ctx, fmt.Sprintf("[BILLING_AUDIT] PreConsumeBilling Tier4失败(getQuota) | user=%d err=%v", meta.UserID, err))
		return 0, "", fmt.Errorf("get user quota: %w", err)
	}
	if userQuota < preConsumedQuota {
		logger.Warn(ctx, fmt.Sprintf("[BILLING_AUDIT] PreConsumeBilling Tier4失败(insufficient) | user=%d quota=%d preQuota=%d",
			meta.UserID, userQuota, preConsumedQuota))
		return 0, "", fmt.Errorf("insufficient balance")
	}
	if err := model.CacheDecreaseUserQuota(meta.UserID, preConsumedQuota); err != nil {
		logger.Warn(ctx, fmt.Sprintf("[BILLING_AUDIT] PreConsumeBilling Tier4失败(decrease) | user=%d err=%v", meta.UserID, err))
		return preConsumedQuota, "", fmt.Errorf("decrease user quota: %w", err)
	}
	if userQuota > 100*preConsumedQuota {
		preConsumedQuota = 0
	}
	if preConsumedQuota > 0 {
		if err := model.PreConsumeTokenQuota(meta.TokenID, preConsumedQuota); err != nil {
			return preConsumedQuota, "quota", fmt.Errorf("pre-consume token quota: %w", err)
		}
	}
	// ── 审计日志：配额扣款成功 ──
	logger.Info(ctx, fmt.Sprintf("[BILLING_AUDIT] PreConsumeBilling → quota | user=%d quota=%d", meta.UserID, userQuota))
	// ── 审计日志 END ──
	return preConsumedQuota, "quota", nil
}

// getActiveSubscriptionRemaining 获取用户活跃订阅的剩余额度
func getActiveSubscriptionRemaining(userId int) int64 {
	subs, err := model.GetAllActiveUserSubscriptions(userId)
	if err != nil || len(subs) == 0 {
		return 0
	}
	sub := subs[0].Subscription
	if sub == nil || sub.Status != "active" {
		return 0
	}
	return sub.AmountTotal - sub.AmountUsed
}

// PostConsumeBilling 统一后扣/退款
func PostConsumeBilling(ctx context.Context, billingSource string, meta *common.RelayInfo, tokenId int, finalQuota int64, preConsumedQuota int64) {
	quotaDelta := preConsumedQuota - finalQuota
	if quotaDelta > 0 {
		switch billingSource {
		case "subscription":
			// subscription refund handled internally
		case "cash":
			model.PlusUserCashBalance(meta.UserID, quotaDelta)
		case "commission":
			model.PlusUserCommissionBalance(meta.UserID, quotaDelta)
		case "quota":
			model.PostConsumeTokenQuota(tokenId, -quotaDelta)
			model.IncreaseUserQuota(meta.UserID, quotaDelta)
		}
	} else if quotaDelta < 0 {
		switch billingSource {
		case "cash":
			model.MinusUserCashBalance(meta.UserID, -quotaDelta)
		case "commission":
			model.MinusUserCommissionBalance(meta.UserID, -quotaDelta)
		case "quota":
			model.PostConsumeTokenQuota(tokenId, -quotaDelta)
			model.DecreaseUserQuota(meta.UserID, -quotaDelta)
		}
	}
}

// PreConsumeQuotaOld 兼容旧调用方
func PreConsumeQuotaOld(ctx context.Context, meta *common.RelayInfo, promptTokens int, ratio float64) (int64, error) {
	q, _, err := PreConsumeBilling(ctx, meta, promptTokens, ratio)
	return q, err
}

func PostConsumeQuotaOld(ctx context.Context, usage any, meta *common.RelayInfo, textRequest any, ratio float64, preConsumedQuota int64, modelRatio float64, groupRatio float64, systemPromptReset bool) {
	logger.SysLog("post-consume quota called (legacy stub)")
}

func getPreConsumedQuota(promptTokens int, maxTokens int, ratio float64) int64 {
	preConsumedTokens := config.PreConsumedQuota + int64(promptTokens)
	if maxTokens != 0 {
		preConsumedTokens += int64(maxTokens)
	}
	return int64(float64(preConsumedTokens) * ratio)
}
