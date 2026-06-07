package model

import (
	"strconv"
	"time"

	"github.com/quantumclaw/quantumclaw/common/helper"
)

const (
	PlatformFeeStatusPending  = "pending"   // ?????????????????
	PlatformFeeStatusDeducted = "deducted"  // ???????????
	PlatformFeeStatusSkipped  = "skipped"   // ???????? < ?100?
)

// PlatformFeeRecord ???????
type PlatformFeeRecord struct {
	Id           int     `json:"id"`
	StoreID      int     `json:"store_id" gorm:"type:int;default:0;index"` // ?? ID
	UserId       int     `json:"user_id" gorm:"type:int;index"`             // ????? ID
	Period       string  `json:"period" gorm:"type:varchar(7)"`            // ???? YYYY-MM
	TotalRevenue int64   `json:"total_revenue" gorm:"bigint"`               // ?????????
	FeeRate      float64 `json:"fee_rate" gorm:"type:decimal(5,2)"`        // ?????%?
	FeeAmount    int64   `json:"fee_amount" gorm:"bigint"`                  // ????????
	Status       string  `json:"status" gorm:"type:varchar(32);default:'pending'"` // pending/deducted/skipped
	CreatedAt    int64   `json:"created_at" gorm:"bigint"`
	DeductedAt   int64   `json:"deducted_at" gorm:"bigint;default:0"`       // ??????
}

func (PlatformFeeRecord) TableName() string {
	return "platform_fee_records"
}

// CreatePlatformFeeRecord ?????????
func CreatePlatformFeeRecord(storeID, userId int, period string, totalRevenue int64, feeRate float64, feeAmount int64, status string) error {
	return DB.Create(&PlatformFeeRecord{
		StoreID:      storeID,
		UserId:       userId,
		Period:       period,
		TotalRevenue: totalRevenue,
		FeeRate:      feeRate,
		FeeAmount:    feeAmount,
		Status:       status,
		CreatedAt:    helper.GetTimestamp(),
	}).Error
}

// GetPendingPlatformFees ?????????
func GetPendingPlatformFees(userId int) ([]PlatformFeeRecord, error) {
	var fees []PlatformFeeRecord
	err := DB.Where("user_id = ? AND status = ?", userId, PlatformFeeStatusPending).
		Order("id asc").Find(&fees).Error
	return fees, err
}

// GetPlatformFeeByPeriod ?????????????
func GetPlatformFeeByPeriod(userId int, period string) (*PlatformFeeRecord, error) {
	var fee PlatformFeeRecord
	err := DB.Where("user_id = ? AND period = ?", userId, period).First(&fee).Error
	return &fee, err
}

// DeductPlatformFee ??????????
func DeductPlatformFee(id int) error {
	return DB.Model(&PlatformFeeRecord{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      PlatformFeeStatusDeducted,
		"deducted_at": helper.GetTimestamp(),
	}).Error
}

// CalculateMonthlyPlatformFee ???????????
// ???? < ?100 ? ????? false
// ??????? ? 5% ??
func CalculateMonthlyPlatformFee(userId int, year int, month time.Month) (bool, error) {
	period := time.Date(year, month, 1, 0, 0, 0, 0, time.Local).Format("2006-01")

	// ?????????????
	existing, _ := GetPlatformFeeByPeriod(userId, period)
	if existing != nil {
		// ??????????
		return existing.Status != PlatformFeeStatusSkipped, nil
	}

	// ????????????????????
	var totalRevenue int64
	startOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.Local).Unix()
	endOfMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, time.Local).Unix()

	err := DB.Model(&ProviderEarning{}).
		Where("user_id = ? AND created_at >= ? AND created_at < ? AND status = ?",
			userId, startOfMonth, endOfMonth, EarningStatusSettled).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&totalRevenue).Error
	if err != nil {
		return false, err
	}

	// ???? ID
	storeID := 0
	var store Store
	if DB.Where("user_id = ?", userId).First(&store).Error == nil {
		storeID = store.ID
	}

	if totalRevenue < 100 { // < ?100 ??
		err = CreatePlatformFeeRecord(storeID, userId, period, totalRevenue, 0, 0, PlatformFeeStatusSkipped)
		return false, err
	}

	// ??? = ???? ? ????
	feeRate := getPlatformFeeRate()
	feeAmount := int64(float64(totalRevenue) * feeRate / 100.0)
	err = CreatePlatformFeeRecord(storeID, userId, period, totalRevenue, feeRate, feeAmount, PlatformFeeStatusPending)
	return true, err
}

// AutoSettleMonthlyFees ??????????????????????
// getPlatformFeeRate ? platform_config ????????? 5%?
func getPlatformFeeRate() float64 {
	var cfg PlatformConfig
	if DB.Where("`key` = ?", "platform_fee_rate_percent").First(&cfg).Error == nil {
		if v, err := strconv.ParseFloat(cfg.Value, 64); err == nil && v > 0 {
			return v
		}
	}
	return 5.0
}

func AutoSettleMonthlyFees() {
	now := time.Now()
	prevMonth := now.AddDate(0, -1, 0)
	year, month, _ := prevMonth.Date()

	// ???????????
	var userIds []int
	DB.Model(&ProviderEarning{}).
		Select("DISTINCT user_id").
		Where("created_at >= ? AND created_at < ?",
			time.Date(year, month, 1, 0, 0, 0, 0, time.Local).Unix(),
			time.Date(year, month+1, 1, 0, 0, 0, 0, time.Local).Unix()).
		Pluck("user_id", &userIds)

	for _, uid := range userIds {
		_, _ = CalculateMonthlyPlatformFee(uid, year, month)
	}
}
