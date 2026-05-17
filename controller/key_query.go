package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/model"
)

// QueryTokenByKey allows a key holder to query their own quota/usage info
// by providing their API key as a query parameter.
// This is a public endpoint (no auth middleware) intended for key-level self-service.
func QueryTokenByKey(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "missing key parameter",
		})
		return
	}

	token, err := model.GetTokenByKey(key, true)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "invalid or unknown key",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"name":            token.Name,
			"used_quota":      token.UsedQuota,
			"remain_quota":    token.RemainQuota,
			"unlimited_quota": token.UnlimitedQuota,
			"status":          token.Status,
		},
	})
}
