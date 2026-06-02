package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/ctxkey"
	"github.com/quantumclaw/quantumclaw/model"
)

// GetCommissionSetting — 获取佣金配置（管理员）
func GetCommissionSetting(c *gin.Context) {
	setting, err := model.GetCommissionSetting()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": setting})
}

// SaveCommissionSetting — 保存佣金配置（管理员）
func SaveCommissionSetting(c *gin.Context) {
	var s model.CommissionSetting
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid request"})
		return
	}
	if err := model.SaveCommissionSetting(&s); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "saved"})
}

// GetMyCommissionRecords — 获取我的佣金记录
func GetMyCommissionRecords(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit > 100 {
		limit = 100
	}
	records, total, err := model.GetUserCommissionRecords(userId, limit, offset)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	totalCommission, _ := model.GetUserTotalCommission(userId)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"records":          records,
			"total":            total,
			"total_commission": totalCommission,
		},
	})
}

// GetMyWithdrawals — 获取我的提现记录
func GetMyWithdrawals(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	ws, err := model.GetWithdrawalByUser(userId, 50)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": ws})
}

// RequestWithdrawal — 申请提现
func RequestWithdrawal(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	var req struct {
		Amount      int64  `json:"amount"`
		AccountInfo string `json:"account_info"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid request"})
		return
	}
	setting, _ := model.GetCommissionSetting()
	if req.Amount < setting.MinWithdraw {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "below minimum withdrawal amount"})
		return
	}
	// 直接查询用户佣金余额
	user, err := model.GetUserById(userId, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "user not found"})
		return
	}
	if user.CommissionBalance < req.Amount {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "insufficient commission balance"})
		return
	}
	w := &model.WithdrawalRequest{
		UserId:      userId,
		Amount:      req.Amount,
		AccountInfo: req.AccountInfo,
	}
	if err := model.CreateWithdrawal(w); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "withdrawal request submitted"})
}

// AdminGetWithdrawals — 管理员查看所有提现
func AdminGetWithdrawals(c *gin.Context) {
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "0"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	ws, total, err := model.GetAllWithdrawals(status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": ws, "total": total, "page": page})
}

// AdminProcessWithdrawal — 管理员处理提现
func AdminProcessWithdrawal(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid id"})
		return
	}
	var req struct {
		Action string `json:"action"`
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid request"})
		return
	}
	switch req.Action {
	case "approve":
		err = model.ApproveWithdrawal(id, req.Remark)
	case "reject":
		err = model.RejectWithdrawal(id, req.Remark)
	case "complete":
		err = model.CompleteWithdrawal(id, req.Remark)
	default:
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid action"})
		return
	}
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "processed"})
}
