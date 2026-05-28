package model

import (
	"strconv"
	"time"
)

// PlatformConfig 全局配置键值对
type PlatformConfig struct {
	Key         string `json:"key" gorm:"primaryKey;type:varchar(255)"`
	Value       string `json:"value" gorm:"type:text;not null"`
	UpdatedTime int64  `json:"updated_time" gorm:"bigint"`
}

func (PlatformConfig) TableName() string {
	return "platform_config"
}

// SettlementConfig 结算配置：每个模型一条
type SettlementConfig struct {
	Id              int     `json:"id" gorm:"primaryKey"`
	ModelName       string  `json:"model_name" gorm:"type:varchar(255);not null;uniqueIndex"`
	UnifiedCost     float64 `json:"unified_cost" gorm:"type:decimal(10,6);not null;default:0.001000"`
	CommissionRate  float64 `json:"commission_rate" gorm:"type:decimal(5,4);not null;default:0.2000"`
	PlatformFeeRate float64 `json:"platform_fee_rate" gorm:"type:decimal(5,4);not null;default:0.1000"`
	Enabled         int     `json:"enabled" gorm:"default:1"`
	CreatedTime     int64   `json:"created_time" gorm:"bigint"`
	UpdatedTime     int64   `json:"updated_time" gorm:"bigint"`
}

func (s *SettlementConfig) TableName() string {
	return "settlement_config"
}

// GetSettlementConfig 按 model_name 查询结算配置（精确匹配 → * 默认）
func GetSettlementConfig(modelName string) (*SettlementConfig, error) {
	var cfg SettlementConfig
	err := DB.Where("model_name = ? AND enabled = 1", modelName).First(&cfg).Error
	if err != nil {
		// 无精确匹配 → 尝试默认配置
		err = DB.Where("model_name = '*' AND enabled = 1").First(&cfg).Error
		if err != nil {
			// 连默认都没有 → 返回硬编码默认值
			return &SettlementConfig{
				ModelName:       "*",
				UnifiedCost:     0.001000,
				CommissionRate:  0.2000,
				PlatformFeeRate: 0.1000,
			}, nil
		}
	}
	return &cfg, nil
}

// TokenTransaction 交易流水：每笔 API 调用一条记录
type TokenTransaction struct {
	Id              int64   `json:"id" gorm:"primaryKey"`
	LogId           int     `json:"log_id" gorm:"default:0"`
	UserId          int     `json:"user_id" gorm:"default:0"`
	ModelName       string  `json:"model_name" gorm:"type:varchar(255)"`
	PromptTokens    int     `json:"prompt_tokens" gorm:"default:0"`
	CompletionTokens int    `json:"completion_tokens" gorm:"default:0"`

	ChannelId       int     `json:"channel_id" gorm:"default:0"`
	ChannelOwnerId  int     `json:"channel_owner_id" gorm:"default:0"`
	PromoterId      int     `json:"promoter_id" gorm:"default:0"`
	IsFallback      int     `json:"is_fallback" gorm:"default:0"`

	UnitPrice       float64 `json:"unit_price" gorm:"type:decimal(16,8)"`
	TotalAmount     float64 `json:"total_amount" gorm:"type:decimal(16,8)"`
	UnifiedCost     float64 `json:"unified_cost" gorm:"type:decimal(16,8)"`
	CommissionAmount float64 `json:"commission_amount" gorm:"type:decimal(16,8)"`
	PlatformFee     float64 `json:"platform_fee" gorm:"type:decimal(16,8)"`
	KeyProviderCost float64 `json:"key_provider_cost" gorm:"type:decimal(16,8);default:0"`

	CreatedTime     int64   `json:"created_time" gorm:"bigint"`
}

func (t *TokenTransaction) TableName() string {
	return "token_transaction"
}

// CreateTransaction 写入交易流水
func CreateTransaction(tx *TokenTransaction) error {
	tx.CreatedTime = time.Now().Unix()
	return DB.Create(tx).Error
}

// Reseller 渠道商/推广者
type Reseller struct {
	Id            int     `json:"id" gorm:"primaryKey"`
	UserId        int     `json:"user_id" gorm:"not null;uniqueIndex"`
	Name          string  `json:"name" gorm:"type:varchar(255)"`
	Description   string  `json:"description" gorm:"type:text"`
	ContactInfo   string  `json:"contact_info" gorm:"type:text"`
	AvatarUrl     string  `json:"avatar_url" gorm:"type:varchar(1024)"`
	Status        int     `json:"status" gorm:"default:1"`
	AffiliateCode string  `json:"affiliate_code" gorm:"type:varchar(64);uniqueIndex"`
	Balance       float64 `json:"balance" gorm:"type:decimal(16,8);default:0"`
	TotalEarned   float64 `json:"total_earned" gorm:"type:decimal(16,8);default:0"`
	CreatedTime   int64   `json:"created_time" gorm:"bigint"`
	UpdatedTime   int64   `json:"updated_time" gorm:"bigint"`
}

func (r *Reseller) TableName() string {
	return "reseller"
}

// AffiliateRelation 推广关系
type AffiliateRelation struct {
	Id          int   `json:"id" gorm:"primaryKey"`
	PromoterId  int   `json:"promoter_id" gorm:"not null"`
	ConsumerId  int   `json:"consumer_id" gorm:"not null;uniqueIndex"`
	CreatedTime int64 `json:"created_time" gorm:"bigint"`
	ExpiresAt   int64 `json:"expires_at" gorm:"bigint;default:0"`
}

func (a *AffiliateRelation) TableName() string {
	return "affiliate_relation"
}

// GetPromoterId 获取用户的推广人 ID（有则返回 reseller.id，无则返回 0）
func GetPromoterId(userId int) int {
	var rel AffiliateRelation
	err := DB.Where("consumer_id = ? AND (expires_at IS NULL OR expires_at = 0 OR expires_at > ?)", userId, time.Now().Unix()).First(&rel).Error
	if err != nil {
		return 0
	}
	return rel.PromoterId
}

// GetSettlementAmounts 计算一笔调用的分账金额
type SettlementAmounts struct {
	TotalAmount     float64 // 用户支付总金额
	UnifiedCost     float64 // 统一成本
	CommissionAmount float64 // 推广佣金
	PlatformFee     float64 // 平台抽佣
}

// GetTransactionFeeRate 获取交易手续费率
// domestic: 国内模型费率
// foreign: 国外模型费率
// foreignMinUsd: 国外最低 $ 手续费
func GetTransactionFeeRate() (domesticPct, foreignPct, foreignMinUsd float64) {
	domesticPct = 1.0
	foreignPct = 3.0
	foreignMinUsd = 5.0

	var cfg PlatformConfig
	if DB.Where("`key` = ?", "transaction_fee_domestic").First(&cfg).Error == nil {
		if v, err := strconv.ParseFloat(cfg.Value, 64); err == nil && v > 0 {
			domesticPct = v
		}
	}
	if DB.Where("`key` = ?", "transaction_fee_foreign").First(&cfg).Error == nil {
		if v, err := strconv.ParseFloat(cfg.Value, 64); err == nil && v > 0 {
			foreignPct = v
		}
	}
	if DB.Where("`key` = ?", "transaction_fee_foreign_min").First(&cfg).Error == nil {
		if v, err := strconv.ParseFloat(cfg.Value, 64); err == nil && v > 0 {
			foreignMinUsd = v
		}
	}
	return
}

func CalculateSettlement(unitPrice float64, totalTokens int, modelName string) *SettlementAmounts {
	cfg, _ := GetSettlementConfig(modelName)
	tokenK := float64(totalTokens) / 1000.0

	return &SettlementAmounts{
		TotalAmount:      unitPrice * tokenK,
		UnifiedCost:      cfg.UnifiedCost * tokenK,
		CommissionAmount: (unitPrice * tokenK) * cfg.CommissionRate,
		PlatformFee:      (unitPrice * tokenK) * cfg.PlatformFeeRate,
	}
}
