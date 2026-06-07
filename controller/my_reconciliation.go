package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/quantumclaw/quantumclaw/model"
)

// GetMyReconciliations 获取当前用户的对账记录（普通用户查看自己的对账明细）
// GET /api/reconciliation/my?page=1&page_size=20&status=open
func GetMyReconciliations(c *gin.Context) {
	// 从 Gin context 获取当前登录用户 ID（由 middleware.Auth() 设置）
	idValue, exists := c.Get("id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "未登录"})
		return
	}
	userId, ok := idValue.(int)
	if !ok || userId <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户信息异常"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status") // open / resolved / ignored / 留空=全部

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	logs, total, err := model.ListUserReconciliationLogs(userId, status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"items":      logs,
			"total":      total,
			"page":       page,
			"page_size":  pageSize,
			"total_pages": totalPages,
		},
	})
}
