package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/quantumclaw/quantumclaw/common/helper"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/service"
)

// GetReconciliations 获取对账记录列表（管理员）
// GET /admin/reconciliations?user_id=xxx&status=open&keyword=xxx&page=1&page_size=20
func GetReconciliations(c *gin.Context) {
	userIdStr := c.Query("user_id")
	status := c.Query("status")
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var userId int
	if userIdStr != "" {
		userId, _ = strconv.Atoi(userIdStr)
	}

	logs, total, err := model.ListReconciliationLogs(userId, status, keyword, page, pageSize)
	if err != nil {
		logger.Errorf(c.Request.Context(), "ListReconciliationLogs failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "查询对账记录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"data":     logs,
		"total":    total,
		"page":     page,
		"page_size": pageSize,
	})
}

// GetReconciliationDiscrepancies 获取对账不一致的记录（管理员）
// GET /admin/reconciliations/discrepancies?page=1&page_size=20
func GetReconciliationDiscrepancies(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	logs, total, err := service.ListReconciliationDiscrepancies(page, pageSize)
	if err != nil {
		logger.Errorf(c.Request.Context(), "ListReconciliationDiscrepancies failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "查询对账异常失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"data":     logs,
		"total":    total,
		"page":     page,
		"page_size": pageSize,
	})
}

// ResolveReconciliation 处理对账记录（管理员）
// POST /admin/reconciliations/:id/resolve
// Body: {"action": "approve_user|approve_channel|adjust_both|ignore", "remark": "..."}
func ResolveReconciliation(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid id"})
		return
	}

	type resolveReq struct {
		Action string `json:"action" binding:"required"`
		Remark string `json:"remark"`
	}
	var req resolveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid request"})
		return
	}

	switch req.Action {
	case "approve_user", "approve_channel", "adjust_both", "ignore":
		// valid
	default:
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "action must be approve_user/approve_channel/adjust_both/ignore"})
		return
	}

	// 获取对账记录
	var log model.ReconciliationLog
	if err := model.DB.First(&log, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "对账记录不存在"})
		return
	}

	if log.Status != model.ReconciliationStatusOpen {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "只有 open 状态的记录才能处理"})
		return
	}

	// 确定新状态
	newStatus := model.ReconciliationStatusResolved
	if req.Action == "ignore" {
		newStatus = model.ReconciliationStatusIgnored
	}

	// 获取操作人
	adminName, _ := c.Get("username")
	if adminName == nil {
		adminName, _ = c.Get("user_id")
	}
	adminStr := fmt.Sprint(adminName)

	// 更新对账记录
	now := helper.GetTimestamp()
	updates := map[string]interface{}{
		"status":      newStatus,
		"remark":      fmt.Sprintf("[%s] %s", req.Action, req.Remark),
		"resolved_by": adminStr,
		"resolved_at": now,
		"updated_at":  now,
	}
	if err := model.DB.Model(&model.ReconciliationLog{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		logger.Errorf(c.Request.Context(), "ResolveReconciliation update failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "处理失败"})
		return
	}

	logger.Infof(c.Request.Context(), "[RECONCILIATION] 处理对账 id=%d action=%s remark=%s admin=%s",
		id, req.Action, req.Remark, adminStr)

	// TODO: 根据 action 生成调整账单（补扣或退款）
	// - approve_user: 渠道成本正确，平台多收了，需要退还用户差额
	// - approve_channel: 用户扣款正确，渠道成本被低估，平台亏损
	// - adjust_both: 双方都有误差，需要手动调整
	// - ignore: 忽略此差异

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "处理成功",
	})
}
