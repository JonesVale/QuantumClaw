package model

import "github.com/quantumclaw/quantumclaw/common/helper"

// PlatformFeeConfig defines tiered membership fee rates
type PlatformFeeConfig struct {
	Tier      StoreTier `json:"tier" gorm:"primaryKey;type:varchar(32)"`
	Rate      float64   `json:"rate" gorm:"type:decimal(5,2);not null"`
	MinSkip   int64     `json:"min_skip" gorm:"bigint;default:10000"`
	UpdatedBy int       `json:"updated_by" gorm:"int;default:0"`
	UpdatedAt int64     `json:"updated_at" gorm:"bigint"`
}

func (PlatformFeeConfig) TableName() string { return "platform_fee_configs" }

func SeedDefaultFeeConfigs() error {
	defaults := []PlatformFeeConfig{
		{Tier: StoreTierBasic, Rate: 10.0, MinSkip: 10000},
		{Tier: StoreTierGold, Rate: 8.0, MinSkip: 10000},
		{Tier: StoreTierFlagship, Rate: 5.0, MinSkip: 10000},
	}
	for _, cfg := range defaults {
		var existing PlatformFeeConfig
		if err := DB.Where("tier = ?", cfg.Tier).First(&existing).Error; err != nil {
			cfg.UpdatedAt = helper.GetTimestamp()
			if err := DB.Create(&cfg).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func GetFeeConfig(tier StoreTier) (*PlatformFeeConfig, error) {
	var cfg PlatformFeeConfig
	err := DB.Where("tier = ?", tier).First(&cfg).Error
	return &cfg, err
}

func GetAllFeeConfigs() ([]PlatformFeeConfig, error) {
	var cfgs []PlatformFeeConfig
	err := DB.Order("rate DESC").Find(&cfgs).Error
	return cfgs, err
}

func UpdateFeeConfig(tier StoreTier, rate float64, minSkip int64, updatedBy int) error {
	return DB.Model(&PlatformFeeConfig{}).Where("tier = ?", tier).Updates(map[string]interface{}{
		"rate":       rate,
		"min_skip":   minSkip,
		"updated_by": updatedBy,
		"updated_at": helper.GetTimestamp(),
	}).Error
}
