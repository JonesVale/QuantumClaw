package controller

// ============================================================
// user_common.go — 共享辅助函数（被其他 user_*.go 引用）
// 包含：SetupLogin, ValidatePasswordStrength, getAdminOrgFilter
// ============================================================

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/ctxkey"
	"github.com/quantumclaw/quantumclaw/model"
)

// SetupLogin 设置 session & cookies 并返回用户信息
func SetupLogin(user *model.User, c *gin.Context) {
	session := sessions.Default(c)
	session.Set("id", user.Id)
	session.Set("username", user.Username)
	session.Set("role", user.Role)
	session.Set("status", user.Status)
	session.Set("organization_id", user.OrganizationID)
	session.Set("login_time", time.Now().Unix())
	err := session.Save()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "无法保存会话信息，请重试",
			"success": false,
		})
		return
	}
	unreadCount, _ := model.GetUnreadNotificationCount(user.Id)

	c.JSON(http.StatusOK, gin.H{
		"message": "",
		"success": true,
		"data": gin.H{
			"id":               user.Id,
			"username":         user.Username,
			"display_name":     user.DisplayName,
			"email":            user.Email,
			"avatar_url":       user.AvatarURL,
			"role":             user.Role,
			"status":           user.Status,
			"organization_id":  user.OrganizationID,
			"unread_count":     unreadCount,
			"quota_for_new_user": user.Quota > 0,
			"trial_balance":    user.CashBalance,
		},
	})
}

// ValidatePasswordStrength 检查密码强度并返回错误信息（空字符串表示通过）
// 强度要求通过 config.PasswordMinLength / config.PasswordRequireUpper 等配置控制
func ValidatePasswordStrength(password string) string {
	if len(password) < config.PasswordMinLength {
		return fmt.Sprintf("密码长度至少%d位", config.PasswordMinLength)
	}
	if config.PasswordRequireUpper {
		hasUpper := false
		for _, c := range password {
			if c >= 'A' && c <= 'Z' {
				hasUpper = true
				break
			}
		}
		if !hasUpper {
			return "密码需要包含大写字母"
		}
	}
	if config.PasswordRequireNumber {
		hasNumber := false
		for _, c := range password {
			if c >= '0' && c <= '9' {
				hasNumber = true
				break
			}
		}
		if !hasNumber {
			return "密码需要包含数字"
		}
	}
	if config.PasswordRequireSpecial {
		hasSpecial := false
		for _, c := range password {
			if c >= 'a' && c <= 'z' {
				continue
			}
			if c >= 'A' && c <= 'Z' {
				continue
			}
			if c >= '0' && c <= '9' {
				continue
			}
			hasSpecial = true
			break
		}
		if !hasSpecial {
			return "密码需要包含特殊字符"
		}
	}
	return ""
}

// getAdminOrgFilter 获取管理员可见的组织范围
// Root (100) 可见全部；其他 admin (10+) 仅限本组织
func getAdminOrgFilter(c *gin.Context) int {
	role := c.GetInt(ctxkey.Role)
	if role >= model.RoleRootUser {
		return 0 // 不限组织
	}
	return c.GetInt(ctxkey.OrganizationID)
}
