package controller

// ============================================================
// user_profile.go — 邮箱绑定 / 升级渠道商 / 团队管理
// 从原 controller/user.go 拆分
// ============================================================

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/model"
)

func EmailBind(c *gin.Context) {
	email := c.Query("email")
	code := c.Query("code")
	if !common.VerifyCodeWithKey(email, code, common.EmailVerificationPurpose) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "验证码错误或已过期",
		})
		return
	}
	id := c.GetInt("id")
	user := model.User{
		Id: id,
	}
	err := user.FillUserById()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	user.Email = email
	err = user.Update(false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if user.Role == model.RoleRootUser {
		config.RootUserEmail = email
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

// UpgradeToProvider 升级为渠道商
// 用户配置好 API 渠道后即可直接升级，无需管理员审批
func UpgradeToProvider(c *gin.Context) {
	userId := c.GetInt("id")

	user, err := model.GetUserById(userId, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "用户不存在"})
		return
	}
	if user.UserType == model.UserTypeProvider {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "已是渠道商，无需重复升级"})
		return
	}

	if err := model.DB.Model(&model.User{}).Where("id = ?", userId).
		Updates(map[string]interface{}{
			"user_type": model.UserTypeProvider,
			"role":      model.RoleSupplier,
		}).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "升级失败"})
		return
	}

	model.RecordLog(c.Request.Context(), userId, model.LogTypeSystem, "用户升级为渠道商")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "升级成功，您现在可以管理 API 渠道了",
	})
}

// GetMyTeam 获取我的团队成员
func GetMyTeam(c *gin.Context) {
	userId := c.GetInt("id")

	members, err := model.GetUsersByInviter(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "获取团队成员失败"})
		return
	}

	type TeamMember struct {
		Id            int    `json:"id"`
		Username     string `json:"username"`
		DisplayName  string `json:"display_name"`
		Role         int    `json:"role"`
		Status       int    `json:"status"`
		Quota        int64  `json:"quota"`
		UsedQuota    int64  `json:"used_quota"`
		RequestCount int    `json:"request_count"`
	}

	var result []TeamMember
	for _, m := range members {
		result = append(result, TeamMember{
			Id:            m.Id,
			Username:     m.Username,
			DisplayName:  m.DisplayName,
			Role:         m.Role,
			Status:       m.Status,
			Quota:        m.Quota,
			UsedQuota:    m.UsedQuota,
			RequestCount: m.RequestCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    result,
	})
}
