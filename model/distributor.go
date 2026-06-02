package model

import (
	"fmt"
	"time"

	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/encrypt"
	"gorm.io/gorm"
)

// Distributor — 分销商
type Distributor struct {
	Id              int       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId          int       `json:"user_id" gorm:"uniqueIndex;not null"`       // 关联的用户
	Name            string    `json:"name" gorm:"type:varchar(100)"`             // 分销商名称
	ContactEmail    string    `json:"contact_email" gorm:"type:varchar(255)"`
	MarkupRate      float64   `json:"markup_rate" gorm:"type:decimal(5,4);default:0"` // 加价率（如 0.2 = +20%）
	ProfitSplit     float64   `json:"profit_split" gorm:"type:decimal(5,4);default:0.5"` // 平台分润比例（0.5 = 50%）
	Status          int       `json:"status" gorm:"default:1"`                   // 1=启用 0=禁用
	ApiKey          string    `json:"api_key" gorm:"type:varchar(64);uniqueIndex"` // 分销商 API 密钥
	TotalRevenue    int64     `json:"total_revenue" gorm:"default:0"`            // 累计营收
	TotalPayout     int64     `json:"total_payout" gorm:"default:0"`             // 累计分成支出
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// DistributorPricingRule — 分销商自定义定价规则
type DistributorPricingRule struct {
	Id             int     `json:"id" gorm:"primaryKey;autoIncrement"`
	DistributorId  int     `json:"distributor_id" gorm:"index;not null"`
	ModelName      string  `json:"model_name" gorm:"type:varchar(100);not null"` // * = 所有模型
	PriceMultiplier float64 `json:"price_multiplier" gorm:"type:decimal(6,4);default:1.0"` // 价格倍率
	FixedPrice     int64   `json:"fixed_price" gorm:"default:0"`                 // 固定价格（0=不启用）
}

// BeforeCreate GORM hook: encrypt ApiKey before storing
func (d *Distributor) BeforeCreate(tx *gorm.DB) error {
	if d.ApiKey != "" && config.CryptoSecret != "" {
		key := encrypt.DeriveKey(config.CryptoSecret)
		encrypted, err := encrypt.Encrypt([]byte(d.ApiKey), key)
		if err != nil {
			return fmt.Errorf("encrypt distributor api_key: %w", err)
		}
		d.ApiKey = encrypted
	}
	return nil
}

// AfterFind GORM hook: decrypt ApiKey after reading
func (d *Distributor) AfterFind(tx *gorm.DB) error {
	if d.ApiKey != "" && config.CryptoSecret != "" {
		key := encrypt.DeriveKey(config.CryptoSecret)
		decrypted, err := encrypt.Decrypt(d.ApiKey, key)
		if err == nil {
			d.ApiKey = string(decrypted)
		} else {
			// Not encrypted yet (plaintext), leave as-is
			// This handles backward compatibility with existing plaintext keys
		}
	}
	return nil
}

// BeforeUpdate GORM hook: encrypt ApiKey before updating
func (d *Distributor) BeforeUpdate(tx *gorm.DB) error {
	if d.ApiKey != "" && config.CryptoSecret != "" {
		key := encrypt.DeriveKey(config.CryptoSecret)
		encrypted, err := encrypt.Encrypt([]byte(d.ApiKey), key)
		if err != nil {
			return fmt.Errorf("encrypt distributor api_key on update: %w", err)
		}
		d.ApiKey = encrypted
	}
	return nil
}

func InitDistributorTables() {
	DB.AutoMigrate(&Distributor{}, &DistributorPricingRule{})
}

// GetDistributorByUserId 通过用户 ID 获取分销商
func GetDistributorByUserId(userId int) (*Distributor, error) {
	var d Distributor
	err := DB.Where("user_id = ?", userId).First(&d).Error
	return &d, err
}

// GetAllDistributors 获取所有分销商
func GetAllDistributors() ([]Distributor, error) {
	var ds []Distributor
	err := DB.Order("id desc").Find(&ds).Error
	return ds, err
}

// GetDistributorPricingRules 获取分销商定价规则
func GetDistributorPricingRules(distributorId int) ([]DistributorPricingRule, error) {
	var rules []DistributorPricingRule
	err := DB.Where("distributor_id = ?", distributorId).Find(&rules).Error
	return rules, err
}

// CalculateDistributorPrice 计算分销商价格（基础价 × 加价率 × 规则倍率）
func CalculateDistributorPrice(distributorId int, modelName string, basePrice int64) int64 {
	var d Distributor
	if err := DB.First(&d, distributorId).Error; err != nil || d.Status == 0 {
		return basePrice
	}
	// 基础加价
	price := float64(basePrice) * (1 + d.MarkupRate)
	// 查找匹配的定价规则（精确模型 > 通配符）
	var rule DistributorPricingRule
	found := false
	if DB.Where("distributor_id = ? AND model_name = ?", distributorId, modelName).First(&rule).Error == nil {
		found = true
	}
	if !found {
		DB.Where("distributor_id = ? AND model_name = '*'", distributorId).First(&rule)
	}
	if found && rule.FixedPrice > 0 {
		return rule.FixedPrice
	}
	if found {
		price = price * rule.PriceMultiplier
	}
	return int64(price)
}

// RecordDistributorRevenue 记录分销商营收
func RecordDistributorRevenue(distributorId int, amount int64) {
	DB.Model(&Distributor{}).Where("id = ?", distributorId).
		Updates(map[string]interface{}{
			"total_revenue": gorm.Expr("total_revenue + ?", amount),
		})
}
