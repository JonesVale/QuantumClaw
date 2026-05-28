package service

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
)

// safeRunSettlement 安全执行对账，panic 时不崩溃整个定时器
func safeRunSettlement() {
	defer func() {
		if r := recover(); r != nil {
			logger.Error(nil, fmt.Sprintf("[HourlySettlement] PANIC RECOVERED: %v\n%s", r, debug.Stack()))
		}
	}()
	model.RunHourlySettlement()
}

// StartHourlySettlement 启动每小时对账定时任务
func StartHourlySettlement() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	// 启动后先对上一小时
	safeRunSettlement()

	for range ticker.C {
		safeRunSettlement()
	}
}

// DeductDebtOnTopup 已删除——债务抵扣逻辑已移至 CompleteTopUp 事务内
// （2026-05-28 重构：TOCTOU 修复 + 原子操作）
