package model

import (
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"gorm.io/gorm"
)

// WebAuthnCredential 存储用户的 WebAuthn 凭证（Passkey）
type WebAuthnCredential struct {
	ID           uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       int    `gorm:"index;not null" json:"user_id"`
	CredentialID string `gorm:"type:text;not null" json:"credential_id"` // base64 encoded
	PublicKey    []byte `gorm:"type:blob" json:"-"`                     // COSE key bytes
	Attestation  string `gorm:"type:varchar(64)" json:"attestation_type"`
	AAGUID       string `gorm:"type:varchar(36)" json:"aaguid"`
	SignCount    uint32 `gorm:"default:0" json:"sign_count"`
	CloneWarning bool   `gorm:"default:false" json:"clone_warning"`
	DeviceName   string `gorm:"type:varchar(128)" json:"device_name"`
	CreatedTime  int64  `gorm:"autoCreateTime" json:"created_time"`
	UpdatedTime  int64  `gorm:"autoUpdateTime" json:"updated_time"`

	// 存储完整的 webauthn.Credential 对象（序列化）
	CredentialBlob []byte `gorm:"type:blob" json:"-"`
}

func (WebAuthnCredential) TableName() string {
	return "webauthn_credentials"
}

// ToWebAuthnCredential 将数据库记录转换为 webauthn.Credential
func (c *WebAuthnCredential) ToWebAuthnCredential() (*webauthn.Credential, error) {
	if len(c.CredentialBlob) > 0 {
		var cred webauthn.Credential
		if err := json.Unmarshal(c.CredentialBlob, &cred); err == nil {
			return &cred, nil
		}
	}
	// 如果没有 blob，尝试从单独字段构建
	return nil, errors.New("credential blob is empty or invalid")
}

// FromWebAuthnCredential 从 webauthn.Credential 填充数据库记录
func (c *WebAuthnCredential) FromWebAuthnCredential(cred *webauthn.Credential) error {
	c.CredentialID = base64.RawURLEncoding.EncodeToString(cred.ID)
	c.PublicKey = cred.PublicKey
	c.SignCount = cred.Authenticator.SignCount
	c.CloneWarning = cred.Authenticator.CloneWarning
	// AAGUID is []byte, store as hex string
	if len(cred.Authenticator.AAGUID) > 0 {
		c.AAGUID = fmt.Sprintf("%x", cred.Authenticator.AAGUID)
	}

	blob, err := json.Marshal(cred)
	if err != nil {
		return err
	}
	c.CredentialBlob = blob
	return nil
}

// WebAuthnCredentialJSON 用于 GORM 存储 JSON 字段
type WebAuthnCredentialJSON webauthn.Credential

func (w WebAuthnCredentialJSON) Value() (driver.Value, error) {
	return json.Marshal(w)
}

func (w *WebAuthnCredentialJSON) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &w)
}

// GetUserCredentials 获取用户的所有凭证
func GetUserCredentials(userID int) ([]*WebAuthnCredential, error) {
	var credentials []*WebAuthnCredential
	err := DB.Where("user_id = ?", userID).Find(&credentials).Error
	return credentials, err
}

// GetWebAuthnCredentialByID 根据 CredentialID 获取凭证
func GetWebAuthnCredentialByID(credentialID string) (*WebAuthnCredential, error) {
	var cred WebAuthnCredential
	err := DB.Where("credential_id = ?", credentialID).First(&cred).Error
	return &cred, err
}

// UpdateCredentialSignCount 更新签名计数
func UpdateCredentialSignCount(credentialID string, signCount uint32, cloneWarning bool) error {
	return DB.Model(&WebAuthnCredential{}).
		Where("credential_id = ?", credentialID).
		Updates(map[string]interface{}{
			"sign_count":     signCount,
			"clone_warning":  cloneWarning,
			"updated_time":   time.Now().Unix(),
		}).Error
}

// DeleteWebAuthnCredential 删除用户的某个凭证
func DeleteWebAuthnCredential(userID int, credentialID string) error {
	return DB.Where("user_id = ? AND credential_id = ?", userID, credentialID).
		Delete(&WebAuthnCredential{}).Error
}

// HasWebAuthnCredential 检查用户是否有 Passkey
func HasWebAuthnCredential(userID int) (bool, error) {
	var count int64
	err := DB.Model(&WebAuthnCredential{}).Where("user_id = ?", userID).Count(&count).Error
	return count > 0, err
}

// AfterFind GORM hook：自动反序列化 CredentialBlob
func (c *WebAuthnCredential) AfterFind(tx *gorm.DB) (err error) {
	// CredentialBlob 已经通过 ToWebAuthnCredential 使用，这里不需要额外处理
	return nil
}
