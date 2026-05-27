package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
)

// CascadeAuth validates X-Cascade-Key header for slave node API calls.
// Only available on the master node.
func CascadeAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-Cascade-Key")
		if key == "" {
			abortWithMessage(c, http.StatusUnauthorized, "missing X-Cascade-Key header")
			return
		}

		node, err := model.CascadeGetNodeByAPIKey(key)
		if err != nil {
			logger.Error(c.Request.Context(), "CascadeAuth lookup error: "+err.Error())
			abortWithMessage(c, http.StatusInternalServerError, "auth internal error")
			return
		}
		if node == nil {
			abortWithMessage(c, http.StatusForbidden, "invalid cascade API key")
			return
		}
		if node.Status != 1 {
			abortWithMessage(c, http.StatusForbidden, "node is disabled")
			return
		}

		c.Set("cascade_node_id", node.Id)
		c.Set("cascade_node_name", node.Name)
		c.Next()
	}
}
