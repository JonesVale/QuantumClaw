package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/model"
)

// GetNotifications 获取当前用户的通知列表
func GetNotifications(c *gin.Context) {
	userId := c.GetInt("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	unreadOnly := c.DefaultQuery("unread_only", "false") == "true"

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	notifs, total, err := model.GetUserNotifications(userId, page, pageSize, unreadOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取通知失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"items":     notifs,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetUnreadNotificationCount 获取未读通知数量
func GetUnreadNotificationCount(c *gin.Context) {
	userId := c.GetInt("id")
	count, err := model.GetUnreadNotificationCount(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取未读通知数失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"unread_count": count,
		},
	})
}

// MarkNotificationRead 标记通知为已读
func MarkNotificationRead(c *gin.Context) {
	userId := c.GetInt("id")
	notifId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的通知ID",
		})
		return
	}

	if err := model.MarkNotificationRead(notifId, userId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "标记已读失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

// MarkAllNotificationsRead 标记所有通知为已读
func MarkAllNotificationsRead(c *gin.Context) {
	userId := c.GetInt("id")
	if err := model.MarkAllNotificationsRead(userId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "标记全部已读失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}
