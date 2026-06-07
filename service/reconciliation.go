package service

import (
	"context"
	"fmt"
	"math"

	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
)

// TriggerReconciliation 在每次 PostConsumeDeduct 完成后异步调用
// 写入 reconciliation_logs 供后续审计
func TriggerReconciliation(ctx context.Context, userId int, channelId int, consumeLogId int, userDeductedCents int64, platformIncomeCents int64) {
	// 异步执行，不阻塞主请求
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error(context.Background(), fmt.Sprintf("[RECONCILIATION] panic: %v", r))
			}
		}()

		reconcileCtx := context.Background()

		// 1. 获取渠道信息，估算上游成本
		channel, err := model.GetChannelById(channelId, true)
		if err != nil {
			logger.Warnf(reconcileCtx, "[RECONCILIATION] get channel %d failed: %v", channelId, err)
			return
		}

		// 2. 估算上游成本（分）
		var channelCostCents float64
		if channel.CostPrice > 0 {
			// CostPrice 单位是 USD/1K-tokens
			// 此处无法知道实际 token 数，使用 userDeductedCents 反推
			// 这是一个保守估计：假设用户付了 userDeductedCents 分，上游成本约为 userDeductedCents * (CostPrice / SellPrice)
			// 简化：先记录，后续人工对账
			channelCostCents = float64(userDeductedCents) * 0.7 // 假设成本约为用户付费的 70%（可调整）
		} else if channel.CostPerUnit > 0 {
			channelCostCents = float64(userDeductedCents) * channel.CostPerUnit / 10000.0
		}

		// 3. 计算差额
		diffCents := float64(userDeductedCents) - channelCostCents - float64(platformIncomeCents)
		status := model.ReconciliationStatusResolved
		if math.Abs(diffCents) > 1.0 { // 差额超过 1 分，认为不一致
			status = model.ReconciliationStatusOpen
		}

		// 4. 写入对账记录
		log := &model.ReconciliationLog{
			UserId:              userId,
			ChannelId:           channelId,
			ConsumeLogId:       consumeLogId,
			UserDeductedCents: userDeductedCents,
			ChannelCostCents:   int64(math.Round(channelCostCents)),
			PlatformIncomeCents: platformIncomeCents,
			Status:              status,
			DiffCents:          int64(math.Round(diffCents)),
			Remark:              fmt.Sprintf("auto-triggered consume_log_id=%d", consumeLogId),
		}

		if err := model.CreateReconciliationLog(log); err != nil {
			logger.Errorf(reconcileCtx, "[RECONCILIATION] create log failed: %v", err)
			return
		}

		if status == model.ReconciliationStatusOpen {
			logger.Errorf(reconcileCtx, "[RECONCILIATION] ⚠️ 对账不一致! user=%d channel=%d diff=%.2f分 consume_log=%d",
				userId, channelId, diffCents, consumeLogId)
			// TODO: 发送告警（邮件/钉钉/企业微信）
		} else {
			logger.Infof(reconcileCtx, "[RECONCILIATION] 对账一致 user=%d channel=%d", userId, channelId)
		}
	}()
}

// ReconcileConsumeLog 对单条 consume_log 执行对账（供 API 手动触发）
func ReconcileConsumeLog(ctx context.Context, consumeLogId int) error {
	// TODO: 实现单条对账逻辑
	// 1. 从 consume_logs 表读取记录
	// 2. 从 provider_earnings 表读取渠道商收益
	// 3. 从 platform_income 表读取平台收入
	// 4. 计算差额，写入 reconciliation_logs
	return nil
}

// ListReconciliationDiscrepancies 列出所有对账不一致的记录（供后台 API 使用）
func ListReconciliationDiscrepancies(page int, pageSize int) ([]model.ReconciliationLog, int64, error) {
	return model.ListReconciliationLogs(0, model.ReconciliationStatusOpen, "", page, pageSize)
}
