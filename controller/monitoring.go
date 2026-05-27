package controller

import (
	"net/http"
	"time"

	"github.com/quantumclaw/quantumclaw/model"
	"github.com/gin-gonic/gin"
)

// AdminMonitorOverview provides a comprehensive admin monitoring snapshot.
type AdminMonitorOverview struct {
	Channels struct {
		Total   int `json:"total"`
		Active  int `json:"active"`
		AutoDisabled int `json:"auto_disabled"`
		ManualDisabled int `json:"manual_disabled"`
		BalanceLow int `json:"balance_low"`
	} `json:"channels"`

	Requests struct {
		TodayTotal    int   `json:"today_total"`
		TodayTokens   int64 `json:"today_tokens"`
		AvgLatencyMs  int64 `json:"avg_latency_ms"`
		SuccessRate   float64 `json:"success_rate"`
	} `json:"requests"`

	Platform struct {
		TotalUsers    int64 `json:"total_users"`
		ActiveTokens  int64 `json:"active_tokens"`
		ActiveToday   int64 `json:"active_today"`
		ChannelsCount int64 `json:"channels_count"`
	} `json:"platform"`

	ModelTop struct {
		ModelName string `json:"model_name"`
		Count      int    `json:"count"`
	} `json:"model_top"`

	Timestamp int64 `json:"timestamp"`
}

func GetAdminMonitor(c *gin.Context) {
	overview := AdminMonitorOverview{}
	overview.Timestamp = time.Now().UnixMilli()
	db := model.DB
	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()

	// ── Channel stats ──
	if db != nil {
		// Total channels
		var totalChans int64
		db.Model(&model.Channel{}).Count(&totalChans)
		overview.Platform.ChannelsCount = totalChans

		// Active channels
		var activeCount int64
		db.Model(&model.Channel{}).Where("status = 1").Count(&activeCount)
		overview.Channels.Active = int(activeCount)

		// Manual disabled
		var manual int64
		db.Model(&model.Channel{}).Where("status = 2").Count(&manual)
		overview.Channels.ManualDisabled = int(manual)

		// Auto disabled
		var auto int64
		db.Model(&model.Channel{}).Where("status = 3").Count(&auto)
		overview.Channels.AutoDisabled = int(auto)

		overview.Channels.Total = int(totalChans)
	}

	// ── Request stats (last 24h) ──
	if db != nil {
		// Today's total requests
		var todayCount int64
		db.Model(&model.Log{}).Where("created_at >= ?", startOfToday).Count(&todayCount)
		overview.Requests.TodayTotal = int(todayCount)

		// Today's total tokens
		var todayTokens struct{ Sum int64 }
		db.Model(&model.Log{}).Select("COALESCE(SUM(prompt_tokens + completion_tokens), 0) as sum").
			Where("created_at >= ?", startOfToday).Scan(&todayTokens)
		overview.Requests.TodayTokens = todayTokens.Sum

		// Platform stats
		var totalUsers int64
		db.Model(&model.User{}).Count(&totalUsers)
		overview.Platform.TotalUsers = totalUsers

		var totalTokens int64
		db.Model(&model.Token{}).Count(&totalTokens)
		overview.Platform.ActiveTokens = totalTokens

		var activeUsersToday int64
		db.Model(&model.Log{}).Select("COUNT(DISTINCT user_id)").
			Where("created_at >= ?", startOfToday).Scan(&activeUsersToday)
		overview.Platform.ActiveToday = activeUsersToday

		// Top model
		type TopModel struct {
			ModelName string `json:"model_name"`
			Cnt       int    `json:"cnt"`
		}
		var top TopModel
		db.Model(&model.Log{}).Select("model_name, COUNT(*) as cnt").
			Where("created_at >= ?", startOfToday).
			Group("model_name").Order("cnt desc").Limit(1).Scan(&top)
		overview.ModelTop.ModelName = top.ModelName
		overview.ModelTop.Count = top.Cnt
	}

	// ── Channel latency & success (last 24h) ──
	if db != nil {
		type LatencyRow struct {
			AvgMs  float64
			ErrCnt int64
			Total  int64
		}
		var row LatencyRow
		if err := db.Raw(`SELECT
			COALESCE(AVG(elapsed_time), 0) as avg_ms,
			COALESCE(SUM(CASE WHEN quota = 0 THEN 1 ELSE 0 END), 0) as err_cnt,
			COUNT(*) as total
			FROM logs WHERE created_at >= ?`, startOfToday).Scan(&row).Error; err == nil {
		overview.Requests.AvgLatencyMs = int64(row.AvgMs)
		if row.Total > 0 {
			overview.Requests.SuccessRate = float64(row.Total-row.ErrCnt) / float64(row.Total) * 100.0
		} else {
			overview.Requests.SuccessRate = 100.0
		}
	}

		// Channels with low balance (< 10 USD)
		var lowBal int64
		db.Model(&model.Channel{}).Where("balance > 0 AND balance < 10").Count(&lowBal)
		overview.Channels.BalanceLow = int(lowBal)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    overview,
	})
}
