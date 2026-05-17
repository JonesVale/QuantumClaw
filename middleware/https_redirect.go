package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/config"
)

// HTTPSRedirect redirects HTTP to HTTPS when configured via FORCE_HTTPS env var
func HTTPSRedirect() gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.ForceHTTPS && c.Request.TLS == nil {
			host := c.Request.Host
			if strings.HasPrefix(host, ":") {
				host = "localhost" + host
			}
			url := "https://" + host + c.Request.RequestURI
			c.Redirect(http.StatusMovedPermanently, url)
			c.Abort()
			return
		}
		c.Next()
	}
}
