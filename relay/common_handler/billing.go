package common_handler

import (
	"context"
	"fmt"

	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/relay/common"
)

func PreConsumeQuota(ctx context.Context, meta *common.RelayInfo, promptTokens int, ratio float64) (int64, error) {
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

func PostConsumeQuota(ctx context.Context, usage any, meta *common.RelayInfo, textRequest any, ratio float64, preConsumedQuota int64, modelRatio float64, groupRatio float64, systemPromptReset bool) {
	// Stub - actual implementation will use BillingSession
	logger.SysLog("post-consume quota called")
}

func getPreConsumedQuota(promptTokens int, maxTokens int, ratio float64) int64 {
	preConsumedTokens := config.PreConsumedQuota + int64(promptTokens)
	if maxTokens != 0 {
		preConsumedTokens += int64(maxTokens)
	}
	return int64(float64(preConsumedTokens) * ratio)
}
