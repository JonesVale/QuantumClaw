package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func Cache() func(c *gin.Context) {
	return func(c *gin.Context) {
		uri := c.Request.RequestURI
		// Hashed static assets (JS/CSS/images): cache long-term
		// SPA routes (no file extension): always fresh
		if strings.Contains(uri, "/static/") {
			c.Header("Cache-Control", "max-age=604800, immutable")
		} else {
			c.Header("Cache-Control", "no-cache")
		}
		c.Next()
	}
}
