package model

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

// ==================== 自定义 OAuth 提供商 ====================

// CustomOAuthProvider 自定义 OAuth 提供商配置
type CustomOAuthProvider struct {
	Id          int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string `json:"name" gorm:"type:varchar(64);uniqueIndex;not null"`
	DisplayName string `json:"display_name" gorm:"type:varchar(128);not null"`
	Enabled     bool   `json:"enabled" gorm:"default:false"`

	ClientId     string `json:"client_id" gorm:"type:varchar(256)"`
	ClientSecret string `json:"-" gorm:"-"` // 不序列化到 JSON

	AuthURL     string `json:"auth_url" gorm:"type:text"`
	TokenURL    string `json:"token_url" gorm:"type:text"`
	UserInfoURL string `json:"user_info_url" gorm:"type:text"`
	Scopes      string `json:"scopes" gorm:"type:varchar(256)"`

	UserIdField   string `json:"user_id_field" gorm:"type:varchar(64);default:'id'"`
	UsernameField string `json:"username_field" gorm:"type:varchar(64);default:'name'"`
	EmailField    string `json:"email_field" gorm:"type:varchar(64);default:'email'"`

	LogoURL     string `json:"logo_url" gorm:"type:text"`
	ButtonColor string `json:"button_color" gorm:"type:varchar(16);default:'#4285F4'"`
	SortOrder   int    `json:"sort_order" gorm:"type:int;default:0"`

	CreatedAt int64 `json:"created_at" gorm:"bigint"`
	UpdatedAt int64 `json:"updated_at" gorm:"bigint"`
}

func (p *CustomOAuthProvider) BeforeCreate(tx *gorm.DB) error {
	now := time.Now().Unix()
	p.CreatedAt = now
	p.UpdatedAt = now
	return nil
}

func (p *CustomOAuthProvider) BeforeUpdate(tx *gorm.DB) error {
	p.UpdatedAt = time.Now().Unix()
	return nil
}

func (p *CustomOAuthProvider) TableName() string {
	return "custom_oauth_providers"
}

// GetAllEnabledOAuthProviders 获取所有启用的 OAuth 提供商
func GetAllEnabledOAuthProviders() ([]CustomOAuthProvider, error) {
	var providers []CustomOAuthProvider
	err := DB.Where("enabled = ?", true).Order("sort_order DESC, id ASC").Find(&providers).Error
	return providers, err
}

// GetOAuthProviderByName 根据名称获取 OAuth 提供商
func GetOAuthProviderByName(name string) (*CustomOAuthProvider, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("provider name is empty")
	}
	var provider CustomOAuthProvider
	err := DB.Where("name = ?", name).First(&provider).Error
	return &provider, err
}

// GetCustomOAuthProviderById 根据 ID 获取 OAuth 提供商
func GetCustomOAuthProviderById(id int) (*CustomOAuthProvider, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}
	var provider CustomOAuthProvider
	err := DB.Where("id = ?", id).First(&provider).Error
	return &provider, err
}

// UpsertCustomOAuthProvider 创建或更新 OAuth 提供商
func UpsertCustomOAuthProvider(provider *CustomOAuthProvider) error {
	if provider.Name == "" {
		return fmt.Errorf("provider name is required")
	}
	provider.Name = strings.TrimSpace(provider.Name)
	provider.DisplayName = strings.TrimSpace(provider.DisplayName)

	var existing CustomOAuthProvider
	err := DB.Where("name = ?", provider.Name).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return DB.Create(provider).Error
	}
	if err != nil {
		return err
	}
	provider.Id = existing.Id
	return DB.Save(provider).Error
}

// GetAllCustomOAuthProviders 获取所有自定义 OAuth 提供商（含禁用的）
func GetAllCustomOAuthProviders() ([]CustomOAuthProvider, error) {
	var providers []CustomOAuthProvider
	err := DB.Order("sort_order DESC, id ASC").Find(&providers).Error
	return providers, err
}

// IsCustomOAuthIdAlreadyTaken 检查自定义 OAuth ID 是否已被占用
func IsCustomOAuthIdAlreadyTaken(id string) bool {
	return DB.Where("custom_oauth_id = ?", id).Find(&User{}).RowsAffected == 1
}

// FillUserByCustomOAuthId 根据 CustomOAuthId 填充用户信息
func (user *User) FillUserByCustomOAuthId() error {
	if user.CustomOAuthId == "" {
		return fmt.Errorf("custom_oauth_id is empty")
	}
	return DB.Where(User{CustomOAuthId: user.CustomOAuthId}).First(user).Error
}

// ==================== OAuth 状态存储（防 CSRF）====================

var oauthStateCache = struct {
	sync.RWMutex
	m map[string]*OAuthStateEntry
}{
	m: make(map[string]*OAuthStateEntry),
}

type OAuthStateEntry struct {
	Provider    string
	RedirectURL string
	CreatedAt   time.Time
}

const oauthStateExpire = 10 * time.Minute

// CreateOAuthState 创建一次性 OAuth 状态
func CreateOAuthState(provider string, redirectURL string) string {
	state := GenerateSecureToken(32)
	oauthStateCache.Lock()
	oauthStateCache.m[state] = &OAuthStateEntry{
		Provider:    provider,
		RedirectURL: redirectURL,
		CreatedAt:   time.Now(),
	}
	oauthStateCache.Unlock()
	return state
}

// ValidateOAuthState 验证并消耗 OAuth 状态（一次性）
func ValidateOAuthState(state string) (string, string, bool) {
	oauthStateCache.Lock()
	entry, ok := oauthStateCache.m[state]
	if ok {
		delete(oauthStateCache.m, state)
	}
	oauthStateCache.Unlock()

	if !ok || entry == nil {
		return "", "", false
	}
	if time.Since(entry.CreatedAt) > oauthStateExpire {
		return "", "", false
	}
	return entry.Provider, entry.RedirectURL, true
}

// GenerateSecureToken 生成伪随机 token（用于 CSRF，足够安全）
func GenerateSecureToken(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}

// ==================== OAuth 响应类型 ====================

type OAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	IDToken      string `json:"id_token"`
	Error        string `json:"error"`
	ErrorDesc    string `json:"error_description"`
}

type OAuthUserInfo struct {
	UserId   string                 `json:"user_id"`
	Username string                 `json:"username"`
	Email    string                 `json:"email"`
	Name     string                 `json:"name"`
	Avatar   string                 `json:"avatar"`
	Raw      map[string]interface{} `json:"raw"`
}

// ParseOAuthUserInfo 根据提供商配置解析用户信息
func ParseOAuthUserInfo(raw map[string]interface{}, provider *CustomOAuthProvider) *OAuthUserInfo {
	info := &OAuthUserInfo{Raw: raw}

	// 解析 ID
	if provider.UserIdField != "" {
		if val, ok := raw[provider.UserIdField]; ok {
			if s, ok := val.(string); ok {
				info.UserId = s
			} else if f, ok := val.(float64); ok {
				info.UserId = fmt.Sprintf("%.0f", f)
			}
		}
	}

	// 解析用户名
	fields := []string{provider.UsernameField, "login", "name", "display_name", "displayName", "username"}
	for _, k := range fields {
		if k == "" {
			continue
		}
		if val, ok := raw[k].(string); ok && val != "" {
			info.Username = val
			break
		}
	}

	// 解析邮箱
	emailFields := []string{provider.EmailField, "email", "mail"}
	for _, k := range emailFields {
		if k == "" {
			continue
		}
		if val, ok := raw[k].(string); ok && val != "" {
			info.Email = val
			break
		}
	}

	// 解析头像
	for _, k := range []string{"avatar", "avatar_url", "picture", "photo_url", "profile_picture"} {
		if val, ok := raw[k].(string); ok && val != "" {
			info.Avatar = val
			break
		}
	}

	return info
}
