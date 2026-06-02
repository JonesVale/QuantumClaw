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
// PreConsumeQuota 消费前检查配额（旧版，走 quota）
// 优先级链：有订阅 → 先扣订阅额度 → 订阅扣完 → 扣 CashBalance → CashBalance 用完 → 扣 Quota
// 链路使用方：relay/controller/helper.go → preConsumeBalance → service.PreConsumeBalance
// 本文件是 relay 层统一入口，上层调用 PreConsumeBilling 即可

// PreConsumeBilling 统一预扣入口（优先级链）
// 返回值：preConsumedQuota（预扣量），billingSource（实际扣款来源），error
// billingSource: "subscription" | "cash" | "quota"
func PreConsumeBilling(ctx context.Context, meta *common.RelayInfo, promptTokens int, ratio float64) (preConsumedQuota int64, billingSource string, err error) {
	preConsumedQuota = getPreConsumedQuota(promptTokens, 0, ratio)

	// ── Tier 1: 检查订阅 ──
	sub, err := model.GetActiveUserSubscription(meta.UserID)
	if err == nil && sub != nil && sub.RemainQuota > 0 {
		if sub.RemainQuota >= preConsumedQuota {
			// 订阅额度足够
			_, err := model.PreConsumeUserSubscription(
				fmt.Sprintf("pc-%d-%d", meta.TokenID, promptTokens),
				meta.UserID, preConsumedQuota)
			if err == nil {
				return preConsumedQuota, "subscription", nil
			}
			logger.SysWarn(fmt.Sprintf("subscription pre-consume failed, fallback: %v", err))
		} else {
			// 订阅额度不足，先扣完订阅剩下的走现金
			_, err := model.PreConsumeUserSubscription(
				fmt.Sprintf("pc-sub-%d", meta.TokenID),
				meta.UserID, sub.RemainQuota)
			if err == nil {
				// 只扣了部分，剩余走现金/配额
				preConsumedQuota -= sub.RemainQuota
			}
		}
	}

	// ── Tier 2: 现金余额 ──
	balance, err := model.GetUserCashBalance(meta.UserID)
	if err == nil && balance >= preConsumedQuota {
		return preConsumedQuota, "cash", nil
	}

	// ── Tier 3: 配额回退 ──
	userQuota, err := model.CacheGetUserQuota(ctx, meta.UserID)
	if err != nil {
		return 0, "", fmt.Errorf("get user quota: %w", err)
	}
	if userQuota < preConsumedQuota {
		return 0, "", fmt.Errorf("insufficient balance (sub=%d cash=%d quota=%d need=%d)",
			getSubRemainQuota(sub), balance, userQuota, preConsumedQuota)
	}
	// 预扣配额
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

func getSubRemainQuota(sub *model.UserSubscription) int64 {
	if sub == nil {
		return 0
	}
	return sub.RemainQuota
}

// PostConsumeBilling 统一后扣入口
func PostConsumeBilling(ctx context.Context, billingSource string, meta *common.RelayInfo, tokenId int, finalQuota int64, preConsumedQuota int64) {
	quotaDelta := preConsumedQuota - finalQuota
	if quotaDelta > 0 {
		// 退还多扣的
		switch billingSource {
		case "subscription":
			model.ReturnPreConsumedUserSubscription(tokenId, quotaDelta)
		case "cash":
			model.IncreaseUserCashBalance(meta.UserID, quotaDelta)
		case "quota":
			model.ReturnPreConsumedTokenQuota(tokenId, quotaDelta)
			model.IncreaseUserQuota(meta.UserID, quotaDelta)
		}
	} else if quotaDelta < 0 {
		// 需要补扣
		switch billingSource {
		case "cash":
			model.MinusUserCashBalance(meta.UserID, -quotaDelta)
		case "quota":
			model.PostConsumeTokenQuota(tokenId, -quotaDelta)
			model.DecreaseUserQuota(meta.UserID, -quotaDelta)
		}
		// subscription post-consume handled in subscription model
	}
}

// 以下为原有函数（兼容旧调用方）
func PreConsumeQuotaOld(ctx context.Context, meta *common.RelayInfo, promptTokens int, ratio float64) (int64, error) {
	preConsumedQuota := getPreConsumedQuota(promptTokens, 0, ratio)
	userQuota, err := model.CacheGetUserQuota(ctx, meta.UserID)
	if err != nil {
		return 0, fmt.Errorf("get user quota: %w", err)
	}
	if userQuota-preConsumedQuota < 0 {
		return 0, fmt.Errorf("insufficient user quota")
	}
	err = model.CacheDecreaseUserQuota(meta.UserID, preConsumedQuota)
	if err != nil {
		return preConsumedQuota, fmt.Errorf("decrease user quota: %w", err)
	}
	if userQuota > 100*preConsumedQuota {
		preConsumedQuota = 0
	}
	if preConsumedQuota > 0 {
		err := model.PreConsumeTokenQuota(meta.TokenID, preConsumedQuota)
		if err != nil {
			return preConsumedQuota, fmt.Errorf("pre-consume token quota: %w", err)
		}
	}
	return preConsumedQuota, nil
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
