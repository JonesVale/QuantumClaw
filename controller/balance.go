package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/quantumclaw/quantumclaw/common/ctxkey"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
)

// AdminAddBalanceRequest 管理员加余额请求
type AdminAddBalanceRequest struct {
	UserId int   `json:"user_id" binding:"required"`
	Amount int64 `json:"amount" binding:"required,min=1"` // 金额，单位：分
	Remark string `json:"remark"`
}

// AdminAddBalance 管理员为用户手动加余额
func AdminAddBalance(c *gin.Context) {
	var req AdminAddBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误：" + err.Error()})
		return
	}

	user, err := model.GetUserById(req.UserId, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "用户不存在"})
		return
	}

	if err := model.PlusUserCashBalance(user.Id, req.Amount); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "加余额失败：" + err.Error()})
		return
	}

	// 记余额流水
	newBalance, _ := model.GetUserCashBalance(user.Id)
	remark := req.Remark
	if remark == "" {
		remark = fmt.Sprintf("管理员加余额 %d 分", req.Amount)
	}
	_ = model.CreateBalanceLog(user.Id, model.BalanceLogTypeAdmin, req.Amount, newBalance, 0, remark)

	logger.Info(c.Request.Context(), fmt.Sprintf("管理员为用户 %d(%s) 加余额 %d 分，当前余额 %d",
		user.Id, user.Username, req.Amount, newBalance))

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": gin.H{
		"user_id":      user.Id,
		"amount":       req.Amount,
		"new_balance":  newBalance,
	}})
}

// GetUserBalance 获取用户余额信息
func GetUserBalance(c *gin.Context) {
	userId := c.GetInt("id")
	if userId <= 0 {
		// 管理员可以查指定用户
		if idStr := c.Query("user_id"); idStr != "" {
			if id, err := strconv.Atoi(idStr); err == nil && id > 0 {
				userId = id
			}
		}
	}

	balance, err := model.GetUserCashBalance(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "查询失败"})
		return
	}

	// 查询最近流水
	logs, _ := model.GetUserBalanceLogs(userId, 20)

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
		"balance":      balance,
		"balance_yuan": float64(balance) / 100.0,
		"logs":         logs,
	}})
}

// GetSelfBalance 获取当前用户余额
func GetSelfBalance(c *gin.Context) {
	userId := c.GetInt("id")
	if userId <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "未登录"})
		return
	}

	balance, err := model.GetUserCashBalance(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "查询失败"})
		return
	}

	logs, _ := model.GetUserBalanceLogs(userId, 20)

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
		"balance":      balance,
		"balance_yuan": float64(balance) / 100.0,
		"logs":         logs,
	}})
}

// GetUserBalanceByAdmin 管理员查看用户余额
func GetUserBalanceByAdmin(c *gin.Context) {
	idStr := c.Param("id")
	userId, err := strconv.Atoi(idStr)
	if err != nil || userId <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的用户ID"})
		return
	}

	balance, err := model.GetUserCashBalance(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "查询失败"})
		return
	}

	logs, _ := model.GetUserBalanceLogs(userId, 50)

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
		"user_id":      userId,
		"balance":      balance,
		"balance_yuan": float64(balance) / 100.0,
		"logs":         logs,
	}})
}
// ==================== Admin Audit Log APIs ====================

// AdminGetAuditLogs 管理员查询所有用户的余额流水（分页）
func AdminGetAuditLogs(c *gin.Context) {
	userIdStr := c.Query("user_id")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	userId, _ := strconv.Atoi(userIdStr)
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	logs, total, err := model.GetAllBalanceLogs(userId, page, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "查询失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"logs":      logs,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// AdminRollbackBalanceLog 管理员回滚一笔余额交易
func AdminRollbackBalanceLog(c *gin.Context) {
	type RollbackReq struct {
		LogId  int    `json:"log_id" binding:"required"`
		Remark string `json:"remark"`
	}
	var req RollbackReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误: " + err.Error()})
		return
	}

	operatorId := c.GetInt(ctxkey.Id)
	if operatorId <= 0 {
		operatorId = 0
	}

	rollbackLog, err := model.RollbackBalanceLog(req.LogId, operatorId, req.Remark)
	if err != nil {
		logger.Error(c.Request.Context(), fmt.Sprintf("rollback log %d failed: %v", req.LogId, err))
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "回滚失败: " + err.Error()})
		return
	}

	logger.Info(c.Request.Context(),
		fmt.Sprintf("管理员回滚交易 #%d (operator=%d) 用户=%d 金额=%d分",
			req.LogId, operatorId, rollbackLog.UserId, rollbackLog.Amount))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "回滚成功",
		"data":    rollbackLog,
	})
}

