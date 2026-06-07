package service

import (
	"fmt"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
)

// CalculateAllStoreFees ??????????????????
func CalculateAllStoreFees(year int, month time.Month) {
	stores, err := model.GetActiveStores()
	if err != nil {
		logger.SysError("calculate store fees: " + err.Error())
		return
	}

	startOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.Local).Unix()
	endOfMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, time.Local).Unix()
	period := fmt.Sprintf("%04d-%02d", year, int(month))

	for _, store := range stores {
		// ?????????
		var totalRevenue int64
		err := model.DB.Model(&model.ProviderEarning{}).
			Where("user_id = ? AND created_at >= ? AND created_at < ? AND status = ?",
				store.UserID, startOfMonth, endOfMonth, model.EarningStatusSettled).
			Select("COALESCE(SUM(net_amount), 0)").Scan(&totalRevenue).Error
		if err != nil {
			logger.SysErrorf("query store %d revenue: %v", store.ID, err)
			continue
		}

		// ??????
		cfg, err := model.GetFeeConfig(store.Tier)
		if err != nil {
			logger.SysErrorf("get fee config for store %d tier %s: %v", store.ID, store.Tier, err)
			continue
		}

		// ?????? ? ??
		if totalRevenue < cfg.MinSkip {
			_ = model.CreatePlatformFeeRecord(store.ID, store.UserID, period, totalRevenue, 0, 0, model.PlatformFeeStatusSkipped)
			continue
		}

		// ????
		feeAmount := int64(float64(totalRevenue) * cfg.Rate / 100.0)
		_ = model.CreatePlatformFeeRecord(store.ID, store.UserID, period, totalRevenue, cfg.Rate, feeAmount, model.PlatformFeeStatusPending)

		// ?????
		changed, _ := model.AutoUpgradeStoreTier(store.ID, store.TotalSales)
		if changed {
			logger.SysLogf("store %d (%s) auto upgraded to %s", store.ID, store.Name, store.Tier)
		}
	}
}
