package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/controller"
	"github.com/quantumclaw/quantumclaw/model"
)

// webauthnInstance 缓存 WebAuthn 实例
var webauthnInstance *webauthn.WebAuthn
var webauthnInitError error

// getWebAuthn 获取或初始化 WebAuthn 实例
func getWebAuthn() (*webauthn.WebAuthn, error) {
	if webauthnInstance != nil || webauthnInitError != nil {
		return webauthnInstance, webauthnInitError
	}
	rpID := config.WebAuthnRPID
	origin := config.WebAuthnOrigin

	// 自动从 ServerAddress 提取 rpID 和 origin
	if rpID == "" || origin == "" {
		addr := config.ServerAddress
		addr = strings.TrimPrefix(addr, "https://")
		addr = strings.TrimPrefix(addr, "http://")
		// 去掉路径部分
		if idx := strings.Index(addr, "/"); idx != -1 {
			addr = addr[:idx]
		}
		if rpID == "" {
			rpID = addr
		}
		if origin == "" {
			origin = addr
		}
	}

	wconfig := &webauthn.Config{
		RPDisplayName: config.WebAuthnRPDisplayName,
		RPID:          rpID,
		RPOrigin:      origin,
	}

	w, err := webauthn.New(wconfig)
	webauthnInstance = w
	webauthnInitError = err
	return w, err
}

// webauthnUser 实现 webauthn.User 接口
type webauthnUser struct {
	user  *model.User
	creds []webauthn.Credential
}

func (u *webauthnUser) WebAuthnID() []byte {
	return []byte(strconv.Itoa(u.user.Id))
}

func (u *webauthnUser) WebAuthnName() string {
	return u.user.Username
}

func (u *webauthnUser) WebAuthnDisplayName() string {
	if u.user.DisplayName != "" {
		return u.user.DisplayName
	}
	return u.user.Username
}

func (u *webauthnUser) WebAuthnCredentials() []webauthn.Credential {
	return u.creds
}

func (u *webauthnUser) WebAuthnIcon() string {
	return ""
}

// loadUserCredentials 从数据库加载用户的凭证
func loadUserCredentials(user *model.User) (*webauthnUser, error) {
	wuser := &webauthnUser{user: user}
	credentials, err := model.GetUserCredentials(user.Id)
	if err != nil {
		return wuser, nil // 没有凭证不是错误
	}
	for _, cred := range credentials {
		c, err := cred.ToWebAuthnCredential()
		if err != nil {
			logger.SysLog("failed to parse credential for user " + strconv.Itoa(user.Id) + ": " + err.Error())
			continue
		}
		wuser.creds = append(wuser.creds, *c)
	}
	return wuser, nil
}

// sessionDataKey session 中存储 SessionData 的 key
const sessionDataKeyReg = "webauthn_reg_session"
const sessionDataKeyAuth = "webauthn_auth_session"
const sessionUserIDKey = "webauthn_auth_user_id"

// saveSessionData 将 SessionData 序列化后存到 session
func saveSessionData(session sessions.Session, key string, data *webauthn.SessionData) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	session.Set(key, base64.StdEncoding.EncodeToString(b))
	return nil
}

// loadSessionData 从 session 加载 SessionData
func loadSessionData(session sessions.Session, key string) (*webauthn.SessionData, error) {
	v := session.Get(key)
	if v == nil {
		return nil, errors.New("session data not found")
	}
	s, ok := v.(string)
	if !ok {
		return nil, errors.New("invalid session data type")
	}
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	var data webauthn.SessionData
	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// clearSessionData 清除 session 中的 SessionData
func clearSessionData(session sessions.Session, keys ...string) {
	for _, key := range keys {
		session.Delete(key)
	}
}

// WebAuthnBeginRegistration 开始 Passkey 注册流程
func WebAuthnBeginRegistration(c *gin.Context) {
	if !config.WebAuthnEnabled {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "WebAuthn 未启用"})
		return
	}

	session := sessions.Default(c)
	userID := session.Get("id")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "请先登录"})
		return
	}

	user, err := model.GetUserById(userID.(int), false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	w, err := getWebAuthn()
	if err != nil {
		logger.SysError("webauthn init failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "WebAuthn 初始化失败"})
		return
	}

	wuser, _ := loadUserCredentials(user)
	options, sessionData, err := w.BeginRegistration(wuser)
	if err != nil {
		logger.SysError("webauthn begin registration failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "注册初始化失败"})
		return
	}

	// 将 sessionData 存入 session
	if err := saveSessionData(session, sessionDataKeyReg, sessionData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "保存会话失败"})
		return
	}
	session.Save()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    options,
	})
}

// WebAuthnFinishRegistration 完成 Passkey 注册
func WebAuthnFinishRegistration(c *gin.Context) {
	if !config.WebAuthnEnabled {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "WebAuthn 未启用"})
		return
	}

	session := sessions.Default(c)
	userID := session.Get("id")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "请先登录"})
		return
	}

	// 解析请求体
	var req struct {
		DeviceName string                 `json:"device_name"`
		Response   map[string]interface{} `json:"response"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的请求"})
		return
	}

	user, err := model.GetUserById(userID.(int), false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	w, err := getWebAuthn()
	if err != nil {
		logger.SysError("webauthn init failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "WebAuthn 初始化失败"})
		return
	}

	// 从 session 获取 sessionData
	sessionData, err := loadSessionData(session, sessionDataKeyReg)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "注册会话已过期，请重新开始"})
		return
	}

	wuser, _ := loadUserCredentials(user)
	credential, err := w.FinishRegistration(wuser, *sessionData, c.Request)
	if err != nil {
		logger.SysError("create credential failed: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "凭证验证失败: " + err.Error()})
		return
	}

	// 保存到数据库
	dbCred := &model.WebAuthnCredential{
		UserID:     user.Id,
		DeviceName: req.DeviceName,
	}
	if err := dbCred.FromWebAuthnCredential(credential); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "保存凭证失败"})
		return
	}

	if err := model.DB.Create(dbCred).Error; err != nil {
		logger.SysError("save credential failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "保存凭证失败"})
		return
	}

	// 清除 session
	clearSessionData(session, sessionDataKeyReg)
	session.Save()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Passkey 注册成功",
	})
}

// WebAuthnBeginAuthentication 开始 Passkey 认证（登录）
func WebAuthnBeginAuthentication(c *gin.Context) {
	if !config.WebAuthnEnabled {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "WebAuthn 未启用"})
		return
	}

	var req struct {
		Username string `json:"username"`
	}
	c.ShouldBindJSON(&req)

	var user *model.User
	if req.Username != "" {
		user = &model.User{Username: req.Username}
		err := user.FillUserByUsername()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "用户不存在"})
			return
		}
	} else {
		// 已经登录的用户也可以发起认证（用于验证身份）
		session := sessions.Default(c)
		uid := session.Get("id")
		if uid == nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请提供用户名或先登录"})
			return
		}
		user, _ = model.GetUserById(uid.(int), false)
	}

	if user == nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "用户不存在"})
		return
	}

	w, err := getWebAuthn()
	if err != nil {
		logger.SysError("webauthn init failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "WebAuthn 初始化失败"})
		return
	}

	wuser, _ := loadUserCredentials(user)
	options, sessionData, err := w.BeginLogin(wuser)
	if err != nil {
		logger.SysError("webauthn begin login failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "认证初始化失败"})
		return
	}

	// 将 sessionData 存入 session
	session := sessions.Default(c)
	if err := saveSessionData(session, sessionDataKeyAuth, sessionData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "保存会话失败"})
		return
	}
	session.Set(sessionUserIDKey, user.Id)
	session.Save()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    options,
	})
}

// WebAuthnFinishAuthentication 完成 Passkey 认证（登录）
func WebAuthnFinishAuthentication(c *gin.Context) {
	if !config.WebAuthnEnabled {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "WebAuthn 未启用"})
		return
	}

	session := sessions.Default(c)

	// 从 session 获取用户 ID 和 sessionData
	userID := session.Get(sessionUserIDKey)
	if userID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "认证会话已过期，请重新开始"})
		return
	}

	sessionData, err := loadSessionData(session, sessionDataKeyAuth)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "认证会话已过期，请重新开始"})
		return
	}

	user, err := model.GetUserById(userID.(int), false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	w, err := getWebAuthn()
	if err != nil {
		logger.SysError("webauthn init failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "WebAuthn 初始化失败"})
		return
	}

	wuser, _ := loadUserCredentials(user)
	credential, err := w.FinishLogin(wuser, *sessionData, c.Request)
	if err != nil {
		logger.SysError("validate login failed: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "认证验证失败: " + err.Error()})
		return
	}

	// 更新签名计数
	credID := base64.RawURLEncoding.EncodeToString(credential.ID)
	model.UpdateCredentialSignCount(credID, credential.Authenticator.SignCount, credential.Authenticator.CloneWarning)

	// 清除 session
	clearSessionData(session, sessionDataKeyAuth, sessionUserIDKey)
	
	// 标记 WebAuthn 已验证
	session.Set("webauthn_verified", true)
	session.Save()

	// 登录用户
	controller.SetupLogin(user, c)
}

// WebAuthnGetCredentials 获取当前用户的 Passkey 列表
func WebAuthnGetCredentials(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("id")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "请先登录"})
		return
	}

	credentials, err := model.GetUserCredentials(userID.(int))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    credentials,
	})
}

// WebAuthnDeleteCredential 删除用户的某个 Passkey
func WebAuthnDeleteCredential(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("id")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "请先登录"})
		return
	}

	credentialID := c.Param("id")
	if credentialID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的凭证 ID"})
		return
	}

	err := model.DeleteWebAuthnCredential(userID.(int), credentialID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  "Passkey 已删除",
	})
}
