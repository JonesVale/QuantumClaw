package controller

// ============================================================
// user_auth.go — 登录 / 注册 / 登出 / 会话管理
// 从原 controller/user.go 拆分，包含认证相关 handler
// ============================================================

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/i18n"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
	"gorm.io/gorm"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(c *gin.Context) {
	if !config.PasswordLoginEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "管理员关闭了密码登录",
			"success": false,
		})
		return
	}
	var loginRequest LoginRequest
	err := json.NewDecoder(c.Request.Body).Decode(&loginRequest)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": i18n.Translate(c, "invalid_parameter"),
			"success": false,
		})
		return
	}
	username := loginRequest.Username
	password := loginRequest.Password
	if username == "" || password == "" {
		c.JSON(http.StatusOK, gin.H{
			"message": i18n.Translate(c, "invalid_parameter"),
			"success": false,
		})
		return
	}
	user := model.User{
		Username: username,
		Password: password,
	}
	err = user.ValidateAndFill()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"success": false,
		})
		return
	}
	SetupLogin(&user, c)
}

func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	err := session.Save()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"success": false,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "",
		"success": true,
	})
}

func Register(c *gin.Context) {
	ctx := c.Request.Context()
	if !config.RegisterEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "管理员关闭了新用户注册",
			"success": false,
		})
		return
	}
	if !config.PasswordRegisterEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "管理员关闭了通过密码进行注册，请使用第三方账户验证的形式进行注册",
			"success": false,
		})
		return
	}
	var user model.User
	err := json.NewDecoder(c.Request.Body).Decode(&user)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_parameter"),
		})
		return
	}
	if err := common.Validate.Struct(&user); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_input"),
		})
		return
	}
	if msg := ValidatePasswordStrength(user.Password); msg != "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": msg,
		})
		return
	}
	if config.EmailVerificationEnabled {
		if user.Email == "" || user.VerificationCode == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "管理员开启了邮箱验证，请输入邮箱地址和验证码",
			})
			return
		}
		if !common.VerifyCodeWithKey(user.Email, user.VerificationCode, common.EmailVerificationPurpose) {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "验证码错误或已过期",
			})
			return
		}
	}
	identifier := user.Username
	if identifier == "" {
		identifier = user.Email
	}
	if identifier == "" {
		identifier = user.Phone
	}
	if identifier == "" {
		identifier = user.QQ
	}
	if identifier == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_parameter"),
		})
		return
	}

	isEmail := strings.Contains(identifier, "@")
	isPhone := false
	isQQ := false
	if !isEmail {
		digitsOnly := true
		for _, ch := range identifier {
			if ch < '0' || ch > '9' {
				digitsOnly = false
				break
			}
		}
		if digitsOnly {
			if len(identifier) == 11 && identifier[0] == '1' {
				isPhone = true
			} else if len(identifier) >= 5 && len(identifier) <= 12 {
				isQQ = true
			}
		}
	}

	var existingUser model.User
	dupField := ""
	if model.DB.Where("username = ?", identifier).First(&existingUser).Error == nil {
		dupField = "username"
	} else if isEmail && model.DB.Where("email = ?", identifier).First(&existingUser).Error == nil {
		dupField = "email"
	} else if isPhone && model.DB.Where("phone = ?", identifier).First(&existingUser).Error == nil {
		dupField = "phone"
	} else if isQQ && model.DB.Where("qq = ?", identifier).First(&existingUser).Error == nil {
		dupField = "qq"
	}
	if dupField != "" {
		var msg string
		switch dupField {
		case "email":
			msg = i18n.Translate(c, "email_exists")
		case "phone":
			msg = i18n.Translate(c, "phone_exists")
		case "qq":
			msg = i18n.Translate(c, "qq_exists")
		default:
			msg = i18n.Translate(c, "username_exists")
		}
		c.JSON(http.StatusOK, gin.H{"success": false, "message": msg})
		return
	}

	affCode := user.AffCode
	inviterId, _ := model.GetUserIdByAffCode(affCode)
	cleanUser := model.User{
		Username:    identifier,
		Password:    user.Password,
		DisplayName: user.Username,
		InviterId:   inviterId,
		Role:        user.Role,
	}
	if isEmail {
		cleanUser.Email = identifier
	}
	if isPhone {
		cleanUser.Phone = identifier
	}
	if isQQ {
		cleanUser.QQ = identifier
	}
	if config.EmailVerificationEnabled && user.Email != "" {
		cleanUser.Email = user.Email
	}
	if user.Phone != "" && user.Phone != identifier {
		cleanUser.Phone = user.Phone
	}
	if user.QQ != "" && user.QQ != identifier {
		cleanUser.QQ = user.QQ
	}

	if cleanUser.Role != model.RoleCommonUser && cleanUser.Role != model.RoleSupplier {
		cleanUser.Role = model.RoleCommonUser
	}

	if err := cleanUser.Insert(ctx, inviterId); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "registration_failed"),
		})
		return
	}

	if inviterId > 0 {
		setting, _ := model.GetCommissionSetting()
		if setting.Enabled && setting.RegisterReward > 0 {
			model.CreateCommissionRecord(&model.CommissionRecord{
				UserId:      inviterId,
				FromUserId:  cleanUser.Id,
				Type:        "register",
				Amount:      setting.RegisterReward,
				Status:      "settled",
				Description: "好友注册奖励",
			})
			model.DB.Model(&model.User{}).Where("id = ?", inviterId).
				UpdateColumn("quota", gorm.Expr("quota + ?", setting.RegisterReward))
		}

		relation := model.AffiliateRelation{
			PromoterId:  inviterId,
			ConsumerId: cleanUser.Id,
			CreatedTime: time.Now().Unix(),
		}
		var existing int64
		model.DB.Model(&model.AffiliateRelation{}).Where("consumer_id = ?", cleanUser.Id).Count(&existing)
		if existing == 0 {
			if err := model.DB.Create(&relation).Error; err != nil {
				logger.SysError(fmt.Sprintf("failed to create affiliate_relation: %v", err))
			}
		}
	}

	SetupLogin(&cleanUser, c)
}
