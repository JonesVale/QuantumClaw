package controller

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/setting/operation_setting"
	"github.com/gin-gonic/gin"
)

// GetCheckinStatus 获取用户签到状态和历史记录
func GetCheckinStatus(c *gin.Context) {
	setting := operation_setting.GetCheckinSetting()
	if !setting.Enabled {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "签到功能未启用"})
		return
	}
	userId := c.GetInt("id")
	month := c.DefaultQuery("month", time.Now().Format("2006-01"))

	stats, err := model.GetUserCheckinStats(userId, month)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"enabled":   setting.Enabled,
			"min_quota": setting.MinQuota,
			"max_quota": setting.MaxQuota,
			"stats":     stats,
		},
	})
}

// DoCheckin 执行用户签到
func DoCheckin(c *gin.Context) {
	setting := operation_setting.GetCheckinSetting()
	if !setting.Enabled {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "签到功能未启用"})
		return
	}

	userId := c.GetInt("id")

	checkin, err := model.UserCheckin(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	model.RecordLog(context.Background(), userId, model.LogTypeSystem, fmt.Sprintf("用户签到，获得额度 %d", checkin.QuotaAwarded))
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "签到成功",
		"data": gin.H{
			"quota_awarded": checkin.QuotaAwarded,
			"checkin_date":  checkin.CheckinDate,
		},
	})
}

// GetCheckinHistory 获取签到历史记录
func GetCheckinHistory(c *gin.Context) {
	userId := c.GetInt("id")
	month := c.DefaultQuery("month", time.Now().Format("2006-01"))
	startDate := month + "-01"
	endDate := month + "-31"

	records, err := model.GetUserCheckinRecords(userId, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	result := make([]gin.H, 0, len(records))
	for _, r := range records {
		result = append(result, gin.H{
			"id":         r.Id,
			"user_id":    r.UserId,
			"date":       r.CheckinDate,
			"amount":     r.QuotaAwarded,
			"created_at": r.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}
