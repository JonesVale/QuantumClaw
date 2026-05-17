package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/config"
)

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// CSP - inline styles/scripts needed for web UI
		csp := "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; font-src 'self' data:; connect-src 'self' https:; frame-src 'none'; object-src 'none'; base-uri 'self'"
		if config.CSPReportOnly {
			c.Header("Content-Security-Policy-Report-Only", csp)
		} else {
			c.Header("Content-Security-Policy", csp)
		}

		// HSTS (HTTPS enforcement)
		hsts := fmt.Sprintf("max-age=%d", config.HSTSMaxAge)
		if config.HSTSIncludeSubDomains {
			hsts += "; includeSubDomains"
		}
		if config.HSTSPreload {
			hsts += "; preload"
		}
		c.Header("Strict-Transport-Security", hsts)

		// Standard security headers
		c.Header("X-Frame-Options", config.XFrameOptions)
		c.Header("X-Content-Type-Options", "nosniff")
		if config.XSSProtectionEnabled {
			c.Header("X-XSS-Protection", "1; mode=block")
		}
		c.Header("Referrer-Policy", config.ReferrerPolicy)
		c.Header("Permissions-Policy", config.PermissionsPolicy)

		c.Next()
	}
}
