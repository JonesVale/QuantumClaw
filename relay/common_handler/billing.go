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

	// ── Tier 1: 订阅 ──
	subRemain := getActiveSubscriptionRemaining(meta.UserID)
	if subRemain > 0 {
		if subRemain >= preConsumedQuota {
			_, err := model.PreConsumeUserSubscription(
				fmt.Sprintf("pc-%d-%d", meta.TokenID, promptTokens),
				meta.UserID, preConsumedQuota)
			if err == nil {
				return preConsumedQuota, "subscription", nil
			}
			logger.SysWarn(fmt.Sprintf("subscription pre-consume failed: %v", err))
		} else {
			_, err := model.PreConsumeUserSubscription(
				fmt.Sprintf("pc-sub-%d", meta.TokenID),
				meta.UserID, subRemain)
			if err == nil {
				preConsumedQuota -= subRemain
			}
		}
	}

	// ── Tier 2: 现金余额 ──
	balance, err := model.GetUserCashBalance(meta.UserID)
	if err == nil && balance >= preConsumedQuota {
		return preConsumedQuota, "cash", nil
	}

	// ── Tier 3: 佣金余额 ──
	user, err := model.GetUserById(meta.UserID, false)
	if err == nil && user.CommissionBalance >= preConsumedQuota {
		return preConsumedQuota, "commission", nil
	}

	// ── Tier 4: 配额回退 ──
	userQuota, err := model.CacheGetUserQuota(ctx, meta.UserID)
	if err != nil {
		return 0, "", fmt.Errorf("get user quota: %w", err)
	}
	if userQuota < preConsumedQuota {
		return 0, "", fmt.Errorf("insufficient balance")
	}
	if err := model.CacheDecreaseUserQuota(meta.UserID, preConsumedQuota); err != nil {
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
