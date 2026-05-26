package model

import (
	"github.com/quantumclaw/quantumclaw/common/logger"
)

// ReferencePrice stores the official reference price for each model.
// Seeded from ModelRatio on startup; updated monthly from official pricing pages.
type ReferencePrice struct {
	ID          uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	ModelName   string  `gorm:"type:varchar(255);not null;uniqueIndex:idx_ref_model" json:"model_name"`
	Provider    string  `gorm:"type:varchar(100);not null;index" json:"provider"`
	InputPrice  float64 `gorm:"type:decimal(12,8);default:0" json:"input_price"`
	OutputPrice float64 `gorm:"type:decimal(12,8);default:0" json:"output_price"`
	Currency    string  `gorm:"type:varchar(10);default:'USD'" json:"currency"`
	Source      string  `gorm:"type:varchar(200);default:'modelratio'" json:"source"`
	FetchedAt   int64   `json:"fetched_at"`
}

func (ReferencePrice) TableName() string {
	return "reference_prices"
}

// GetReferencePrice returns the reference price for a model name.
// Returns nil if not found.
func GetReferencePrice(modelName string) *ReferencePrice {
	var rp ReferencePrice
	if err := DB.Where("model_name = ?", modelName).First(&rp).Error; err != nil {
		return nil
	}
	return &rp
}

// InitReferencePriceTable ensures the table exists.
func InitReferencePriceTable() {
	if err := DB.AutoMigrate(&ReferencePrice{}); err != nil {
		logger.SysError("InitReferencePriceTable AutoMigrate failed: " + err.Error())
	}
}
