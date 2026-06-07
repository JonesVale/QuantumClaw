package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/ctxkey"
	"github.com/quantumclaw/quantumclaw/common/i18n"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/common/random"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/relay/channeltype"
	"github.com/quantumclaw/quantumclaw/common/blacklist"
	"gorm.io/gorm"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(c *gin.Context) {
	if !config.PasswordLoginEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "管理员关闭了密码登录",
			"success": false,
		})
		return
	}
	var loginRequest LoginRequest
	err := json.NewDecoder(c.Request.Body).Decode(&loginRequest)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": i18n.Translate(c, "invalid_parameter"),
			"success": false,
		})
		return
	}
	username := loginRequest.Username
	password := loginRequest.Password
	if username == "" || password == "" {
		c.JSON(http.StatusOK, gin.H{
			"message": i18n.Translate(c, "invalid_parameter"),
			"success": false,
		})
		return
	}
	user := model.User{
		Username: username,
		Password: password,
	}
	err = user.ValidateAndFill()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"success": false,
		})
		return
	}
	SetupLogin(&user, c)
}

// setup session & cookies and then return user info
func SetupLogin(user *model.User, c *gin.Context) {
	session := sessions.Default(c)
	session.Set("id", user.Id)
	session.Set("username", user.Username)
	session.Set("role", user.Role)
	session.Set("status", user.Status)
	session.Set("organization_id", user.OrganizationID)
	session.Set("login_time", time.Now().Unix())
	err := session.Save()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "无法保存会话信息，请重试",
			"success": false,
		})
		return
	}
	// Get unread notification count
	unreadCount, _ := model.GetUnreadNotificationCount(user.Id)

	c.JSON(http.StatusOK, gin.H{
		"message": "",
		"success": true,
		"data": gin.H{
			"id":               user.Id,
			"username":         user.Username,
			"display_name":     user.DisplayName,
			"email":            user.Email,
			"avatar_url":       user.AvatarURL,
			"role":             user.Role,
			"status":           user.Status,
			"organization_id":  user.OrganizationID,
			"unread_count":     unreadCount,
			"quota_for_new_user": user.Quota > 0,
			"trial_balance":    user.CashBalance,
		},
	})
}

func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	err := session.Save()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"success": false,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "",
		"success": true,
	})
}

// ValidatePasswordStrength checks password strength and returns an error message if weak
// 强度要求通过 config.PasswordMinLength / config.PasswordRequireUpper 等配置控制
func ValidatePasswordStrength(password string) string {
	if len(password) < config.PasswordMinLength {
		return fmt.Sprintf("密码长度至少%d位", config.PasswordMinLength)
	}
	if config.PasswordRequireUpper {
		hasUpper := false
		for _, c := range password {
			if c >= 'A' && c <= 'Z' {
				hasUpper = true
				break
			}
		}
		if !hasUpper {
			return "密码需要包含大写字母"
		}
	}
	if config.PasswordRequireNumber {
		hasNumber := false
		for _, c := range password {
			if c >= '0' && c <= '9' {
				hasNumber = true
				break
			}
		}
		if !hasNumber {
			return "密码需要包含数字"
		}
	}
	if config.PasswordRequireSpecial {
		hasSpecial := false
		for _, c := range password {
			if c >= 'a' && c <= 'z' {
				continue
			}
			if c >= 'A' && c <= 'Z' {
				continue
			}
			if c >= '0' && c <= '9' {
				continue
			}
			hasSpecial = true
			break
		}
		if !hasSpecial {
			return "密码需要包含特殊字符"
		}
	}
	return ""
}

func Register(c *gin.Context) {
	ctx := c.Request.Context()
	if !config.RegisterEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "管理员关闭了新用户注册",
			"success": false,
		})
		return
	}
	if !config.PasswordRegisterEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "管理员关闭了通过密码进行注册，请使用第三方账户验证的形式进行注册",
			"success": false,
		})
		return
	}
	var user model.User
	err := json.NewDecoder(c.Request.Body).Decode(&user)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_parameter"),
		})
		return
	}
	if err := common.Validate.Struct(&user); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_input"),
		})
		return
	}
	// Enforce password strength policy
	if msg := ValidatePasswordStrength(user.Password); msg != "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": msg,
		})
		return
	}
	if config.EmailVerificationEnabled {
		if user.Email == "" || user.VerificationCode == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "管理员开启了邮箱验证，请输入邮箱地址和验证码",
			})
			return
		}
		if !common.VerifyCodeWithKey(user.Email, user.VerificationCode, common.EmailVerificationPurpose) {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "验证码错误或已过期",
			})
			return
		}
	}
	identifier := user.Username
	if identifier == "" {
		identifier = user.Email
	}
	if identifier == "" {
		identifier = user.Phone
	}
	if identifier == "" {
		identifier = user.QQ
	}
	if identifier == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_parameter"),
		})
		return
	}

	// Auto-detect identifier type
	isEmail := strings.Contains(identifier, "@")
	isPhone := false
	isQQ := false
	if !isEmail {
		// Pure digits only
		digitsOnly := true
		for _, ch := range identifier {
			if ch < '0' || ch > '9' {
				digitsOnly = false
				break
			}
		}
		if digitsOnly {
			if len(identifier) == 11 && identifier[0] == '1' {
				isPhone = true
			} else if len(identifier) >= 5 && len(identifier) <= 12 {
				isQQ = true
			}
		}
	}

	// Check uniqueness across all identity fields
	var existingUser model.User
	dupField := ""
	if model.DB.Where("username = ?", identifier).First(&existingUser).Error == nil {
		dupField = "username"
	} else if isEmail && model.DB.Where("email = ?", identifier).First(&existingUser).Error == nil {
		dupField = "email"
	} else if isPhone && model.DB.Where("phone = ?", identifier).First(&existingUser).Error == nil {
		dupField = "phone"
	} else if isQQ && model.DB.Where("qq = ?", identifier).First(&existingUser).Error == nil {
		dupField = "qq"
	}
	if dupField != "" {
		var msg string
		switch dupField {
		case "email":
			msg = i18n.Translate(c, "email_exists")
		case "phone":
			msg = i18n.Translate(c, "phone_exists")
		case "qq":
			msg = i18n.Translate(c, "qq_exists")
		default:
			msg = i18n.Translate(c, "username_exists")
		}
		c.JSON(http.StatusOK, gin.H{"success": false, "message": msg})
		return
	}

	affCode := user.AffCode
	inviterId, _ := model.GetUserIdByAffCode(affCode)
	cleanUser := model.User{
		Username:    identifier,
		Password:    user.Password,
		DisplayName: user.Username,
		InviterId:   inviterId,
		Role:        user.Role,
	}
	if isEmail {
		cleanUser.Email = identifier
	}
	if isPhone {
		cleanUser.Phone = identifier
	}
	if isQQ {
		cleanUser.QQ = identifier
	}
	if config.EmailVerificationEnabled && user.Email != "" {
		cleanUser.Email = user.Email
	}
	if user.Phone != "" && user.Phone != identifier {
		cleanUser.Phone = user.Phone
	}
	if user.QQ != "" && user.QQ != identifier {
		cleanUser.QQ = user.QQ
	}

	// Enforce role: only valid user types
	if cleanUser.Role != model.RoleCommonUser && cleanUser.Role != model.RoleSupplier {
		cleanUser.Role = model.RoleCommonUser
	}

	if err := cleanUser.Insert(ctx, inviterId); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "registration_failed"),
		})
		return
	}

	// 自动发放邀请注册奖励 + 绑定推广关系
	if inviterId > 0 {
		setting, _ := model.GetCommissionSetting()
		if setting.Enabled && setting.RegisterReward > 0 {
			model.CreateCommissionRecord(&model.CommissionRecord{
				UserId:      inviterId,
				FromUserId:  cleanUser.Id,
				Type:        "register",
				Amount:      setting.RegisterReward,
				Status:      "settled",
				Description: "好友注册奖励",
			})
			// 直接增加邀请人余额
			model.DB.Model(&model.User{}).Where("id = ?", inviterId).
				UpdateColumn("quota", gorm.Expr("quota + ?", setting.RegisterReward))
		}

		// 创建 affiliate_relation（推广关系绑定）
		relation := model.AffiliateRelation{
			PromoterId:  inviterId,
			ConsumerId:  cleanUser.Id,
			CreatedTime: time.Now().Unix(),
		}
		// 检查是否已有推广关系，避免重复绑定
		var existing int64
		model.DB.Model(&model.AffiliateRelation{}).Where("consumer_id = ?", cleanUser.Id).Count(&existing)
		if existing == 0 {
			if err := model.DB.Create(&relation).Error; err != nil {
				logger.SysError(fmt.Sprintf("failed to create affiliate_relation: %v", err))
			}
		}
	}

	// 注册成功后自动登录
	SetupLogin(&cleanUser, c)
}

// getAdminOrgFilter 获取管理员可见的组织范围
// Root (100) 可见全部；其他 admin (10+) 仅限本组织
func getAdminOrgFilter(c *gin.Context) int {
	role := c.GetInt(ctxkey.Role)
	if role >= model.RoleRootUser {
		return 0 // 不限组织
	}
	return c.GetInt(ctxkey.OrganizationID)
}

func GetAllUsers(c *gin.Context) {
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}

	order := c.DefaultQuery("order", "")
	orgId := getAdminOrgFilter(c)
	users, err := model.GetAllUsers(p*config.ItemsPerPage, config.ItemsPerPage, order, orgId)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    users,
	})
}

func SearchUsers(c *gin.Context) {
	keyword := c.Query("keyword")
	orgId := getAdminOrgFilter(c)
	users, err := model.SearchUsers(keyword, orgId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    users,
	})
	return
}

func GetUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	user, err := model.GetUserById(id, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	myRole := c.GetInt(ctxkey.Role)
	if myRole <= user.Role && myRole != model.RoleRootUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权获取同级或更高等级用户的信息",
		})
		return
	}

	// 非 Root admin 只能看本组织用户
	orgId := getAdminOrgFilter(c)
	if orgId > 0 && user.OrganizationID != orgId {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "该用户不在您管理的组织中",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    user,
	})
	return
}

// DailyStat — 每日请求统计
type DailyStat struct {
	Date         string `json:"date"`
	RequestCount int    `json:"request_count"`
	TokenCount   int    `json:"token_count"`
	QuotaUsed    int    `json:"quota_used"`
}

// ModelStat — 模型维度统计
type ModelStat struct {
	ModelName    string `json:"model_name"`
	RequestCount int    `json:"request_count"`
	TokenCount   int    `json:"token_count"`
	QuotaUsed    int    `json:"quota_used"`
}

// ProviderStat — 提供商维度统计
type ProviderStat struct {
	Provider     string `json:"provider"`
	RequestCount int    `json:"request_count"`
	TokenCount   int    `json:"token_count"`
	QuotaUsed    int    `json:"quota_used"`
}

func GetUserDashboard(c *gin.Context) {
	id := c.GetInt(ctxkey.Id)
	now := time.Now()
	startOfDay := now.Truncate(24*time.Hour).AddDate(0, 0, -6).Unix()
	endOfDay := now.Truncate(24 * time.Hour).Add(24*time.Hour - time.Second).Unix()

	dashboards, err := model.SearchLogsByDayAndModel(id, int(startOfDay), int(endOfDay))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无法获取统计信息",
			"data":    nil,
		})
		return
	}

	// 1. 按日聚合 — DailyRequests
	dayMap := make(map[string]*DailyStat)
	// 2. 按 model 聚合 — ModelBreakdown
	modelMap := make(map[string]*ModelStat)
	// 3. 按提供商聚合 — ProviderBreakdown（需要 model→provider 映射）
	// 通过已配置渠道构建 model→provider 映射
	modelProvider := buildModelProviderMap()
	providerMap := make(map[string]*ProviderStat)

	for _, d := range dashboards {
		tokens := d.PromptTokens + d.CompletionTokens

		// 每日统计
		if _, ok := dayMap[d.Day]; !ok {
			dayMap[d.Day] = &DailyStat{Date: d.Day}
		}
		dayMap[d.Day].RequestCount += d.RequestCount
		dayMap[d.Day].TokenCount += tokens
		dayMap[d.Day].QuotaUsed += d.Quota

		// 模型统计
		if _, ok := modelMap[d.ModelName]; !ok {
			modelMap[d.ModelName] = &ModelStat{ModelName: d.ModelName}
		}
		modelMap[d.ModelName].RequestCount += d.RequestCount
		modelMap[d.ModelName].TokenCount += tokens
		modelMap[d.ModelName].QuotaUsed += d.Quota

		// 提供商统计
		provider := modelProvider[d.ModelName]
		if provider == "" {
			provider = "其他"
		}
		if _, ok := providerMap[provider]; !ok {
			providerMap[provider] = &ProviderStat{Provider: provider}
		}
		providerMap[provider].RequestCount += d.RequestCount
		providerMap[provider].TokenCount += tokens
		providerMap[provider].QuotaUsed += d.Quota
	}

	// 转换为有序切片
	var dailyRequests []DailyStat
	for _, v := range dayMap {
		dailyRequests = append(dailyRequests, *v)
	}
	// 按日期排序
	for i := 0; i < len(dailyRequests); i++ {
		for j := i + 1; j < len(dailyRequests); j++ {
			if dailyRequests[i].Date > dailyRequests[j].Date {
				dailyRequests[i], dailyRequests[j] = dailyRequests[j], dailyRequests[i]
			}
		}
	}

	var modelBreakdown []ModelStat
	for _, v := range modelMap {
		modelBreakdown = append(modelBreakdown, *v)
	}

	var providerBreakdown []ProviderStat
	for _, v := range providerMap {
		providerBreakdown = append(providerBreakdown, *v)
	}

	if dailyRequests == nil {
		dailyRequests = []DailyStat{}
	}
	if modelBreakdown == nil {
		modelBreakdown = []ModelStat{}
	}
	if providerBreakdown == nil {
		providerBreakdown = []ProviderStat{}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"logs":               dashboards,
			"daily_requests":     dailyRequests,
			"model_breakdown":    modelBreakdown,
			"provider_breakdown": providerBreakdown,
		},
	})
	return
}

// buildModelProviderMap — 从已配置渠道构建 model_name → provider 名称的映射
func buildModelProviderMap() map[string]string {
	result := make(map[string]string)
	allCh, _ := model.GetAllChannels(0, 0, "all")
	channelTypeNames := channeltype.ChannelTypeNames

	for _, ch := range allCh {
		if ch.Key == "" || strings.HasPrefix(ch.Key, "PUT_YOUR") {
			continue
		}
		provider := ""
		if name, ok := channelTypeNames[ch.Type]; ok {
			provider = name
		}
		if provider == "" {
			continue
		}

		var modelNames []string
		if ch.Models == "" {
			if modelList, ok := channelId2Models[ch.Type]; ok {
				modelNames = modelList
			}
		} else {
			for _, m := range strings.Split(ch.Models, ",") {
				m = strings.TrimSpace(m)
				if m != "" {
					modelNames = append(modelNames, m)
				}
			}
		}
		for _, m := range modelNames {
			if _, exists := result[m]; !exists {
				result[m] = provider
			}
		}
	}
	return result
}

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

func GetSelf(c *gin.Context) {
	id := c.GetInt(ctxkey.Id)
	user, err := model.GetUserById(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    user,
	})
	return
}

func UpdateUser(c *gin.Context) {
	ctx := c.Request.Context()
	var updatedUser model.User
	err := json.NewDecoder(c.Request.Body).Decode(&updatedUser)
	if err != nil || updatedUser.Id == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_parameter"),
		})
		return
	}
	if updatedUser.Password == "" {
		updatedUser.Password = "$I_LOVE_U" // make Validator happy :)
	}
	if err := common.Validate.Struct(&updatedUser); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_input"),
		})
		return
	}
	originUser, err := model.GetUserById(updatedUser.Id, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	myRole := c.GetInt(ctxkey.Role)
	if myRole <= originUser.Role && myRole != model.RoleRootUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权更新同权限等级或更高权限等级的用户信息",
		})
		return
	}
	if myRole <= updatedUser.Role && myRole != model.RoleRootUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权将其他用户权限等级提升到大于等于自己的权限等级",
		})
		return
	}
	if updatedUser.Password == "$I_LOVE_U" {
		updatedUser.Password = "" // rollback to what it should be
	}
	updatePassword := updatedUser.Password != ""
	if err := updatedUser.Update(updatePassword); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if originUser.Quota != updatedUser.Quota {
		model.RecordLog(ctx, originUser.Id, model.LogTypeManage, fmt.Sprintf("管理员将用户额度从 %s修改为 %s", common.LogQuota(originUser.Quota), common.LogQuota(updatedUser.Quota)))
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

func UpdateSelf(c *gin.Context) {
	var user model.User
	err := json.NewDecoder(c.Request.Body).Decode(&user)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_parameter"),
		})
		return
	}
	if user.Password == "" {
		user.Password = "$I_LOVE_U" // make Validator happy :)
	}
	if err := common.Validate.Struct(&user); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "输入不合法 " + err.Error(),
		})
		return
	}

	cleanUser := model.User{
		Id:          c.GetInt(ctxkey.Id),
		Username:    user.Username,
		Password:    user.Password,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		AvatarURL:   user.AvatarURL,
	}
	if user.Password == "$I_LOVE_U" {
		user.Password = "" // rollback to what it should be
		cleanUser.Password = ""
	}
	updatePassword := user.Password != ""
	if err := cleanUser.Update(updatePassword); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

func DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	originUser, err := model.GetUserById(id, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	myRole := c.GetInt("role")
	if myRole <= originUser.Role {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权删除同权限等级或更高权限等级的用户",
		})
		return
	}

	// 非 Root admin 只能删本组织用户
	orgId := getAdminOrgFilter(c)
	if orgId > 0 && originUser.OrganizationID != orgId {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "该用户不在您管理的组织中",
		})
		return
	}
	err = model.DeleteUserById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "",
		})
		return
	}
}

func DeleteSelf(c *gin.Context) {
	id := c.GetInt("id")
	user, _ := model.GetUserById(id, false)

	if user.Role == model.RoleRootUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "不能删除超级管理员账户",
		})
		return
	}

	err := model.DeleteUserById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

func CreateUser(c *gin.Context) {
	ctx := c.Request.Context()
	var user model.User
	err := json.NewDecoder(c.Request.Body).Decode(&user)
	if err != nil || user.Username == "" || user.Password == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_parameter"),
		})
		return
	}
	if err := common.Validate.Struct(&user); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_input"),
		})
		return
	}
	if user.DisplayName == "" {
		user.DisplayName = user.Username
	}
	myRole := c.GetInt("role")
	if user.Role >= myRole {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无法创建权限大于等于自己的用户",
		})
		return
	}
	// Even for admin users, we cannot fully trust them!
	// 非 Root admin 创建的用户自动归入本组织
	userOrgId := 0
	orgId := getAdminOrgFilter(c)
	if orgId > 0 {
		userOrgId = orgId
	}

	cleanUser := model.User{
		Username:         user.Username,
		Password:         user.Password,
		DisplayName:      user.DisplayName,
		OrganizationID:   userOrgId,
	}
	if err := cleanUser.Insert(ctx, 0); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

type ManageRequest struct {
	Username string `json:"username"`
	Action   string `json:"action"`
}

// ManageUser Only admin user can do this
func ManageUser(c *gin.Context) {
	var req ManageRequest
	err := json.NewDecoder(c.Request.Body).Decode(&req)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_parameter"),
		})
		return
	}
	user := model.User{
		Username: req.Username,
	}
	// Fill attributes
	model.DB.Where(&user).First(&user)
	if user.Id == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户不存在",
		})
		return
	}
	myRole := c.GetInt("role")
	if myRole <= user.Role && myRole != model.RoleRootUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权更新同权限等级或更高权限等级的用户信息",
		})
		return
	}

	// 非 Root admin 只能管理本组织用户
	orgId := getAdminOrgFilter(c)
	if orgId > 0 && user.OrganizationID != orgId {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "该用户不在您管理的组织中",
		})
		return
	}
	switch req.Action {
	case "disable":
		if user.Role == model.RoleRootUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法禁用超级管理员用户",
			})
			return
		}
		user.Status = model.UserStatusDisabled
		// Cascade: 禁用所有活跃 Token + 加入黑名单
		if err := model.DB.Model(&model.Token{}).
			Where("user_id = ? AND status = ? AND status != ?", user.Id, model.TokenStatusEnabled, model.TokenStatusDisabled).
			Update("status", model.TokenStatusDisabled).Error; err != nil {
			logger.SysErrorf("failed to disable tokens for user %d: %v", user.Id, err)
		}
		blacklist.BanUser(user.Id)
	case "enable":
		user.Status = model.UserStatusEnabled
		// Re-enable tokens + remove from blacklist
		if err := model.DB.Model(&model.Token{}).
			Where("user_id = ? AND status = ?", user.Id, model.TokenStatusDisabled).
			Update("status", model.TokenStatusEnabled).Error; err != nil {
			logger.SysErrorf("failed to re-enable tokens for user %d: %v", user.Id, err)
		}
		blacklist.UnbanUser(user.Id)
	case "delete":
		if user.Role == model.RoleRootUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法删除超级管理员用户",
			})
			return
		}
		if err := user.Delete(); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	case "promote":
		if myRole != model.RoleRootUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "普通管理员用户无法提升其他用户为管理员",
			})
			return
		}
		if user.Role >= model.RoleAdminUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "该用户已经是管理员",
			})
			return
		}
		user.Role = model.RoleAdminUser
	case "demote":
		if user.Role == model.RoleRootUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法降级超级管理员用户",
			})
			return
		}
		if user.Role == model.RoleCommonUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "该用户已经是普通用户",
			})
			return
		}
		user.Role = model.RoleCommonUser
	}

	if err := user.Update(false); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	clearUser := model.User{
		Role:   user.Role,
		Status: user.Status,
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    clearUser,
	})
	return
}

func EmailBind(c *gin.Context) {
	email := c.Query("email")
	code := c.Query("code")
	if !common.VerifyCodeWithKey(email, code, common.EmailVerificationPurpose) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "验证码错误或已过期",
		})
		return
	}
	id := c.GetInt("id")
	user := model.User{
		Id: id,
	}
	err := user.FillUserById()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	user.Email = email
	// no need to check if this email already taken, because we have used verification code to check it
	err = user.Update(false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if user.Role == model.RoleRootUser {
		config.RootUserEmail = email
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

type topUpRequest struct {
	Key string `json:"key"`
}

func TopUp(c *gin.Context) {
	ctx := c.Request.Context()
	req := topUpRequest{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	id := c.GetInt("id")
	quota, err := model.Redeem(ctx, req.Key, id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    quota,
	})
	return
}

type adminTopUpRequest struct {
	UserId int    `json:"user_id"`
	Quota  int    `json:"quota"`
	Remark string `json:"remark"`
}

func AdminTopUp(c *gin.Context) {
	ctx := c.Request.Context()
	req := adminTopUpRequest{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	err = model.IncreaseUserQuota(req.UserId, int64(req.Quota))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if req.Remark == "" {
		req.Remark = fmt.Sprintf("通过 API 充值 %s", common.LogQuota(int64(req.Quota)))
	}
	model.RecordTopupLog(ctx, req.UserId, req.Remark, req.Quota)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

// UpgradeToProvider 升级为渠道商
// 用户配置好 API 渠道后即可直接升级，无需管理员审批
// 提现时才需要身份信息审核（见 withdrawal.go）
func UpgradeToProvider(c *gin.Context) {
	userId := c.GetInt("id")

	// 检查当前用户类型
	user, err := model.GetUserById(userId, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "用户不存在"})
		return
	}
	if user.UserType == model.UserTypeProvider {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "已是渠道商，无需重复升级"})
		return
	}

	// 立即升级为渠道商（无需审核，不要求已有渠道）
	if err := model.DB.Model(&model.User{}).Where("id = ?", userId).
		Updates(map[string]interface{}{
			"user_type": model.UserTypeProvider,
			"role":      model.RoleSupplier,
		}).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "升级失败"})
		return
	}

	model.RecordLog(c.Request.Context(), userId, model.LogTypeSystem, "用户升级为渠道商")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "升级成功，您现在可以管理 API 渠道了",
	})
}

// GetMyTeam 获取我的团队成员
func GetMyTeam(c *gin.Context) {
	userId := c.GetInt("id")

	members, err := model.GetUsersByInviter(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "获取团队成员失败"})
		return
	}

	type TeamMember struct {
		Id          int    `json:"id"`
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		Role        int    `json:"role"`
		Status      int    `json:"status"`
		Quota       int64  `json:"quota"`
		UsedQuota   int64  `json:"used_quota"`
		RequestCount int   `json:"request_count"`
	}

	var result []TeamMember
	for _, m := range members {
		result = append(result, TeamMember{
			Id:           m.Id,
			Username:     m.Username,
			DisplayName:  m.DisplayName,
			Role:         m.Role,
			Status:       m.Status,
			Quota:        m.Quota,
			UsedQuota:    m.UsedQuota,
			RequestCount: m.RequestCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    result,
	})
}
