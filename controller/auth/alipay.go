package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/controller"
	"github.com/quantumclaw/quantumclaw/model"
)

// AlipayOAuthResponse 支付宝 OAuth Token 响应
type AlipayOAuthResponse struct {
	AccessToken  string `json:"access_token"`
	AlipayUserID string `json:"alipay_user_id"`
	UserID       string `json:"user_id"`
	ExpiresIn    int    `json:"expires_in"`
}

// AlipayUserInfo 支付宝用户信息
type AlipayUserInfo struct {
	NickName string `json:"nick_name"`
	Avatar   string `json:"avatar"`
	UserID   string `json:"user_id"`
}

// AlipayAuth 支付宝授权登录入口
// GET /api/oauth/alipay
func AlipayAuth(c *gin.Context) {
	if config.AlipayAppId == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "支付宝登录未配置",
		})
		return
	}

	// 生成 state 防 CSRF
	state := c.Query("state")
	if state == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}

	// 构建支付宝 OAuth URL
	redirectURI := fmt.Sprintf("%s/api/oauth/alipay", config.ServerAddress)
	authURL := fmt.Sprintf(
		"https://openauth.alipay.com/oauth2/publicAppAuthorize.htm?app_id=%s&scope=auth_user&redirect_uri=%s&state=%s",
		config.AlipayAppId,
		url.QueryEscape(redirectURI),
		url.QueryEscape(state),
	)

	c.Redirect(http.StatusFound, authURL)
}

// processAlipayCallback 处理支付宝回调
func processAlipayCallback(code string) (*AlipayUserInfo, error) {
	if code == "" {
		return nil, errors.New("无效的参数")
	}

	// 1. 用 code 换取 access_token
	tokenValues := url.Values{
		"client_id":     {config.AlipayAppId},
		"client_secret": {config.AlipayPrivateKey},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {fmt.Sprintf("%s/api/oauth/alipay", config.ServerAddress)},
	}

	tokenResp, err := http.PostForm("https://openid.alipay.com/oauth/token", tokenValues)
	if err != nil {
		return nil, err
	}
	defer tokenResp.Body.Close()

	var tokenRes AlipayOAuthResponse
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokenRes); err != nil {
		return nil, err
	}

	if tokenRes.AccessToken == "" {
		return nil, errors.New("获取支付宝 access_token 失败")
	}

	// 2. 用 access_token 获取用户信息
	userValues := url.Values{
		"access_token": {tokenRes.AccessToken},
		"client_id":    {config.AlipayAppId},
	}

	userResp, err := http.PostForm("https://openid.alipay.com/oauth/userinfo", userValues)
	if err != nil {
		return nil, err
	}
	defer userResp.Body.Close()

	var userInfo AlipayUserInfo
	if err := json.NewDecoder(userResp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	if userInfo.UserID == "" {
		return nil, errors.New("获取支付宝用户信息失败")
	}

	return &userInfo, nil
}

// AlipayAuthCallback 支付宝授权回调处理
// 请求会直接打到 /api/oauth/alipay 同一个路由（GET）
// 区别在于有无 code 参数
func AlipayAuthCallback(c *gin.Context) {
	code := c.Query("auth_code") // 支付宝回调参数是 auth_code
	if code == "" {
		code = c.Query("code")
	}
	if code == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的授权码",
		})
		return
	}

	// 获取用户信息
	userInfo, err := processAlipayCallback(code)
	if err != nil {
		logger.SysError("Alipay OAuth error: " + err.Error())
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// 查找或创建用户
	alipayId := "alipay_" + userInfo.UserID
	user, err := model.GetUserByWechatID(alipayId) // 复用 wechat_id 字段存储第三方ID
	if err != nil {
		// 创建新用户
		username := "alipay_" + userInfo.UserID
		if len(username) > 20 {
			username = username[:20]
		}
		displayName := userInfo.NickName
		if displayName == "" {
			displayName = username
		}

		user = &model.User{
			Username:    username,
			DisplayName: displayName,
			WeChatId:    alipayId,
			Role:        1,
			Status:      1,
		}

		if err := model.CreateUser(user); err != nil {
			logger.SysError("Alipay OAuth create user error: " + err.Error())
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "创建用户失败",
			})
			return
		}
	}

	// 设置 session
	session := sessions.Default(c)
	session.Set("id", user.Id)
	session.Set("username", user.Username)
	session.Set("role", user.Role)
	session.Set("status", user.Status)

	if err := session.Save(); err != nil {
		logger.SysError("Alipay OAuth session save error: " + err.Error())
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "session 保存失败",
		})
		return
	}

	// 登录成功后重定向
	controller.SetLoginCookie(c, user.Id)
	c.Redirect(http.StatusFound, fmt.Sprintf("%s/login?oauth=alipay", config.ServerAddress))
}
