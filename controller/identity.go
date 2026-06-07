package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/i18n"
	"github.com/quantumclaw/quantumclaw/model"
)

type submitIdentityRequest struct {
	IdentityName   string `json:"identity_name" binding:"required"`
	IdentityNumber string `json:"identity_number" binding:"required"`
}

// SubmitIdentityUpload 用户提交实名认证信息
func SubmitIdentityUpload(c *gin.Context) {
	var req submitIdentityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_parameter"),
		})
		return
	}

	userId := c.GetInt("id")
	user, err := model.GetUserById(userId, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "用户不存在"})
		return
	}

	if user.IdentityVerified {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "已通过实名认证，无需重复提交",
		})
		return
	}

	if err := model.DB.Model(user).Updates(map[string]interface{}{
		"identity_name":   req.IdentityName,
		"identity_number": req.IdentityNumber,
	}).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "提交身份信息失败，请稍后重试",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "身份信息已提交，等待管理员审核",
	})
}

// AdminVerifyIdentity 管理员审核通过/拒绝用户实名认证
type verifyIdentityRequest struct {
	UserID int  `json:"user_id" binding:"required"`
	Approve bool `json:"approve"`
}

func AdminVerifyIdentity(c *gin.Context) {
	var req verifyIdentityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_parameter"),
		})
		return
	}

	user, err := model.GetUserById(req.UserID, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "用户不存在"})
		return
	}

	if req.Approve {
		if err := model.DB.Model(user).Update("identity_verified", true).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "审核操作失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "实名认证已通过",
		})
	} else {
		if err := model.DB.Model(user).Updates(map[string]interface{}{
			"identity_name":   "",
			"identity_number": "",
		}).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "驳回操作失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "实名认证已驳回，请通知用户重新提交",
		})
	}
}
