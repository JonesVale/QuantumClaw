package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/model"
)

// SubmitResellerWithdrawal 渠道商提交提现申请
// POST /api/reseller/withdraw
func SubmitResellerWithdrawal(c *gin.Context) {
	userId := c.GetInt("id")

	// 查找 reseller
	var reseller model.Reseller
	if err := model.DB.Where("user_id = ? AND status = 1", userId).First(&reseller).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "仅渠道商可提现"})
		return
	}

	if reseller.Balance <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "余额不足"})
		return
	}

	// 获取最低提现金额
	var minAmount float64 = 10.0
	var cfg model.PlatformConfig
	if err := model.DB.Where("`key` = ?", "min_withdraw_amount").First(&cfg).Error; err == nil {
		if v, err := strToFloat(cfg.Value); err == nil && v > 0 {
			minAmount = v
		}
	}

	if reseller.Balance < minAmount {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "最低提现金额为 $" + fmt.Sprintf("%.2f", minAmount) + "，当前余额 $" + fmt.Sprintf("%.2f", reseller.Balance),
		})
		return
	}

	// 创建提现记录（复用现有 WithdrawalRequest 表）
	withdraw := &model.WithdrawalRequest{
		UserId: userId,
		Amount: int64(reseller.Balance * 100), // 转为分
		Status: "pending",
	}
	if err := model.CreateWithdrawal(withdraw); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "创建提现记录失败"})
		return
	}

	// 清零余额（已申请提现）
	model.DB.Model(&model.Reseller{}).Where("id = ?", reseller.Id).Update("balance", 0)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "提现申请已提交，等待审核"})
}

// GetResellerBalance 获取渠道商余额和提现记录
// GET /api/reseller/balance
func GetResellerBalance(c *gin.Context) {
	userId := c.GetInt("id")

	var reseller model.Reseller
	if err := model.DB.Where("user_id = ?", userId).First(&reseller).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"balance":      0,
				"total_earned": 0,
				"withdrawals":  []model.WithdrawalRequest{},
			},
		})
		return
	}

	var withdrawals []model.WithdrawalRequest
	model.DB.Where("user_id = ?", userId).Order("created_at DESC").Limit(5).Find(&withdrawals)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"balance":      reseller.Balance,
			"total_earned": reseller.TotalEarned,
			"withdrawals":  withdrawals,
		},
	})
}

func strToFloat(s string) (float64, error) {
	var v float64
	_, err := fmt.Sscanf(s, "%f", &v)
	return v, err
}

// ListResellers 管理员查看所有渠道商
// GET /api/admin/resellers
func ListResellers(c *gin.Context) {
	var resellers []model.Reseller
	model.DB.Order("created_time DESC").Find(&resellers)

	// 获取每个渠道商的用户名
	type ResellerWithUser struct {
		model.Reseller
		Username string `json:"username"`
	}
	result := make([]ResellerWithUser, 0, len(resellers))
	for _, r := range resellers {
		u, err := model.GetUserById(r.UserId, false)
		username := ""
		if err == nil && u != nil {
			username = u.Username
		}
		result = append(result, ResellerWithUser{Reseller: r, Username: username})
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// ListWithdrawals 管理员查看所有提现记录
// GET /api/admin/withdrawals?status=pending
func ListWithdrawals(c *gin.Context) {
	status := c.Query("status")
	query := model.DB.Model(&model.WithdrawalRequest{}).Order("created_at DESC")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	var withdrawals []model.WithdrawalRequest
	query.Find(&withdrawals)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": withdrawals})
}

// ApproveWithdrawal 管理员审核提现
// POST /api/admin/withdrawals/:id/approve
func ApproveWithdrawal(c *gin.Context) {
	id := c.Param("id")
	var w model.WithdrawalRequest
	if err := model.DB.First(&w, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "记录不存在"})
		return
	}
	model.DB.Model(&w).Updates(map[string]interface{}{"status": "approved"})
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "已审核通过"})
}
