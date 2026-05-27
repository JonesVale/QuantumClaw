package model

import (
	"time"
)

// Sub2APIProvider defines which subscription service this credential belongs to.
type Sub2APIProvider string

const (
	Sub2ProviderChatGPTPlus  Sub2APIProvider = "chatgpt_plus"   // ChatGPT Plus ($20/mo)
	Sub2ProviderChatGPTPro   Sub2APIProvider = "chatgpt_pro"    // ChatGPT Pro ($200/mo)
	Sub2ProviderChatGPTTeam  Sub2APIProvider = "chatgpt_team"   // ChatGPT Team ($25/user/mo)
	Sub2ProviderClaudePro    Sub2APIProvider = "claude_pro"     // Claude Pro ($20/mo)
	Sub2ProviderClaudeTeam   Sub2APIProvider = "claude_team"    // Claude Team ($30/user/mo)
)

// Sub2APICredential stores a user's subscription credential.
// The Token field is AES-GCM encrypted at rest.
// Multiple credentials of the same provider can be added for redundancy.
type Sub2APICredential struct {
	Id            int              `json:"id" gorm:"primaryKey"`
	UserId        int              `json:"user_id" gorm:"not null;index"`
	Provider      Sub2APIProvider  `json:"provider" gorm:"type:varchar(50);not null;index"`
	Label         string           `json:"label" gorm:"type:varchar(128);default:''"`      // user-friendly name
	Token         string           `json:"-" gorm:"type:text;not null"`                     // encrypted session token
	TokenHash     string           `json:"-" gorm:"type:char(64);index"`                    // SHA-256 of plaintext for dedup
	Status        int              `json:"status" gorm:"default:1"`                         // 1=active, 2=paused, 3=invalid
	DailyCap      int64            `json:"daily_cap" gorm:"bigint;default:0"`               // 0=unlimited daily requests
	UsedToday     int64            `json:"used_today" gorm:"bigint;default:0"`               // requests used today (reset daily)
	LastHealthAt  int64            `json:"last_health_at" gorm:"bigint;default:0"`           // last successful health check
	LastError     string           `json:"last_error" gorm:"type:text"`                      // last error message
	ExpiresAt     int64            `json:"expires_at" gorm:"bigint;default:0"`               // 0 = unknown
	CreatedTime   int64            `json:"created_time" gorm:"bigint"`
	UpdatedTime   int64            `json:"updated_time" gorm:"bigint"`
}

func (Sub2APICredential) TableName() string {
	return "sub2api_credentials"
}

// Sub2APIUsage tracks daily usage per credential.
type Sub2APIUsage struct {
	Id           int   `json:"id" gorm:"primaryKey"`
	CredentialId int   `json:"credential_id" gorm:"not null;index"`
	UserId       int   `json:"user_id" gorm:"not null;index"`
	Date         string `json:"date" gorm:"type:varchar(10);not null"` // YYYY-MM-DD
	RequestCount int64 `json:"request_count" gorm:"bigint;default:0"`
	TokenCount   int64 `json:"token_count" gorm:"bigint;default:0"`
}

func (Sub2APIUsage) TableName() string {
	return "sub2api_usage"
}

// ── CRUD ──

func CreateSub2Credential(c *Sub2APICredential) error {
	now := time.Now().UnixMilli()
	c.CreatedTime = now
	c.UpdatedTime = now
	return DB.Create(c).Error
}

func GetUserSub2Credentials(userId int) ([]Sub2APICredential, error) {
	var creds []Sub2APICredential
	err := DB.Where("user_id = ?", userId).Order("created_time desc").Find(&creds).Error
	return creds, err
}

func GetSub2Credential(id, userId int) (*Sub2APICredential, error) {
	var c Sub2APICredential
	err := DB.Where("id = ? AND user_id = ?", id, userId).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func GetSub2CredentialById(id int) (*Sub2APICredential, error) {
	var c Sub2APICredential
	err := DB.Where("id = ?", id).First(&c).Error
	return &c, err
}

func UpdateSub2Credential(c *Sub2APICredential) error {
	c.UpdatedTime = time.Now().UnixMilli()
	return DB.Select("*").Omit("created_time").Updates(c).Error
}

func DeleteSub2Credential(id, userId int) error {
	return DB.Where("id = ? AND user_id = ?", id, userId).Delete(&Sub2APICredential{}).Error
}

// ListAllSub2Credentials returns all credentials (admin view, excludes token value).
func ListAllSub2Credentials() ([]Sub2APICredential, error) {
	var creds []Sub2APICredential
	err := DB.Select("id, user_id, provider, label, status, daily_cap, used_today, last_health_at, last_error, expires_at, created_time, updated_time").
		Order("created_time desc").Find(&creds).Error
	return creds, err
}

// GetActiveSub2Credentials returns active, healthy credentials for a provider.
func GetActiveSub2Credentials(provider Sub2APIProvider) ([]Sub2APICredential, error) {
	var creds []Sub2APICredential
	err := DB.Where("provider = ? AND status = 1", provider).
		Order("used_today asc, last_health_at desc").
		Find(&creds).Error
	return creds, err
}

// ListSub2CredentialsByProviderPrefix finds active credentials whose provider starts with the given prefix.
// e.g. prefix "chatgpt" matches "chatgpt_plus", "chatgpt_pro", "chatgpt_team"
func ListSub2CredentialsByProviderPrefix(userId int, prefix string) ([]Sub2APICredential, error) {
	var creds []Sub2APICredential
	err := DB.Where("user_id = ? AND status = 1 AND provider LIKE ?", userId, prefix+"%").
		Order("used_today asc, last_health_at desc").
		Find(&creds).Error
	return creds, err
}

// ResetDailyUsage resets UsedToday for all credentials (called daily by scheduler).
func ResetSub2DailyUsage() error {
	return DB.Model(&Sub2APICredential{}).Where("1 = 1").Update("used_today", 0).Error
}

// IncrementSub2Usage atomically increments used_today for a credential.
func IncrementSub2Usage(id int, tokens int64) error {
	return DB.Model(&Sub2APICredential{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"used_today":   DB.Raw("used_today + 1"),
			"updated_time": time.Now().UnixMilli(),
		}).Error
}
