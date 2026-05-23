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
