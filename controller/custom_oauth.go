package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/common/random"
	"github.com/quantumclaw/quantumclaw/model"
)

// ==================== 自定义 OAuth 提供商管理 API ====================

// ListCustomOAuthProviders 获取所有自定义 OAuth 提供商
func ListCustomOAuthProviders(c *gin.Context) {
	providers, err := model.GetAllCustomOAuthProviders()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": providers})
}

// CreateCustomOAuthProvider 创建自定义 OAuth 提供商
func CreateCustomOAuthProvider(c *gin.Context) {
	var req model.CustomOAuthProvider
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误: " + err.Error()})
		return
	}
	if req.Name == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "提供商名称不能为空"})
		return
	}
	if req.AuthURL == "" || req.TokenURL == "" || req.UserInfoURL == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "AuthURL、TokenURL、UserInfoURL 均不能为空"})
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	err := model.UpsertCustomOAuthProvider(&req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// UpdateCustomOAuthProvider 更新自定义 OAuth 提供商
func UpdateCustomOAuthProvider(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的 ID"})
		return
	}
	existing, err := model.GetCustomOAuthProviderById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "提供商不存在"})
		return
	}

	// 部分更新：只更新请求中提供的字段
	var req model.CustomOAuthProvider
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误: " + err.Error()})
		return
	}
	if req.DisplayName != "" {
		existing.DisplayName = req.DisplayName
	}
	if req.AuthURL != "" {
		existing.AuthURL = req.AuthURL
	}
	if req.TokenURL != "" {
		existing.TokenURL = req.TokenURL
	}
	if req.UserInfoURL != "" {
		existing.UserInfoURL = req.UserInfoURL
	}
	if req.ClientId != "" {
		existing.ClientId = req.ClientId
	}
	// ClientSecret 只在非空时更新（防止覆盖已有密钥）
	if req.ClientSecret != "" {
		existing.ClientSecret = req.ClientSecret
	}
	if req.Scopes != "" {
		existing.Scopes = req.Scopes
	}
	if req.UserIdField != "" {
		existing.UserIdField = req.UserIdField
	}
	if req.UsernameField != "" {
		existing.UsernameField = req.UsernameField
	}
	if req.EmailField != "" {
		existing.EmailField = req.EmailField
	}
	if req.LogoURL != "" {
		existing.LogoURL = req.LogoURL
	}
	if req.ButtonColor != "" {
		existing.ButtonColor = req.ButtonColor
	}
	existing.Enabled = req.Enabled
	existing.SortOrder = req.SortOrder

	if err := model.DB.Save(existing).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// DeleteCustomOAuthProvider 删除自定义 OAuth 提供商
func DeleteCustomOAuthProvider(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的 ID"})
		return
	}
	err = model.DB.Delete(&model.CustomOAuthProvider{}, "id = ?", id).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ==================== 自定义 OAuth 登录流程 ====================

// CustomOAuthLogin 发起自定义 OAuth 登录
func CustomOAuthLogin(c *gin.Context) {
	name := strings.TrimSpace(c.Param("name"))
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "提供商名称不能为空"})
		return
	}

	provider, err := model.GetOAuthProviderByName(name)
	if err != nil || !provider.Enabled {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "未找到或未启用该 OAuth 提供商"})
		return
	}

	// 构建授权 URL
	callbackURL := getBaseURL(c) + "/api/oauth/custom/" + provider.Name + "/callback"
	state := model.CreateOAuthState(name, callbackURL)
	redirectURL := buildOAuthAuthorizationURL(provider, state, callbackURL)

	// 重定向到授权页面
	c.Redirect(http.StatusMovedPermanently, redirectURL)
}

// CustomOAuthCallback 处理自定义 OAuth 回调
func CustomOAuthCallback(c *gin.Context) {
	name := strings.TrimSpace(c.Param("name"))
	state := c.Query("state")
	code := c.Query("code")
	oauthErr := c.Query("error")

	if oauthErr != "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "用户拒绝了授权"})
		return
	}
	if state == "" || code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "state 或 code 为空"})
		return
	}

	// 验证 state（CSRF 防护）
	stateProvider, _, valid := model.ValidateOAuthState(state)
	if !valid || stateProvider != name {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "state 验证失败或已过期，请重试"})
		return
	}

	provider, err := model.GetOAuthProviderByName(name)
	if err != nil || !provider.Enabled {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "OAuth 提供商不可用"})
		return
	}

	// 交换 access token
	callbackURL := getBaseURL(c) + "/api/oauth/custom/" + provider.Name + "/callback"
	tokenResp, err := exchangeOAuthToken(provider, code, callbackURL)
	if err != nil {
		logger.SysError("CustomOAuthCallback token exchange failed: " + err.Error())
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "获取访问令牌失败，请稍后重试"})
		return
	}

	// 获取用户信息
	userInfo, err := fetchOAuthUserInfo(provider, tokenResp)
	if err != nil {
		logger.SysError("CustomOAuthCallback userinfo failed: " + err.Error())
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "获取用户信息失败"})
		return
	}

	if userInfo.UserId == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "OAuth 提供商未返回用户 ID"})
		return
	}

	// 构造唯一 OAuth 标识：providerName:externalId
	oauthID := name + ":" + userInfo.UserId

	if model.IsCustomOAuthIdAlreadyTaken(oauthID) {
		// 已存在用户，直接登录
		var user model.User
		user.CustomOAuthId = oauthID
		if err := user.FillUserByCustomOAuthId(); err != nil {
			logger.SysError("FillUserByCustomOAuthId failed: " + err.Error())
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "用户查询失败"})
			return
		}
		if user.Status != model.UserStatusEnabled {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "用户已被封禁"})
			return
		}
		SetupLogin(&user, c)
		return
	}

	// 新用户注册
	if !config.RegisterEnabled {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "管理员关闭了新用户注册"})
		return
	}

	username := name + "_" + userInfo.UserId
	if userInfo.Username != "" {
		username = userInfo.Username
	}
	// 确保用户名唯一
	if model.IsUsernameAlreadyTaken(username) {
		username = username + "_" + random.GetRandomString(8)
	}

	user := model.User{
		Username:       username,
		DisplayName:   userInfo.Name,
		Email:         userInfo.Email,
		Role:          model.RoleCommonUser,
		Status:        model.UserStatusEnabled,
		CustomOAuthId: oauthID,
	}

	ctx := c.Request.Context()
	if err := user.Insert(ctx, 0); err != nil {
		logger.SysError("CustomOAuth user insert failed: " + err.Error())
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "创建用户失败"})
		return
	}
	SetupLogin(&user, c)
}

// ==================== 辅助函数 ====================

// buildOAuthAuthorizationURL 构建 OAuth 2.0 授权 URL
func buildOAuthAuthorizationURL(provider *model.CustomOAuthProvider, state, callbackURL string) string {
	base := strings.TrimRight(provider.AuthURL, "?")
	sep := "?"
	if strings.Contains(base, "?") {
		sep = "&"
	}
	params := url.Values{
		"client_id":    {provider.ClientId},
		"redirect_uri": {callbackURL},
		"response_type": {"code"},
		"scope":        {provider.Scopes},
		"state":        {state},
	}
	return base + sep + params.Encode()
}

// exchangeOAuthToken 用授权码换取 Access Token
func exchangeOAuthToken(provider *model.CustomOAuthProvider, code, callbackURL string) (*model.OAuthTokenResponse, error) {
	payload := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {provider.ClientId},
		"client_secret": {provider.ClientSecret},
		"code":          {code},
		"redirect_uri":  {callbackURL},
	}

	req, err := http.NewRequest("POST", provider.TokenURL, strings.NewReader(payload.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tokenResp model.OAuthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}
	if tokenResp.Error != "" {
		return nil, fmt.Errorf("oauth error: %s - %s", tokenResp.Error, tokenResp.ErrorDesc)
	}
	return &tokenResp, nil
}

// fetchOAuthUserInfo 从用户信息端点获取用户资料
func fetchOAuthUserInfo(provider *model.CustomOAuthProvider, tokenResp *model.OAuthTokenResponse) (*model.OAuthUserInfo, error) {
	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("access token is empty")
	}

	req, err := http.NewRequest("GET", provider.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	return model.ParseOAuthUserInfo(raw, provider), nil
}

// getBaseURL 从请求中获取服务器基础 URL
func getBaseURL(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	host := c.GetHeader("X-Forwarded-Host")
	if host == "" {
		host = c.GetHeader("Host")
	}
	if host == "" {
		host = c.Request.Host
	}
	return scheme + "://" + host
}
