package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/model"
)

// GetSecurityActivity 获取用户的安全活动记录
// 返回最近的登录记录、密码修改、支付操作等
func GetSecurityActivity(c *gin.Context) {
	userId := c.GetInt("id")

	// 获取最近的交易记录
	transactionLogs, err := model.GetRecentTransactionLogs(userId, 20)
	if err != nil {
		transactionLogs = nil // 不影响返回其他数据
	}

	// 获取用户日志中的安全相关记录（登录、管理操作等）
	// 类型：LogTypeManage (3) = 管理操作
	userLogs, err := model.GetUserLogs(userId, 0, 0, 0, "", "", 0, 20)
	if err != nil {
		userLogs = nil
	}

	// 获取 WebAuthn 凭证数量
	credentials, _ := model.GetUserCredentials(userId)
	passkeyCount := len(credentials)

	// 获取 2FA 状态
	twoFA, _ := model.GetTwoFAByUserId(userId)
	twoFAEnabled := twoFA != nil

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"transaction_logs": transactionLogs,
			"activity_logs":    userLogs,
			"security_status": gin.H{
				"twofa_enabled":    twoFAEnabled,
				"passkey_count":    passkeyCount,
			},
		},
	})
}
