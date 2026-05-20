package controller

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/ctxkey"
	"github.com/quantumclaw/quantumclaw/model"
)

// AdminResetUserPassword — 管理员重置用户密码（不需要旧密码）
func AdminResetUserPassword(c *gin.Context) {
	var req struct {
		UserId      int    `json:"user_id"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid request"})
		return
	}
	if req.NewPassword == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "password cannot be empty"})
		return
	}
	user, err := model.GetUserById(req.UserId, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "user not found"})
		return
	}
	hashedPassword, err := common.Password2Hash(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "failed to hash password"})
		return
	}
	err = model.DB.Model(&model.User{}).Where("id = ?", user.Id).Update("password", hashedPassword).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "password reset successfully"})
}

// ChangePassword — 当前登录用户修改自己的密码（需要旧密码）
func ChangePassword(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid request"})
		return
	}
	if req.OldPassword == "" || req.NewPassword == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "old and new password required"})
		return
	}
	var user model.User
	err := model.DB.Where("id = ?", userId).First(&user).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "user not found"})
		return
	}
	if !common.ValidatePasswordAndHash(req.OldPassword, user.Password) {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "current password is incorrect"})
		return
	}
	hashedPassword, err := common.Password2Hash(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "failed to hash password"})
		return
	}
	err = model.DB.Model(&model.User{}).Where("id = ?", userId).Update("password", hashedPassword).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	// Invalidate current session so user re-logs in
	session := sessions.Default(c)
	session.Clear()
	session.Save()

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "password changed, please login again"})
}

// EmergencyPasswordReset — 紧急重置管理员密码（需要 EMERGENCY_RESET_TOKEN 环境变量）
// 不需要登录，无需旧密码。如果管理员被锁在登录界面外，用此接口直接重置
func EmergencyPasswordReset(c *gin.Context) {
	var req struct {
		ResetToken  string `json:"reset_token"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid request"})
		return
	}

	envToken := os.Getenv("EMERGENCY_RESET_TOKEN")
	if envToken == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "emergency reset not configured, set EMERGENCY_RESET_TOKEN env var"})
		return
	}
	if req.ResetToken != envToken {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid reset token"})
		return
	}
	if len(req.NewPassword) < 6 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "password too short (min 6 chars)"})
		return
	}

	// 找最高权限的管理员
	var adminUser model.User
	if err := model.DB.Where("role >= ?", model.RoleAdminUser).Order("role desc, id asc").First(&adminUser).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "no admin user found"})
		return
	}

	hashed, err := common.Password2Hash(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "failed to hash password"})
		return
	}
	model.DB.Model(&adminUser).Update("password", hashed)

	// 清除所有登录锁定状态
	middleware.ClearAllLoginLocks()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "password reset for admin: " + adminUser.Username,
		"data": gin.H{
			"username": adminUser.Username,
			"hint":     "login immediately with the new password",
		},
	})
}

func GetUserInfo(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	user, err := model.GetUserById(userId, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":           user.Id,
			"username":     user.Username,
			"display_name": user.DisplayName,
			"role":         user.Role,
			"status":       user.Status,
			"email":        user.Email,
			"quota":        user.Quota,
			"used_quota":   user.UsedQuota,
			"aff_code":     user.AffCode,
		},
	})
}
