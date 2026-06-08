package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/blacklist"
	"github.com/quantumclaw/quantumclaw/common/ctxkey"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/common/network"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/service"
)

// sessionMaxAge is the maximum session lifetime in seconds (24 hours)
const sessionMaxAge = 86400

func authHelper(c *gin.Context, minRole int) {
	session := sessions.Default(c)
	username := session.Get("username")
	role := session.Get("role")
	id := session.Get("id")
	status := session.Get("status")
	orgId := session.Get("organization_id")

	// Check session age if present
	if loginTimeRaw := session.Get("login_time"); loginTimeRaw != nil {
		if loginTime, ok := loginTimeRaw.(int64); ok {
			if time.Now().Unix()-loginTime > sessionMaxAge {
				session.Clear()
				_ = session.Save()
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"message": "会话已过期，请重新登录",
				})
				c.Abort()
				return
			}
		}
	}

	if username == nil {
		// 1) Try JWT token (for distributed / stateless auth)
		jwtHeader := c.Request.Header.Get("Authorization")
		jwtToken := strings.TrimPrefix(jwtHeader, "Bearer ")
		if jwtToken != "" && jwtToken != jwtHeader {
			jwtUser, jwtErr := service.VerifyJWT(jwtToken)
			if jwtErr == nil && jwtUser != nil {
				// Token is valid
				username = jwtUser.Username
				role = jwtUser.Role
				id = jwtUser.Id
				status = jwtUser.Status
				orgId = jwtUser.OrganizationID
			}
		}

		// 2) Fall back to access token
		if username == nil {
			accessToken := c.Request.Header.Get("Authorization")
			if accessToken == "" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"message": "无权进行此操作，未登录且未提供 access token",
				})
				c.Abort()
				return
			}
			user := model.ValidateAccessToken(accessToken)
			if user != nil && user.Username != "" {
				// Token is valid
				username = user.Username
				role = user.Role
				id = user.Id
				status = user.Status
				orgId = user.OrganizationID
			} else {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": "无权进行此操作，access token 无效",
				})
				c.Abort()
				return
			}
		}
	}

	// 安全性检查：status/role/id 必须在有值后才能断言
	if status == nil || role == nil || id == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "会话数据不完整，请重新登录",
		})
		session.Clear()
		_ = session.Save()
		c.Abort()
		return
	}

	statusInt, statusOk := status.(int)
	roleInt, roleOk := role.(int)
	idInt, idOk := id.(int)
	if !statusOk || !roleOk || !idOk {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "会话数据类型异常，请重新登录",
		})
		session.Clear()
		_ = session.Save()
		c.Abort()
		return
	}

	// 黑名单直接拦截（管理员手动封禁）
	// 注意：UserStatusDisabled 不再自动阻断登录和浏览
	// 禁用状态的用户依然可以登录查看、充值、联系客服
	// API 层面的权限由配额/余额逻辑控制
	if blacklist.IsUserBanned(idInt) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户已被封禁",
		})
		session := sessions.Default(c)
		session.Clear()
		_ = session.Save()
		c.Abort()
		return
	}
	// 绂佺敤鐢ㄦ埛闃绘鎵€鏈塏PI璇锋眰
	if statusInt == model.UserStatusDisabled {
		logger.SysWarnf("disabled user %d blocked on route: %s", idInt, c.Request.URL.Path)
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "璐﹀彿宸插仠鐢紝璇疯仈绯荤鐞嗗憳",
		})
		c.Abort()
		return
	}
	if roleInt < minRole {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权进行此操作，权限不足",
		})
		c.Abort()
		return
	}
	c.Set("username", username)
	c.Set("role", role)
	c.Set("id", id)
	if orgId != nil {
		c.Set("organization_id", orgId)
	}
	c.Next()
}

func UserAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		authHelper(c, model.RoleCommonUser)
	}
}

func AdminAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		authHelper(c, model.RoleAdminUser)
	}
}

func RootAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		authHelper(c, model.RoleRootUser)
	}
}

func TokenAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// 优先尝试 session 认证（admin 面板登录后可直接调 /v1）
		session := sessions.Default(c)
		if sid := session.Get("id"); sid != nil {
			if id, ok := sid.(int); ok && id > 0 {
				c.Set(ctxkey.Id, id)
				// 从 session 获取 request model（如果有）
				requestModel, err := getRequestModel(c)
				if err == nil {
					c.Set(ctxkey.RequestModel, requestModel)
				}
				c.Next()
				return
			}
		}

		key := c.Request.Header.Get("Authorization")
		key = strings.TrimPrefix(key, "Bearer ")
		key = strings.TrimPrefix(key, "sk-")
		parts := strings.Split(key, "-")
		key = parts[0]
		token, err := model.ValidateUserToken(key)
		if err != nil {
			abortWithMessage(c, http.StatusUnauthorized, err.Error())
			return
		}
		if token.Subnet != nil && *token.Subnet != "" {
			if !network.IsIpInSubnets(ctx, c.ClientIP(), *token.Subnet) {
				abortWithMessage(c, http.StatusForbidden, fmt.Sprintf("该令牌只能在指定网段使用：%s，当前 ip：%s", *token.Subnet, c.ClientIP()))
				return
			}
		}
		userEnabled, err := model.CacheIsUserEnabled(token.UserId)
		if err != nil {
			abortWithMessage(c, http.StatusInternalServerError, err.Error())
			return
		}
		if !userEnabled || blacklist.IsUserBanned(token.UserId) {
			abortWithMessage(c, http.StatusForbidden, "用户已被封禁")
			return
		}
		requestModel, err := getRequestModel(c)
		if err != nil && shouldCheckModel(c) {
			abortWithMessage(c, http.StatusBadRequest, err.Error())
			return
		}
		c.Set(ctxkey.RequestModel, requestModel)
		if token.Models != nil && *token.Models != "" {
			c.Set(ctxkey.AvailableModels, *token.Models)
			if requestModel != "" && !isModelInList(requestModel, *token.Models) {
				abortWithMessage(c, http.StatusForbidden, fmt.Sprintf("该令牌无权使用模型：%s", requestModel))
				return
			}
		}
		c.Set(ctxkey.Id, token.UserId)
		c.Set(ctxkey.TokenId, token.Id)
		c.Set(ctxkey.TokenName, token.Name)
		if len(parts) > 1 {
			if model.IsAdmin(token.UserId) {
				c.Set(ctxkey.SpecificChannelId, parts[1])
			} else {
				abortWithMessage(c, http.StatusForbidden, "普通用户不支持指定渠道")
				return
			}
		}

		// set channel id for proxy relay
		if channelId := c.Param("channelid"); channelId != "" {
			c.Set(ctxkey.SpecificChannelId, channelId)
		}

		c.Next()
	}
}

func shouldCheckModel(c *gin.Context) bool {
	if strings.HasPrefix(c.Request.URL.Path, "/v1/completions") {
		return true
	}
	if strings.HasPrefix(c.Request.URL.Path, "/v1/chat/completions") {
		return true
	}
	if strings.HasPrefix(c.Request.URL.Path, "/v1/images") {
		return true
	}
	if strings.HasPrefix(c.Request.URL.Path, "/v1/audio") {
		return true
	}
	if strings.HasPrefix(c.Request.URL.Path, "/v1/quantum") {
		return true
	}
	return false
}


