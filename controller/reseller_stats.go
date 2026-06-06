package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/model"
)

// GetResellerStats 渠道商统计 API
// GET /api/reseller/stats?days=7
func GetResellerStats(c *gin.Context) {
	userId := c.GetInt("id")
	daysStr := c.DefaultQuery("days", "7")
	daysInt := 7
	if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d < 365 {
		daysInt = d
	}
	since := time.Now().AddDate(0, 0, -daysInt).Unix()

	// 查询所有相关交易
	var txns []model.TokenTransaction
	model.DB.Where("(promoter_id = ? OR channel_owner_id = ?) AND created_time >= ?", userId, userId, since).
		Order("created_time ASC").
		Find(&txns)

	// 按天聚合
	type DayBucket struct {
		Total   float64
		Comm    float64
		Cost    float64
		Count   int
	}
	dayMap := make(map[string]*DayBucket)
	modelMap := make(map[string]*DayBucket)

	for _, t := range txns {
		dateStr := time.Unix(t.CreatedTime, 0).Format("2006-01-02")
		if _, ok := dayMap[dateStr]; !ok {
			dayMap[dateStr] = &DayBucket{}
		}
		dayMap[dateStr].Total += t.TotalAmount
		dayMap[dateStr].Comm += t.CommissionAmount
		dayMap[dateStr].Cost += t.UnifiedCost
		dayMap[dateStr].Count++

		// By model
		if _, ok := modelMap[t.ModelName]; !ok {
			modelMap[t.ModelName] = &DayBucket{}
		}
		modelMap[t.ModelName].Total += t.TotalAmount
		modelMap[t.ModelName].Count++
	}

	// Build response
	type DailyItem struct {
		Date         string  `json:"date"`
		TotalAmount  float64 `json:"total_amount"`
		Commission   float64 `json:"commission"`
		UnifiedCost  float64 `json:"unified_cost"`
		RequestCount int     `json:"request_count"`
	}
	var daily []DailyItem
	for date, b := range dayMap {
		daily = append(daily, DailyItem{
			Date: date, TotalAmount: b.Total, Commission: b.Comm,
			UnifiedCost: b.Cost, RequestCount: b.Count,
		})
	}

	type ModelItem struct {
		ModelName    string  `json:"model_name"`
		TotalAmount  float64 `json:"total_amount"`
		RequestCount int     `json:"request_count"`
	}
	var byModel []ModelItem
	for name, b := range modelMap {
		byModel = append(byModel, ModelItem{
			ModelName: name, TotalAmount: b.Total, RequestCount: b.Count,
		})
	}

	// Fallback stats
	var fbTotal, fbAmount float64
	for _, t := range txns {
		if t.IsFallback == 1 && t.PromoterId == userId {
			fbTotal++
			fbAmount += t.TotalAmount
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"daily":     daily,
			"by_model":  byModel,
			"fallbacks": gin.H{"total_fallbacks": fbTotal, "fallback_amount": fbAmount},
		},
	})
}

// GetProviderCustomers 渠道商的客户分析
// GET /api/provider/customers?days=30
func GetProviderCustomers(c *gin.Context) {
	userId := c.GetInt("id")
	daysStr := c.DefaultQuery("days", "30")
	daysInt := 30
	if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d < 365 {
		daysInt = d
	}
	since := time.Now().AddDate(0, 0, -daysInt).Unix()

	// 按客户聚合
	type CustomerStat struct {
		UserId       int     `json:"user_id"`
		Username     string  `json:"username"`
		TotalAmount  float64 `json:"total_amount"`
		Commission   float64 `json:"commission"`
		RequestCount int64   `json:"request_count"`
		LastActive   int64   `json:"last_active"`
	}
	var stats []CustomerStat
	model.DB.Raw(`
		SELECT 
			t.user_id,
			COALESCE(u.username, 'unknown') as username,
			SUM(t.total_amount) as total_amount,
			SUM(t.commission_amount) as commission,
			COUNT(*) as request_count,
			MAX(t.created_time) as last_active
		FROM token_transactions t
		LEFT JOIN users u ON u.id = t.user_id
		WHERE t.channel_owner_id = ? AND t.created_time >= ?
		GROUP BY t.user_id
		ORDER BY total_amount DESC
		LIMIT 50
	`, userId, since).Scan(&stats)
	if stats == nil {
		stats = []CustomerStat{}
	}

	// 按渠道聚合
	type ChannelStat struct {
		ChannelId    int     `json:"channel_id"`
		ChannelName  string  `json:"channel_name"`
		TotalAmount  float64 `json:"total_amount"`
		RequestCount int64   `json:"request_count"`
	}
	var channelStats []ChannelStat
	model.DB.Raw(`
		SELECT 
			t.channel_id,
			COALESCE(c.name, 'unknown') as channel_name,
			SUM(t.total_amount) as total_amount,
			COUNT(*) as request_count
		FROM token_transactions t
		LEFT JOIN channels c ON c.id = t.channel_id
		WHERE t.channel_owner_id = ? AND t.created_time >= ?
		GROUP BY t.channel_id
		ORDER BY total_amount DESC
	`, userId, since).Scan(&channelStats)
	if channelStats == nil {
		channelStats = []ChannelStat{}
	}

	// 按部门聚合（Provider的内部团队业绩）
	type DeptTeamStat struct {
		DepartmentId int     `json:"department_id"`
		DeptName     string  `json:"dept_name"`
		TotalAmount  float64 `json:"total_amount"`
		Commission   float64 `json:"commission"`
		RequestCount int64   `json:"request_count"`
		MemberCount  int64   `json:"member_count"`
	}
	var deptStats []DeptTeamStat
	model.DB.Raw(`
		SELECT
			COALESCE(u.department_id, 0) as department_id,
			COALESCE(d.name, '未分配') as dept_name,
			SUM(t.total_amount) as total_amount,
			SUM(t.commission_amount) as commission,
			COUNT(*) as request_count,
			COUNT(DISTINCT u.id) as member_count
		FROM token_transactions t
		JOIN users u ON u.id = t.channel_owner_id
		LEFT JOIN departments d ON d.id = u.department_id
		WHERE t.channel_owner_id = ? AND t.created_time >= ?
		GROUP BY u.department_id
		ORDER BY total_amount DESC
	`, userId, since).Scan(&deptStats)
	if deptStats == nil {
		deptStats = []DeptTeamStat{}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"customers":  stats,
			"channels":   channelStats,
			"dept_teams": deptStats,
		},
	})
}
