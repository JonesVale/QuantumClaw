package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/model"
)

func GetHealth(c *gin.Context) {
	dbOK := "ok"
	if sqlDB, err := model.DB.DB(); err == nil {
		if sqlDB.Ping() != nil {
			dbOK = "down"
		}
	} else {
		dbOK = "down"
	}

	var activeChannels int64
	model.DB.Model(&model.Channel{}).Where("status = ?", model.ChannelStatusEnabled).Count(&activeChannels)

	var activeStores int64
	model.DB.Model(&model.Store{}).Where("status = ?", model.StoreStatusActive).Count(&activeStores)

	var pendingFeeTotal int64
	model.DB.Model(&model.PlatformFeeRecord{}).Where("status = ?", model.PlatformFeeStatusPending).
		Select("COALESCE(SUM(fee_amount), 0)").Scan(&pendingFeeTotal)

	c.JSON(http.StatusOK, gin.H{
		"version":          common.Version,
		"start_time":       common.StartTime,
		"db":               dbOK,
		"channels_active":  activeChannels,
		"stores_active":    activeStores,
		"pending_fee_total": pendingFeeTotal,
		"debug_enabled":    config.DebugEnabled,
	})
}
