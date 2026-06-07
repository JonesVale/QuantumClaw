package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/ctxkey"
	"github.com/quantumclaw/quantumclaw/model"
)

// GetConsumeRecords returns user's consumption history
func GetConsumeRecords(c *gin.Context) {
	userID := c.GetInt(ctxkey.Id)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize > 100 {
		pageSize = 100
	}
	records, total, err := model.GetUserConsumeRecords(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    records,
		"total":   total,
		"page":    page,
	})
}

// GetConsumeSummary returns user's spending summary
func GetConsumeSummary(c *gin.Context) {
	userID := c.GetInt(ctxkey.Id)
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days > 365 {
		days = 365
	}
	totalPrice, totalQuota, err := model.GetUserConsumeSummary(userID, days)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total_price_cents": totalPrice,
			"total_quota":       totalQuota,
			"days":              days,
		},
	})
}
