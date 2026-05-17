package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/controller"
	"github.com/quantumclaw/quantumclaw/model"
)

// TelegramUser represents the user data from Telegram Login Widget
type TelegramUser struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
	AuthDate  int64  `json:"auth_date"`
	Hash      string `json:"hash"`
}

// validateTelegramHash validates Telegram Login Widget data hash
func validateTelegramHash(tgUser *TelegramUser) error {
	if tgUser.Hash == "" {
		return errors.New("Telegram hash is empty")
	}

	// 1. Create secret key from bot token using SHA256
	botToken := config.TelegramBotToken
	if botToken == "" {
		return errors.New("Telegram bot token is not configured")
	}

	secretKey := sha256.Sum256([]byte(botToken))

	// 2. Build data-check-string: sort all fields alphabetically, key=value format
	// exclude "hash" field
	dataFields := map[string]string{
		"auth_date":  strconv.FormatInt(tgUser.AuthDate, 10),
		"first_name": tgUser.FirstName,
		"id":         strconv.Itoa(tgUser.ID),
	}

	if tgUser.LastName != "" {
		dataFields["last_name"] = tgUser.LastName
	}
	if tgUser.Username != "" {
		dataFields["username"] = tgUser.Username
	}
	if tgUser.PhotoURL != "" {
		dataFields["photo_url"] = tgUser.PhotoURL
	}

	// Sort keys alphabetically
	var keys []string
	for k := range dataFields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var pairs []string
	for _, k := range keys {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, dataFields[k]))
	}
	dataCheckString := strings.Join(pairs, "\n")

	// 3. Compute HMAC-SHA256
	mac := hmac.New(sha256.New, secretKey[:])
	mac.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(mac.Sum(nil))

	// 4. Compare
	if !hmac.Equal([]byte(expectedHash), []byte(tgUser.Hash)) {
		return errors.New("Telegram data hash validation failed")
	}

	// 5. Check auth_date is not too old (allow up to 1 day)
	now := time.Now().Unix()
	if now-tgUser.AuthDate > 86400 {
		return errors.New("Telegram auth data is expired")
	}

	return nil
}

// parseTelegramAuthData parses Telegram login widget query parameters
func parseTelegramAuthData(params url.Values) (*TelegramUser, error) {
	idStr := params.Get("id")
	if idStr == "" {
		return nil, errors.New("Telegram id is empty")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, errors.New("Invalid Telegram id")
	}

	authDateStr := params.Get("auth_date")
	authDate, err := strconv.ParseInt(authDateStr, 10, 64)
	if err != nil {
		return nil, errors.New("Invalid Telegram auth_date")
	}

	return &TelegramUser{
		ID:        id,
		FirstName: params.Get("first_name"),
		LastName:  params.Get("last_name"),
		Username:  params.Get("username"),
		PhotoURL:  params.Get("photo_url"),
		AuthDate:  authDate,
		Hash:      params.Get("hash"),
	}, nil
}

// TelegramOAuth handles Telegram login via Login Widget callback
func TelegramOAuth(c *gin.Context) {
	ctx := c.Request.Context()
	session := sessions.Default(c)

	// Check if this is a bind request
	username := session.Get("username")
	if username != nil {
		TelegramBind(c)
		return
	}

	if !config.TelegramOAuthEnabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "管理员未开启通过 Telegram 登录以及注册",
		})
		return
	}

	// Parse Telegram auth data from query params
	tgUser, err := parseTelegramAuthData(c.Request.URL.Query())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Validate hash
	if err := validateTelegramHash(tgUser); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	user := model.User{
		TelegramId: strconv.Itoa(tgUser.ID),
	}

	// Check if Telegram ID is already bound
	if model.IsTelegramIdAlreadyTaken(user.TelegramId) {
		// Existing user, login directly
		err := user.FillUserByTelegramId()
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
			user.Username = "telegram_" + strconv.Itoa(model.GetMaxUserId()+1)
			displayName := tgUser.FirstName
			if tgUser.LastName != "" {
				displayName = displayName + " " + tgUser.LastName
			}
			if displayName == "" {
				displayName = "Telegram User"
			}
			user.DisplayName = displayName
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

	if user.Status != model.UserStatusEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "用户已被封禁",
			"success": false,
		})
		return
	}

	controller.SetupLogin(&user, c)
}

// TelegramBind binds a Telegram account to an existing user
func TelegramBind(c *gin.Context) {
	if !config.TelegramOAuthEnabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "管理员未开启通过 Telegram 登录以及注册",
		})
		return
	}

	// Parse Telegram auth data
	tgUser, err := parseTelegramAuthData(c.Request.URL.Query())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Validate hash
	if err := validateTelegramHash(tgUser); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	user := model.User{
		TelegramId: strconv.Itoa(tgUser.ID),
	}

	// Check if Telegram ID is already taken by another user
	if model.IsTelegramIdAlreadyTaken(user.TelegramId) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "该 Telegram 账号已被绑定",
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

	// Bind Telegram ID
	user.TelegramId = strconv.Itoa(tgUser.ID)
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

// GenerateTelegramWidgetOptions returns the Telegram Login Widget data-* attributes
func GenerateTelegramWidgetOptions(c *gin.Context) {
	if !config.TelegramOAuthEnabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "管理员未开启通过 Telegram 登录以及注册",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"bot_username": config.TelegramBotUsername,
			"callback_url": config.ServerAddress + "/api/oauth/telegram",
		},
	})
}

// TelegramAuthHandler handles POST with JSON body (for client-side widget callback)
func TelegramAuthHandler(c *gin.Context) {
	ctx := c.Request.Context()
	session := sessions.Default(c)

	// Check if this is a bind request
	username := session.Get("username")
	if username != nil {
		// For bind, we redirect to bind logic
		TelegramBindHandler(c)
		return
	}

	if !config.TelegramOAuthEnabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "管理员未开启通过 Telegram 登录以及注册",
		})
		return
	}

	var tgUser TelegramUser
	if err := c.ShouldBindJSON(&tgUser); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的请求参数",
		})
		return
	}

	if err := validateTelegramHash(&tgUser); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	user := model.User{
		TelegramId: strconv.Itoa(tgUser.ID),
	}

	if model.IsTelegramIdAlreadyTaken(user.TelegramId) {
		err := user.FillUserByTelegramId()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	} else {
		if config.RegisterEnabled {
			user.Username = "telegram_" + strconv.Itoa(model.GetMaxUserId()+1)
			displayName := tgUser.FirstName
			if tgUser.LastName != "" {
				displayName = displayName + " " + tgUser.LastName
			}
			if displayName == "" {
				displayName = "Telegram User"
			}
			user.DisplayName = displayName
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

	if user.Status != model.UserStatusEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "用户已被封禁",
			"success": false,
		})
		return
	}

	controller.SetupLogin(&user, c)
}

// TelegramBindHandler handles POST bind request with JSON body
func TelegramBindHandler(c *gin.Context) {
	if !config.TelegramOAuthEnabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "管理员未开启通过 Telegram 登录以及注册",
		})
		return
	}

	var tgUser TelegramUser
	if err := c.ShouldBindJSON(&tgUser); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的请求参数",
		})
		return
	}

	if err := validateTelegramHash(&tgUser); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	user := model.User{
		TelegramId: strconv.Itoa(tgUser.ID),
	}

	if model.IsTelegramIdAlreadyTaken(user.TelegramId) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "该 Telegram 账号已被绑定",
		})
		return
	}

	// Get current user from session
	session := sessions.Default(c)
	id := session.Get("id")
	user.Id = id.(int)
	err := user.FillUserById()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	user.TelegramId = strconv.Itoa(tgUser.ID)
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
