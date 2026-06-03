package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/common/random"
	"github.com/quantumclaw/quantumclaw/controller"
	"github.com/quantumclaw/quantumclaw/model"
)

// DiscordOAuthResponse Discord OAuth 令牌响应
type DiscordOAuthResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// DiscordUser Discord 用户信息
type DiscordUser struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	Discriminator string `json:"discriminator"`
	GlobalName   string `json:"global_name"`
	Email        string `json:"email"`
	Verified     bool   `json:"verified"`
}

// getDiscordUserInfoByCode 用授权码获取 Discord 用户信息
func getDiscordUserInfoByCode(code string) (*DiscordUser, error) {
	if code == "" {
		return nil, errors.New("无效的参数")
	}

	// 1. 用 code 换 access token
	values := map[string]string{
		"client_id":     config.DiscordClientId,
		"client_secret": config.DiscordClientSecret,
		"grant_type":    "authorization_code",
		"code":          code,
		"redirect_uri":  config.ServerAddress + "/api/oauth/discord/callback",
	}
	jsonData, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://discord.com/api/oauth2/token", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		logger.SysLog(err.Error())
		return nil, errors.New("无法连接至 Discord 服务器，请稍后重试！")
	}
	defer res.Body.Close()

	var tokenResp DiscordOAuthResponse
	err = json.NewDecoder(res.Body).Decode(&tokenResp)
	if err != nil {
		return nil, err
	}
	if tokenResp.AccessToken == "" {
		return nil, errors.New("获取 Discord 访问令牌失败")
	}

	// 2. 用 access token 获取用户信息
	req2, err := http.NewRequest("GET", "https://discord.com/api/users/@me", nil)
	if err != nil {
		return nil, err
	}
	req2.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenResp.AccessToken))
	req2.Header.Set("Accept", "application/json")

	res2, err := client.Do(req2)
	if err != nil {
		logger.SysLog(err.Error())
		return nil, errors.New("无法连接至 Discord 服务器，请稍后重试！")
	}
	defer res2.Body.Close()

	var discordUser DiscordUser
	err = json.NewDecoder(res2.Body).Decode(&discordUser)
	if err != nil {
		return nil, err
	}
	if discordUser.ID == "" {
		return nil, errors.New("Discord 返回值非法，用户 ID 为空，请稍后重试！")
	}

	return &discordUser, nil
}

// DiscordOAuth Discord OAuth 登录入口
func DiscordOAuth(c *gin.Context) {
	ctx := c.Request.Context()
	session := sessions.Default(c)
	state := c.Query("state")
	if state == "" || session.Get("oauth_state") == nil || state != session.Get("oauth_state").(string) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "state 为空或不匹配",
		})
		return
	}

	// 检查是否是绑定请求
	username := session.Get("username")
	if username != nil {
		DiscordBind(c)
		return
	}

	if !config.DiscordOAuthEnabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "管理员未开启通过 Discord 登录以及注册",
		})
		return
	}

	code := c.Query("code")
	discordUser, err := getDiscordUserInfoByCode(code)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	user := model.User{
		DiscordId: discordUser.ID,
	}

	// 检查 Discord ID 是否已被绑定
	if model.IsDiscordIdAlreadyTaken(user.DiscordId) {
		// 已存在用户，直接登录
		err := user.FillUserByDiscordId()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	} else {
		// 新用户注册
		if config.RegisterEnabled {
			user.Username = "discord_" + strconv.Itoa(model.GetMaxUserId()+1)
			displayName := discordUser.GlobalName
			if displayName == "" {
				displayName = discordUser.Username
			}
			user.DisplayName = displayName
			user.Email = discordUser.Email
			user.Role = model.RoleCommonUser
			user.Status = model.UserStatusEnabled

			if err := user.Insert(ctx, 0); err != nil {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": err.Error(),
				})
				return
			}
		} else {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "管理员关闭了新用户注册",
			})
			return
		}
	}

	// 无论状态都允许登录
	if user.Status != model.UserStatusEnabled {
		logger.SysWarnf("user %d (%s) login with Discord OAuth, status=%d", user.Id, user.Username, user.Status)
	}

	controller.SetupLogin(&user, c)
}

// DiscordBind 将 Discord 账号绑定到已有用户
func DiscordBind(c *gin.Context) {
	if !config.DiscordOAuthEnabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "管理员未开启通过 Discord 登录以及注册",
		})
		return
	}

	code := c.Query("code")
	discordUser, err := getDiscordUserInfoByCode(code)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	user := model.User{
		DiscordId: discordUser.ID,
	}

	// 检查 Discord ID 是否已被其他用户绑定
	if model.IsDiscordIdAlreadyTaken(user.DiscordId) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "该 Discord 账号已被绑定",
		})
		return
	}

	// 获取当前登录用户
	session := sessions.Default(c)
	id := session.Get("id")
	user.Id = id.(int)
	err = user.FillUserById()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// 绑定 Discord ID
	user.DiscordId = discordUser.ID
	err = user.Update(false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "bind",
	})
}

// GenerateDiscordOAuthURL 生成 Discord 授权 URL（供前端调用）
func GenerateDiscordOAuthURL(c *gin.Context) {
	session := sessions.Default(c)
	state := random.GetRandomString(12)
	session.Set("oauth_state", state)
	err := session.Save()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// 构建 Discord 授权 URL
	authURL := fmt.Sprintf(
		"https://discord.com/api/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=identify+email&state=%s",
		config.DiscordClientId,
		config.ServerAddress+"/api/oauth/discord/callback",
		state,
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    authURL,
	})
}
