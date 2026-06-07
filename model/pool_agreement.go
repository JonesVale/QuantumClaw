package model

import "github.com/quantumclaw/quantumclaw/common/helper"

// PlatformPoolAgreement 平台资源池协议版本
type PlatformPoolAgreement struct {
	ID          int    `json:"id" gorm:"primaryKey"`
	Version     int    `json:"version" gorm:"uniqueIndex;not null"`
	Title       string `json:"title" gorm:"type:varchar(255)"`
	Content     string `json:"content" gorm:"type:text;not null"`
	Required    bool   `json:"required" gorm:"default:true"`
	PublishedAt int64  `json:"published_at" gorm:"bigint"`
	CreatedAt   int64  `json:"created_at" gorm:"bigint"`
}

func (PlatformPoolAgreement) TableName() string { return "platform_pool_agreements" }

// UserPoolConsent 用户对平台池协议的同意记录
type UserPoolConsent struct {
	ID            int    `json:"id" gorm:"primaryKey"`
	UserID        int    `json:"user_id" gorm:"uniqueIndex;not null"`
	Agreed        bool   `json:"agreed" gorm:"default:false"`
	AgreedVersion int    `json:"agreed_version" gorm:"int;default:0"`
	AgreedAt      int64  `json:"agreed_at" gorm:"bigint"`
	UpdatedAt     int64  `json:"updated_at" gorm:"bigint"`
}

func (UserPoolConsent) TableName() string { return "user_pool_consents" }

// SeedDefaultAgreement 初始化默认协议
func SeedDefaultAgreement() error {
	var existing PlatformPoolAgreement
	if err := DB.First(&existing).Error; err == nil {
		return nil // 已有协议，不覆盖
	}
	return DB.Create(&PlatformPoolAgreement{
		Version:     1,
		Title:       "平台资源池使用协议",
		Content:     "一、定义\n平台资源池是指本平台所有渠道商（含平台自营）提供的同类资源的集合。\n\n二、同意条款\n当您所选的渠道商资源出现异常（包括但不限于服务不可用、响应超时、频繁报错）时，平台有权自动将您的请求切换到平台资源池中其他渠道商的同型号资源继续为您服务。\n\n三、切换规则\n1. 优先切换至价格最接近的可用资源。\n2. 切换过程对您透明，您无需重新配置。\n3. 平台记录切换日志以供后续查询。\n\n四、数据说明\n请求切至其他渠道商后，您的请求数据将由该渠道商处理，平台不干预数据处理过程。\n\n五、退出\n您可以随时在用户设置中关闭此授权。关闭后，平台将不再进行自动切换。",
		Required:    true,
		PublishedAt: helper.GetTimestamp(),
		CreatedAt:   helper.GetTimestamp(),
	}).Error
}

// GetLatestAgreement 获取最新版本的协议
func GetLatestAgreement() (*PlatformPoolAgreement, error) {
	var a PlatformPoolAgreement
	err := DB.Order("version DESC").First(&a).Error
	return &a, err
}

// GetAgreementByVersion 获取指定版本的协议
func GetAgreementByVersion(version int) (*PlatformPoolAgreement, error) {
	var a PlatformPoolAgreement
	err := DB.Where("version = ?", version).First(&a).Error
	return &a, err
}

// GetAllAgreements 获取所有协议版本列表
func GetAllAgreements() ([]PlatformPoolAgreement, error) {
	var list []PlatformPoolAgreement
	err := DB.Order("version DESC").Find(&list).Error
	return list, err
}

// PublishNewAgreement 发布新版本协议
func PublishNewAgreement(title, content string) (*PlatformPoolAgreement, error) {
	var latest PlatformPoolAgreement
	newVersion := 1
	if err := DB.Order("version DESC").First(&latest).Error; err == nil {
		newVersion = latest.Version + 1
	}
	a := &PlatformPoolAgreement{
		Version:     newVersion,
		Title:       title,
		Content:     content,
		Required:    true,
		PublishedAt: helper.GetTimestamp(),
		CreatedAt:   helper.GetTimestamp(),
	}
	err := DB.Create(a).Error
	return a, err
}

// GetUserPoolConsent 获取用户对平台池的同意记录
func GetUserPoolConsent(userID int) (*UserPoolConsent, error) {
	var c UserPoolConsent
	err := DB.Where("user_id = ?", userID).First(&c).Error
	return &c, err
}

// UpsertUserPoolConsent 创建或更新用户的同意记录
func UpsertUserPoolConsent(userID int, agreed bool) error {
	latest, err := GetLatestAgreement()
	if err != nil {
		return err
	}
	agreedVersion := 0
	if agreed {
		agreedVersion = latest.Version
	}
	now := helper.GetTimestamp()

	return DB.Model(&UserPoolConsent{}).
		Where("user_id = ?", userID).
		Assign(UserPoolConsent{
			Agreed:        agreed,
			AgreedVersion: agreedVersion,
			AgreedAt:      now,
			UpdatedAt:     now,
		}).
		FirstOrCreate(&UserPoolConsent{UserID: userID}).Error
}

// IsConsentValid 检查用户的同意是否仍有效（协议未更新）
func IsConsentValid(userID int) bool {
	consent, err := GetUserPoolConsent(userID)
	if err != nil {
		return false
	}
	if !consent.Agreed {
		return false
	}
	latest, err := GetLatestAgreement()
	if err != nil {
		return false
	}
	return consent.AgreedVersion >= latest.Version
}
