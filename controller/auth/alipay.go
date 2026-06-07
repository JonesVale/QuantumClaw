package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

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

	// 构建支付宝 OAuth URL — 回调地址指向专门的回调处理端点
	redirectURI := fmt.Sprintf("%s/api/oauth/alipay/callback", config.ServerAddress)
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
		"redirect_uri":  {fmt.Sprintf("%s/api/oauth/alipay/callback", config.ServerAddress)},
	}

	httpClient := &http.Client{Timeout: 15 * time.Second}
	tokenResp, err := httpClient.PostForm("https://openid.alipay.com/oauth/token", tokenValues)
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

	userResp, err := httpClient.PostForm("https://openid.alipay.com/oauth/userinfo", userValues)
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
func AlipayAuthCallback(c *gin.Context) {
	ctx := c.Request.Context()
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

	// 查找或创建用户（复用 WeChatId 字段存储支付宝第三方 ID）
	alipayId := "alipay_" + userInfo.UserID
	user := model.User{
		WeChatId: alipayId,
	}

	if model.IsWeChatIdAlreadyTaken(user.WeChatId) {
		// 已存在用户，直接登录
		if err := user.FillUserByWeChatId(); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	} else {
		// 新用户注册
		if !config.RegisterEnabled {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "管理员关闭了新用户注册",
			})
			return
		}
		user.Username = "alipay_" + strconv.Itoa(model.GetMaxUserId()+1)
		displayName := userInfo.NickName
		if displayName == "" {
			displayName = user.Username
		}
		user.DisplayName = displayName
		user.Role = model.RoleCommonUser
		user.Status = model.UserStatusEnabled

		if err := user.Insert(ctx, 0); err != nil {
			logger.SysError("Alipay OAuth create user error: " + err.Error())
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	}

	if user.Status != model.UserStatusEnabled {
		logger.SysWarnf("user %d (%s) login with Alipay OAuth, status=%d", user.Id, user.Username, user.Status)
	}

	controller.SetupLogin(&user, c)
}
