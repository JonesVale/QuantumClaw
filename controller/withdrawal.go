package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
)

// SubmitWithdrawalRequest 供应商提交提现申请
type SubmitWithdrawalRequest struct {
	Amount   int64  `json:"amount" binding:"required,min=100"` // 提现金额（分）
	BankInfo string `json:"bank_info" binding:"required"`      // 收款信息 JSON
}

// SubmitWithdrawal 提交提现申请
func SubmitWithdrawal(c *gin.Context) {
	var req SubmitWithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误：" + err.Error()})
		return
	}

	userId := c.GetInt("id")

	// 0. 身份信息审核 — 提现前必须完成实名认证
	user, err := model.GetUserById(userId, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "用户不存在"})
		return
	}
	if !user.IdentityVerified {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "请先完成实名认证后再申请提现。请在「个人资料」中提交身份信息",
		})
		return
	}

	// 1. 检查可提现金额
	available, err := model.GetUserWithdrawableBalance(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "查询可提现金额失败"})
		return
	}
	if req.Amount > available {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("可提现金额不足，当前可提 %d 分 (¥%.2f)，申请 %d 分",
				available, float64(available)/100, req.Amount),
		})
		return
	}
	if req.Amount < model.WithdrawMinAmount {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("最低提现金额为 %d 分 (¥%.2f)", model.WithdrawMinAmount, float64(model.WithdrawMinAmount)/100),
		})
		return
	}

	// 2. 计算待扣入驻费
	var totalPendingFee int64
	pendingFees, _ := model.GetPendingPlatformFees(userId)
	for _, fee := range pendingFees {
		totalPendingFee += fee.FeeAmount
	}

	// 3. 如果可提现包含入驻费，优先扣入驻费
	platformFeeDeduct := totalPendingFee
	if platformFeeDeduct > req.Amount {
		platformFeeDeduct = req.Amount // 入驻费不能超过提现金额
	}
	netAmount := req.Amount - platformFeeDeduct
	if netAmount < 0 {
		netAmount = 0
	}

	// 4+5. 创建提现记录 + 入驻费扣除（同一事务）
	w := &model.WithdrawalRequest{
		UserId:            userId,
		Amount:            req.Amount,
		PlatformFeeAmount: platformFeeDeduct,
		NetAmount:         netAmount,
		Status:            model.WithdrawStatusPending,
		AccountInfo:       req.BankInfo,
	}
	if err := model.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(w).Error; err != nil {
			return fmt.Errorf("创建提现记录: %w", err)
		}
		// 标记入驻费为已扣除
		remaining := platformFeeDeduct
		for _, fee := range pendingFees {
			if remaining <= 0 {
				break
			}
			if fee.FeeAmount <= remaining {
				if err := tx.Model(&model.PlatformFeeRecord{}).Where("id = ?", fee.Id).
					Update("status", model.PlatformFeeStatusDeducted).Error; err != nil {
					return fmt.Errorf("扣除入驻费 %d: %w", fee.Id, err)
				}
				remaining -= fee.FeeAmount
			}
		}
		return nil
	}); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	logger.Info(c.Request.Context(), fmt.Sprintf("供应商 %d 提交提现申请: amount=%d, net=%d, fee=%d, id=%d",
		userId, req.Amount, netAmount, platformFeeDeduct, w.Id))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"id":                    w.Id,
			"amount":                w.Amount,
			"platform_fee_deducted": w.PlatformFeeAmount,
			"net_amount":            w.NetAmount,
			"status":                w.Status,
		},
	})
}

// GetMyEarningsByChannel 获取各渠道收益汇总
func GetMyEarningsByChannel(c *gin.Context) {
	userId := c.GetInt("id")
	summaries, err := model.GetUserEarningsByChannel(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": summaries})
}

// GetMyWithdrawable 获取我可提现金额
func GetMyWithdrawable(c *gin.Context) {
	userId := c.GetInt("id")

	available, err := model.GetUserWithdrawableBalance(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "查询失败"})
		return
	}

	totalEarned, _ := model.GetUserEarningsSum(userId, model.EarningStatusSettled)
	pendingFees, _ := model.GetPendingPlatformFees(userId)
	var totalPendingFee int64
	for _, fee := range pendingFees {
		totalPendingFee += fee.FeeAmount
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"available":         available,
			"available_yuan":    float64(available) / 100,
			"total_earned":      totalEarned,
			"total_earned_yuan": float64(totalEarned) / 100,
			"pending_fee":       totalPendingFee,
			"pending_fee_yuan":  float64(totalPendingFee) / 100,
		},
	})
}

// ==================== 管理员 ====================

// AdminApproveWithdrawal 管理员审批通过提现
func AdminApproveWithdrawal(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的ID"})
		return
	}

	remark := c.DefaultQuery("remark", "管理员审批通过")
	if err := model.ApproveWithdrawal(id, remark); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "已通过"})
}

// AdminRejectWithdrawal 管理员拒绝提现
func AdminRejectWithdrawal(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的ID"})
		return
	}

	remark := c.DefaultQuery("remark", "管理员拒绝")
	if err := model.RejectWithdrawal(id, remark); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "已拒绝"})
}

// AdminCompleteWithdrawal 管理员标记提现已完成（已打款）
func AdminCompleteWithdrawal(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的ID"})
		return
	}

	remark := c.DefaultQuery("remark", "已打款")
	if err := model.CompleteWithdrawal(id, remark); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "已完成"})
}
