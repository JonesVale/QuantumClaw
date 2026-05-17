package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/model"
)

// RequirePaymentAuth enforces extra authentication for payment operations.
// If the admin has WebAuthn/2FA configured, the current session must have
// been verified (TOTP or WebAuthn) before proceeding.
func RequirePaymentAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.GetInt("id")

		// Check if admin user has WebAuthn credentials configured
		hasPasskey, err := model.HasWebAuthnCredential(userId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "验证 2FA 状态失败",
			})
			c.Abort()
			return
		}

		// Check if admin user has TOTP 2FA enabled
		_, twofaErr := model.GetTwoFAByUserId(userId)
		has2FA := hasPasskey || twofaErr == nil

		if has2FA {
			session := sessions.Default(c)
			webauthnVerified := session.Get("webauthn_verified")
			twofaVerified := session.Get("twofa_verified")

			verified := false
			if webauthnVerified != nil {
				if v, ok := webauthnVerified.(bool); ok && v {
					verified = true
				}
			}
			if !verified && twofaVerified != nil {
				if v, ok := twofaVerified.(bool); ok && v {
					verified = true
				}
			}

			if !verified {
				c.JSON(http.StatusForbidden, gin.H{
					"success": false,
					"message": "需要二次验证才能执行支付操作，请先完成 WebAuthn 或 TOTP 验证",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
