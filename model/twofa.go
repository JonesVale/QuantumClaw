package model

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

// TwoFA 用户两步验证配置
type TwoFA struct {
	Id     int    `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId int    `json:"user_id" gorm:"uniqueIndex;not null"`
	Secret string `json:"secret" gorm:"type:text"` // TOTP 密钥（加密存储更好）

	// 备用码（加密或哈希存储）
	BackupCodes string `json:"backup_codes" gorm:"type:text"` // JSON 数组，哈希存储

	Enabled   bool  `json:"enabled" gorm:"default:false"`
	CreatedAt int64 `json:"created_at" gorm:"bigint"`
	UpdatedAt int64 `json:"updated_at" gorm:"bigint"`
}

func (t *TwoFA) BeforeCreate(tx *gorm.DB) error {
	now := time.Now().Unix()
	t.CreatedAt = now
	t.UpdatedAt = now
	return nil
}

func (t *TwoFA) BeforeUpdate(tx *gorm.DB) error {
	t.UpdatedAt = time.Now().Unix()
	return nil
}

func (TwoFA) TableName() string {
	return "two_fas"
}

// ==================== 临时密钥存储（内存中，验证前有效）====================

var twoFATempSecrets = sync.Map{} // userId -> secret

func SetTwoFATempSecret(userId int, secret string) {
	twoFATempSecrets.Store(userId, secret)
}

func GetTwoFATempSecret(userId int) string {
	val, ok := twoFATempSecrets.Load(userId)
	if !ok {
		return ""
	}
	return val.(string)
}

func ClearTwoFATempSecret(userId int) {
	twoFATempSecrets.Delete(userId)
}

// ==================== 2FA 数据库操作 ====================

var ErrTwoFANotFound = errors.New("twofa not found")

// GetTwoFAByUserId 获取用户的 2FA 配置
func GetTwoFAByUserId(userId int) (*TwoFA, error) {
	if userId <= 0 {
		return nil, errors.New("invalid userId")
	}
	var twoFA TwoFA
	err := DB.Where("user_id = ? AND enabled = ?", userId, true).First(&twoFA).Error
	if err != nil {
		return nil, ErrTwoFANotFound
	}
	return &twoFA, nil
}

// EnableTwoFA 启用用户的 2FA
func EnableTwoFA(userId int, secret string, backupCodes []string) error {
	if userId <= 0 {
		return errors.New("invalid userId")
	}
	// 哈希备用码
	hashedCodes := make([]string, len(backupCodes))
	for i, code := range backupCodes {
		hashedCodes[i] = hashBackupCode(code)
	}

	codeJSON, _ := json.Marshal(hashedCodes)
	twoFA := &TwoFA{
		UserId:       userId,
		Secret:       secret,
		BackupCodes:  string(codeJSON),
		Enabled:      true,
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}

	// 查找或创建
	var existing TwoFA
	err := DB.Where("user_id = ?", userId).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return DB.Create(twoFA).Error
	}
	if err != nil {
		return err
	}
	twoFA.Id = existing.Id
	return DB.Save(twoFA).Error
}

// DisableTwoFA 禁用用户的 2FA
func DisableTwoFA(userId int) error {
	if userId <= 0 {
		return errors.New("invalid userId")
	}
	return DB.Model(&TwoFA{}).Where("user_id = ?", userId).
		Updates(map[string]interface{}{
			"enabled":   false,
			"updated_at": time.Now().Unix(),
		}).Error
}

// ValidateTwoFABackupCode 验证备用码（使用哈希比对）
func ValidateTwoFABackupCode(userId int, code string) bool {
	twoFA, err := GetTwoFAByUserId(userId)
	if err != nil {
		return false
	}
	if twoFA.BackupCodes == "" {
		return false
	}
	code = strings.ToUpper(strings.TrimSpace(code))
	hashed := hashBackupCode(code)
	var codes []string
	if err := json.Unmarshal([]byte(twoFA.BackupCodes), &codes); err != nil {
		return false
	}
	for _, c := range codes {
		if c == hashed {
			return true
		}
	}
	return false
}

// ConsumeTwoFABackupCode 使用备用码（标记为已用）
func ConsumeTwoFABackupCode(userId int, code string) {
	twoFA, err := GetTwoFAByUserId(userId)
	if err != nil {
		return
	}
	code = strings.ToUpper(strings.TrimSpace(code))
	hashed := hashBackupCode(code)
	var codes []string
	if err := json.Unmarshal([]byte(twoFA.BackupCodes), &codes); err != nil {
		return
	}
	// 移除已使用的码
	newCodes := make([]string, 0)
	for _, c := range codes {
		if c != hashed {
			newCodes = append(newCodes, c)
		}
	}
	codeJSON, _ := json.Marshal(newCodes)
	DB.Model(twoFA).Update("backup_codes", string(codeJSON))
}

// CountRemainingBackupCodes 统计剩余备用码数量
func CountRemainingBackupCodes(userId int) int {
	twoFA, err := GetTwoFAByUserId(userId)
	if err != nil {
		return 0
	}
	if twoFA.BackupCodes == "" {
		return 0
	}
	var codes []string
	if err := json.Unmarshal([]byte(twoFA.BackupCodes), &codes); err != nil {
		return 0
	}
	return len(codes)
}

// hashBackupCode 哈希备用码
func hashBackupCode(code string) string {
	// 使用 SHA256（生产环境建议使用 bcrypt）
	h := sha256.Sum256([]byte(strings.ToUpper(code)))
	return hex.EncodeToString(h[:])
}
