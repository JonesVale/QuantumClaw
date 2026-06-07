package model

import (
	"github.com/quantumclaw/quantumclaw/common/helper"
)

// ConsumeRecord tracks each billing deduction for user transparency
type ConsumeRecord struct {
	ID           int    `json:"id" gorm:"primaryKey"`
	UserID       int    `json:"user_id" gorm:"index;not null"`
	TokenID      int    `json:"token_id" gorm:"index;default:0"`
	ChannelID    int    `json:"channel_id" gorm:"default:0"`
	ModelName    string `json:"model_name" gorm:"type:varchar(128)"`
	PriceCents   int64  `json:"price_cents" gorm:"bigint;not null"`
	QuotaConsumed int64 `json:"quota_consumed" gorm:"bigint;default:0"`
	Source       string `json:"source" gorm:"type:varchar(32)"` // cash / commission / quota / subscription / debt
	BalanceAfter int64  `json:"balance_after" gorm:"bigint;default:0"`
	CreatedAt    int64  `json:"created_at" gorm:"bigint;index"`
}

func (ConsumeRecord) TableName() string { return "consume_records" }

// RecordConsume logs a single billing deduction event
func RecordConsume(userID, tokenID, channelID int, modelName string, priceCents, quotaConsumed int64, source string, balanceAfter int64) {
	DB.Create(&ConsumeRecord{
		UserID:        userID,
		TokenID:       tokenID,
		ChannelID:     channelID,
		ModelName:     modelName,
		PriceCents:    priceCents,
		QuotaConsumed: quotaConsumed,
		Source:        source,
		BalanceAfter:  balanceAfter,
		CreatedAt:     helper.GetTimestamp(),
	})
}

// GetUserConsumeRecords returns paginated consumption records for a user
func GetUserConsumeRecords(userID int, page, pageSize int) ([]ConsumeRecord, int64, error) {
	var records []ConsumeRecord
	var total int64
	DB.Model(&ConsumeRecord{}).Where("user_id = ?", userID).Count(&total)
	err := DB.Where("user_id = ?", userID).Order("id desc").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&records).Error
	if records == nil {
		records = []ConsumeRecord{}
	}
	return records, total, err
}

// GetUserConsumeSummary returns spending summary for a period
func GetUserConsumeSummary(userID int, days int) (totalPrice int64, totalQuota int64, err error) {
	err = DB.Model(&ConsumeRecord{}).Where("user_id = ? AND created_at > ?",
		userID, helper.GetTimestamp()-int64(days*86400)).
		Select("COALESCE(SUM(price_cents),0) as price, COALESCE(SUM(quota_consumed),0) as quota").
		Row().Scan(&totalPrice, &totalQuota)
	return
}
