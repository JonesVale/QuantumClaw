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

// LinuxDOOAuthResponse LinuxDO OAuth token response
type LinuxDOOAuthResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// LinuxDOUser LinuxDO user info
type LinuxDOUser struct {
	ID          int    `json:"id"`
	Username    string `json:"username"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	AvatarURL   string `json:"avatar_url"`
}

// getLinuxDOUserInfoByCode exchanges auth code for token and fetches user profile
func getLinuxDOUserInfoByCode(code string) (*LinuxDOUser, error) {
	if code == "" {
		return nil, errors.New("invalid parameter")
	}

	// 1. Exchange code for access token
	values := map[string]string{
		"client_id":     config.LinuxDOClientId,
		"client_secret": config.LinuxDOClientSecret,
		"grant_type":    "authorization_code",
		"code":          code,
		"redirect_uri":  config.ServerAddress + "/api/oauth/linuxdo",
	}
	jsonData, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://connect.linux.do/oauth/token", bytes.NewBuffer(jsonData))
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
		return nil, errors.New("cannot connect to LinuxDO server, please try again later!")
	}
	defer res.Body.Close()

	var tokenResp LinuxDOOAuthResponse
	err = json.NewDecoder(res.Body).Decode(&tokenResp)
	if err != nil {
		return nil, err
	}
	if tokenResp.AccessToken == "" {
		return nil, errors.New("failed to obtain LinuxDO access token")
	}

	// 2. Fetch user info with access token
	req2, err := http.NewRequest("GET", "https://connect.linux.do/api/user", nil)
	if err != nil {
		return nil, err
	}
	req2.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenResp.AccessToken))
	req2.Header.Set("Accept", "application/json")

	res2, err := client.Do(req2)
	if err != nil {
		logger.SysLog(err.Error())
		return nil, errors.New("cannot connect to LinuxDO server, please try again later!")
	}
	defer res2.Body.Close()

	var linuxdoUser LinuxDOUser
	err = json.NewDecoder(res2.Body).Decode(&linuxdoUser)
	if err != nil {
		return nil, err
	}
	if linuxdoUser.ID == 0 {
		return nil, errors.New("invalid LinuxDO response, user ID is empty!")
	}

	return &linuxdoUser, nil
}

// LinuxDoOAuth LinuxDO OAuth login entry point
func LinuxDoOAuth(c *gin.Context) {
	ctx := c.Request.Context()
	session := sessions.Default(c)
	state := c.Query("state")
	if state == "" || session.Get("oauth_state") == nil || state != session.Get("oauth_state").(string) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "state is empty or does not match",
		})
		return
	}

	// Check bind request
	username := session.Get("username")
	if username != nil {
		LinuxDoBind(c)
		return
	}

	if !config.LinuxDOOAuthEnabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Admin has not enabled LinuxDO login/registration",
		})
		return
	}

	code := c.Query("code")
	linuxdoUser, err := getLinuxDOUserInfoByCode(code)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	user := model.User{
		LinuxDOId: strconv.Itoa(linuxdoUser.ID),
	}

	// Check if LinuxDO ID is already bound
	if model.IsLinuxDOIdAlreadyTaken(user.LinuxDOId) {
		// Existing user, directly login
		err := user.FillUserByLinuxDOId()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	} else {
		// New user registration
		if config.RegisterEnabled {
			user.Username = "linuxdo_" + strconv.Itoa(model.GetMaxUserId()+1)
			displayName := linuxdoUser.Name
			if displayName == "" {
				displayName = linuxdoUser.Username
			}
			user.DisplayName = displayName
			user.Email = linuxdoUser.Email
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
				"message": "Admin has disabled new user registration",
			})
			return
		}
	}

	if user.Status != model.UserStatusEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "User has been banned",
			"success": false,
		})
		return
	}

	controller.SetupLogin(&user, c)
}

// LinuxDoBind binds LinuxDO account to an existing user
func LinuxDoBind(c *gin.Context) {
	if !config.LinuxDOOAuthEnabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Admin has not enabled LinuxDO login/registration",
		})
		return
	}

	code := c.Query("code")
	linuxdoUser, err := getLinuxDOUserInfoByCode(code)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	user := model.User{
		LinuxDOId: strconv.Itoa(linuxdoUser.ID),
	}

	// Check if LinuxDO ID is already taken by another user
	if model.IsLinuxDOIdAlreadyTaken(user.LinuxDOId) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "This LinuxDO account is already bound",
		})
		return
	}

	// Get current logged-in user
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

	// Bind LinuxDO ID
	user.LinuxDOId = strconv.Itoa(linuxdoUser.ID)
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

// GenerateLinuxDOAuthURL generates the LinuxDO authorization URL for the frontend
func GenerateLinuxDOAuthURL(c *gin.Context) {
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

	// Build LinuxDO authorization URL
	authURL := fmt.Sprintf(
		"https://connect.linux.do/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=read&state=%s",
		config.LinuxDOClientId,
		config.ServerAddress+"/api/oauth/linuxdo",
		state,
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    authURL,
	})
}
