package controller

import (
	"context"
	"encoding/base64"
	"fmt"
	"image/png"
	"bytes"
	"net/http"
	"strings"
	"time"

	"github.com/quantumclaw/quantumclaw/model"
	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// ==================== 2FA/TOTP 类型 ====================

type TwoFASetupResponse struct {
	Secret   string `json:"secret"`
	QRCodeURL string `json:"qr_code_url"` // data URI
}

// ==================== 初始化 2FA（生成密钥）====================

// InitTwoFA 初始化两步验证
func InitTwoFA(c *gin.Context) {
	userId := c.GetInt("id")

	// 生成新密钥
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "QuantumClaw",
		AccountName: fmt.Sprintf("user_%d", userId),
		Period:     30,
		SecretSize: 20,
		Algorithm:  otp.AlgorithmSHA1,
		Digits:     otp.DigitsSix,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "生成密钥失败: " + err.Error()})
		return
	}

	// 生成二维码图片 data URL
	img, err := key.Image(200, 200)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "生成二维码失败: " + err.Error()})
		return
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "编码二维码失败: " + err.Error()})
		return
	}
	qrDataURL := "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())

	// 临时存储密钥（未验证，不生效）
	tempSecret := key.Secret()
	model.SetTwoFATempSecret(userId, tempSecret)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"secret":      tempSecret,
			"qr_code_url": qrDataURL,
		},
	})
}

// ==================== 启用 2FA（验证 TOTP 码后生效）====================

type VerifyTwoFARequest struct {
	Code string `json:"code" binding:"required"`
}

func VerifyAndEnableTwoFA(c *gin.Context) {
	userId := c.GetInt("id")

	var req VerifyTwoFARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	code := strings.TrimSpace(req.Code)
	if len(code) != 6 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "验证码必须是6位数字"})
		return
	}

	// 获取临时密钥
	tempSecret := model.GetTwoFATempSecret(userId)
	if tempSecret == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请先初始化2FA"})
		return
	}

	// 验证 TOTP
	if !totp.Validate(code, tempSecret) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "验证码错误"})
		return
	}

	// 生成备用码
	backupCodes := generateBackupCodes(10)

	// 保存 2FA 密钥和备用码
	if err := model.EnableTwoFA(userId, tempSecret, backupCodes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "保存失败: " + err.Error()})
		return
	}

	// 清除临时密钥
	model.ClearTwoFATempSecret(userId)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "2FA 已启用，请妥善保管以下备用码",
		"data": gin.H{
			"backup_codes": backupCodes,
		},
	})
}

// ==================== 验证 2FA（登录时使用）====================

type LoginTwoFARequest struct {
	UserId int    `json:"user_id"`
	Code   string `json:"code" binding:"required"`
}

func VerifyLoginTwoFA(c *gin.Context) {
	var req LoginTwoFARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	code := strings.TrimSpace(req.Code)
	if len(code) != 6 && len(code) != 10 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "验证码格式错误"})
		return
	}

	// 获取用户的 2FA 密钥
	twoFA, err := model.GetTwoFAByUserId(req.UserId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "该用户未启用2FA"})
		return
	}

	if len(code) == 10 {
		// 备用码验证
		if !model.ValidateTwoFABackupCode(req.UserId, code) {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "备用码无效或已使用"})
			return
		}
		// 备用码一次性使用
		model.ConsumeTwoFABackupCode(req.UserId, code)
		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"verified": true,
			"method":   "backup_code",
		})
		return
	}

	// TOTP 验证（允许 ±1 步长容差）
	valid := totp.Validate(code, twoFA.Secret)
	if !valid {
		opts := totp.ValidateOpts{
			Period:    30,
			Skew:      1,
			Digits:    otp.DigitsSix,
			Algorithm: otp.AlgorithmSHA1,
		}
		var valErr error
		valid, valErr = totp.ValidateCustom(code, twoFA.Secret, time.Now(), opts)
		if valErr != nil || !valid {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "验证码错误"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"verified": true,
		"method":   "totp",
	})
}

// ==================== 禁用 2FA ====================

type DisableTwoFARequest struct {
	Code string `json:"code" binding:"required"`
}

func DisableTwoFA(c *gin.Context) {
	userId := c.GetInt("id")

	var req DisableTwoFARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	twoFA, err := model.GetTwoFAByUserId(userId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "未启用2FA"})
		return
	}

	// 验证当前 2FA 代码
	if !totp.Validate(req.Code, twoFA.Secret) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "验证码错误，无法禁用2FA"})
		return
	}

	if err := model.DisableTwoFA(userId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "禁用失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "2FA 已禁用",
	})
}

// ==================== 获取用户 2FA 状态 ====================

func GetTwoFAStatus(c *gin.Context) {
	userId := c.GetInt("id")
	twoFA, err := model.GetTwoFAByUserId(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success":       true,
			"twofa_enabled": false,
		})
		return
	}
	backupCodesRemaining := model.CountRemainingBackupCodes(userId)
	c.JSON(http.StatusOK, gin.H{
		"success":                true,
		"twofa_enabled":          true,
		"backup_codes_remaining": backupCodesRemaining,
	})
	_ = twoFA
}

// ==================== 工具函数 ====================

// generateBackupCodes 生成 N 个备用码（每个 10 位）
func generateBackupCodes(n int) []string {
	codes := make([]string, n)
	for i := 0; i < n; i++ {
		codes[i] = generateBackupCode()
	}
	return codes
}

func generateBackupCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 10)
	r := time.Now().UnixNano()
	for i := range b {
		b[i] = charset[int(r/int64(i+1))%len(charset)]
	}
	return string(b)
}

// avoid unused import
var _ = context.Background
