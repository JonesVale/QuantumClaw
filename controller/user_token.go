package controller

// ============================================================
// user_token.go — Access Token / 推广码管理
// 从原 controller/user.go 拆分
// ============================================================

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/ctxkey"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/common/random"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/service"
)

func GenerateAccessToken(c *gin.Context) {
	id := c.GetInt(ctxkey.Id)
	user, err := model.GetUserById(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	rawToken := random.GetUUID()
	user.AccessToken = common.SHA256Hash(rawToken)

	if model.DB.Where("access_token = ?", user.AccessToken).First(user).RowsAffected != 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "请重试，系统生成的 UUID 竟然重复了！",
		})
		return
	}

	if err := user.Update(false); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    rawToken,
	})
	return
}

// GenerateJWT issues a stateless JWT token for the authenticated user.
// Useful for distributed / multi-instance deployments where session state
// should not be pinned to a single server.
func GenerateJWT(c *gin.Context) {
	id := c.GetInt(ctxkey.Id)
	username := c.GetString(ctxkey.Username)
	role := c.GetInt(ctxkey.Role)

	token, err := service.IssueJWT(id, username, role)
	if err != nil {
		logger.SysErrorf("failed to issue JWT for user %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "无法生成 JWT 令牌，请稍后重试",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    token,
	})
}

func GetAffCode(c *gin.Context) {
	id := c.GetInt(ctxkey.Id)
	user, err := model.GetUserById(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if user.AffCode == "" {
		user.AffCode = random.GetRandomString(8)
		if err := user.Update(false); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    user.AffCode,
	})
	return
}
