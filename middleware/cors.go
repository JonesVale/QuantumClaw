package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/config"
)

func CORS() gin.HandlerFunc {
	cfg := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-Id"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: config.CORSAllowCredentials,
		MaxAge:           time.Duration(config.CORSMaxAge) * time.Second,
	}

	origins := config.GetAllowedOrigins()
	if len(origins) > 0 && origins[0] != "*" {
		cfg.AllowOrigins = origins
		cfg.AllowAllOrigins = false
	} else if len(origins) > 0 && origins[0] == "*" {
		// Explicit wildcard
		cfg.AllowAllOrigins = true
		if cfg.AllowCredentials {
			// Cannot use AllowAllOrigins with credentials; use explicit wildcard origin instead
			cfg.AllowAllOrigins = false
			cfg.AllowOrigins = []string{"*"}
		}
	} else {
		// No origins configured - default to same-origin only
		cfg.AllowAllOrigins = false
		cfg.AllowOrigins = nil // browsers enforce same-origin
	}

	return cors.New(cfg)
}
