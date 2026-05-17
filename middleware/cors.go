package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/config"
)

func CORS() gin.HandlerFunc {
	cfg := cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-Id"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: config.CORSAllowCredentials,
		MaxAge:           time.Duration(config.CORSMaxAge) * time.Second,
	}

	origins := config.GetAllowedOrigins()
	if len(origins) > 0 && origins[0] != "*" {
		cfg.AllowAllOrigins = false
		cfg.AllowOrigins = origins
	}

	return cors.New(cfg)
}
