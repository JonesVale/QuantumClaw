package middleware

import (
	"net/http"
	"strings"

	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/logger"

	"github.com/gin-gonic/gin"
)

// WebhookIPWhitelist checks if the request comes from an allowed IP
func WebhookIPWhitelist() gin.HandlerFunc {
	whitelist := strings.Split(config.PaymentWebhookIPWhitelist, ",")
	if len(whitelist) == 0 || (len(whitelist) == 1 && whitelist[0] == "") {
		return func(c *gin.Context) { c.Next() } // No whitelist configured = skip
	}
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		for _, ip := range whitelist {
			if strings.TrimSpace(ip) == clientIP {
				c.Next()
				return
			}
		}
		logger.Warn(c.Request.Context(), "webhook rejected by IP whitelist: "+clientIP)
		c.AbortWithStatus(http.StatusForbidden)
	}
}
