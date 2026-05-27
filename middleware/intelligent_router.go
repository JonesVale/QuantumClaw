package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/service"
)

// IntelligentRouterContext sets up the routing context for the intelligent router.
// This runs before Distribute() and signals that routing should be performance-aware.
func IntelligentRouterContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := service.GetRouterConfig()
		c.Set("intelligent_router_config", cfg)
		c.Set("intelligent_router_enabled", cfg.Enabled)
		c.Next()
	}
}
