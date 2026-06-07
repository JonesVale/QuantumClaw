package model

import (
	"time"
)

// PaymentIdempotencyKey ensures every payment webhook is processed exactly once
type PaymentIdempotencyKey struct {
	ID          int    `json:"id" gorm:"primaryKey"`
	KeyStr      string `json:"key_str" gorm:"uniqueIndex;type:varchar(255);not null"`
	TradeNo     string `json:"trade_no" gorm:"index;type:varchar(128)"`
	Provider    string `json:"provider" gorm:"index;type:varchar(64)"`
	Status      string `json:"status" gorm:"type:varchar(32);default:'processing'"` // processing / completed / failed
	ProcessedAt int64  `json:"processed_at" gorm:"bigint"`
	ExpiresAt   int64  `json:"expires_at" gorm:"bigint"` // auto-cleanup after 7 days
}

func (PaymentIdempotencyKey) TableName() string { return "payment_idempotency_keys" }

// TryClaimIdempotency returns true if this key has not been processed before
func TryClaimIdempotency(provider, tradeNo string) (bool, error) {
	keyStr := provider + ":" + tradeNo
	now := time.Now().Unix()
	record := &PaymentIdempotencyKey{
		KeyStr:      keyStr,
		TradeNo:     tradeNo,
		Provider:    provider,
		Status:      "processing",
		ProcessedAt: now,
		ExpiresAt:   now + 7*24*3600, // 7 day TTL
	}
	err := DB.Create(record).Error
	if err != nil {
		if IsUniqueConstraintError(err) {
			return false, nil // already claimed
		}
		return false, err
	}
	return true, nil
}

// MarkIdempotencyCompleted marks a previously claimed key as successfully processed
func MarkIdempotencyCompleted(provider, tradeNo string) error {
	return DB.Model(&PaymentIdempotencyKey{}).
		Where("key_str = ?", provider+":"+tradeNo).
		Updates(map[string]interface{}{
			"status": "completed",
		}).Error
}

// CleanupExpiredIdempotencyKeys removes keys older than 7 days
func CleanupExpiredIdempotencyKeys() error {
	return DB.Where("expires_at < ?", time.Now().Unix()).Delete(&PaymentIdempotencyKey{}).Error
}

// IsUniqueConstraintError checks if the error is a unique constraint violation
func IsUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return containsAny(msg, "UNIQUE constraint", "Duplicate entry", "unique constraint", "duplicate key")
}

func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if len(s) >= len(sub) {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}
