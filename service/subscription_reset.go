package service

import (
	"fmt"
	"time"

	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
)

// StartSubscriptionQuotaResetTask periodically checks and resets
// subscription quotas that are due for reset (daily/weekly/monthly/custom period).
// Runs every 10 minutes.
func StartSubscriptionQuotaResetTask() {
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		// Run once at startup
		resetDueSubscriptions()

		for range ticker.C {
			resetDueSubscriptions()
		}
	}()
	logger.SysLog("subscription quota reset task started")
}

func resetDueSubscriptions() {
	defer common.RecoverAndLog()

	count, err := model.ResetDueSubscriptions(200)
	if err != nil {
		logger.SysError("failed to reset due subscriptions: " + err.Error())
		return
	}

	expired, err := model.ExpireDueSubscriptions(200)
	if err != nil {
		logger.SysError("failed to expire due subscriptions: " + err.Error())
	}

	if count > 0 || expired > 0 {
		logger.SysLog(fmt.Sprintf("subscription maintenance: %d reset, %d expired", count, expired))
	}
}
