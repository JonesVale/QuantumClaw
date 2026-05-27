package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/service"
)

// GetRouterConfig returns the current intelligent router configuration.
func GetRouterConfig(c *gin.Context) {
	cfg := service.GetRouterConfig()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    cfg,
	})
}

// UpdateRouterConfig updates the intelligent router configuration.
func UpdateRouterConfig(c *gin.Context) {
	var req service.RouterConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid request"})
		return
	}

	service.UpdateRouterConfig(req)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "router config updated"})
}

// GetRouterPerformance returns the current channel performance snapshot.
func GetRouterPerformance(c *gin.Context) {
	snapshot := model.GetChannelPerformanceSnapshot()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    snapshot,
	})
}

// ResetRouterPerformance resets all channel performance data.
func ResetRouterPerformance(c *gin.Context) {
	model.ResetChannelPerformance()
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "router performance data reset"})
}
