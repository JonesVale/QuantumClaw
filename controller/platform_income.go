package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/model"
)

// GetPlatformIncomeSummary 获取平台收入汇总
// GET /api/platform/income/summary
//
// 返回:
//   - today_income:    今日平台抽成（分）
//   - month_income:    本月平台抽成（分）
//   - balance:         当前可提现余额（分）
//   - total_income:    历史累计总收入（分）
func GetPlatformIncomeSummary(c *gin.Context) {
	today, thisMonth, balance, totalIncome := model.GetPlatformIncomeSummary()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "ok",
		"data": gin.H{
			"today_income": today,
			"month_income": thisMonth,
			"balance":      balance,
			"total_income": totalIncome,
			// 方便前端展示：转元（保留2位小数）
			"today_income_yuan":    float64(today) / 100.0,
			"month_income_yuan":   float64(thisMonth) / 100.0,
			"balance_yuan":        float64(balance) / 100.0,
			"total_income_yuan":   float64(totalIncome) / 100.0,
		},
	})
}

// GetPlatformIncomeHistory 获取平台收入流水
// GET /api/platform/income/history?page=1&page_size=20
func GetPlatformIncomeHistory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	list, total, err := model.GetPlatformIncomeHistory(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to query platform income history",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "ok",
		"data": gin.H{
			"list":  list,
			"total": total,
			"page":  page,
			"page_size": pageSize,
		},
	})
}
