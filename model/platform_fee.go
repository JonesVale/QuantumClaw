package model

import (
	"strconv"
	"time"

	"github.com/quantumclaw/quantumclaw/common/helper"
)

const (
	PlatformFeeStatusPending  = "pending"   // 待扣除（已算出入驻费，等提现时扣）
	PlatformFeeStatusDeducted = "deducted"  // 已扣除（提现时扣掉了）
	PlatformFeeStatusSkipped  = "skipped"   // 已跳过（月营业额 < ¥100）
)

// PlatformFeeRecord 入驻费结算记录
type PlatformFeeRecord struct {
	Id           int    `json:"id"`
	UserId       int    `json:"user_id" gorm:"type:int;index"`             // 供应商用户 ID
	Period       string `json:"period" gorm:"type:varchar(7)"`            // 结算月份 YYYY-MM
	TotalRevenue int64  `json:"total_revenue" gorm:"bigint"`               // 当月总营业额（分）
	FeeRate      float64 `json:"fee_rate" gorm:"type:decimal(5,2)"`        // 入驻费率（%）
	FeeAmount    int64  `json:"fee_amount" gorm:"bigint"`                  // 入驻费金额（分）
	Status       string `json:"status" gorm:"type:varchar(32);default:'pending'"` // pending/deducted/skipped
	CreatedAt    int64  `json:"created_at" gorm:"bigint"`
	DeductedAt   int64  `json:"deducted_at" gorm:"bigint;default:0"`       // 实际扣除时间
}

func (PlatformFeeRecord) TableName() string {
	return "platform_fee_records"
}

// CreatePlatformFeeRecord 创建入驻费结算记录
func CreatePlatformFeeRecord(userId int, period string, totalRevenue int64, feeRate float64, feeAmount int64, status string) error {
	return DB.Create(&PlatformFeeRecord{
		UserId:       userId,
		Period:       period,
		TotalRevenue: totalRevenue,
		FeeRate:      feeRate,
		FeeAmount:    feeAmount,
		Status:       status,
		CreatedAt:    helper.GetTimestamp(),
	}).Error
}

// GetPendingPlatformFees 获取待扣入驻费记录
func GetPendingPlatformFees(userId int) ([]PlatformFeeRecord, error) {
	var fees []PlatformFeeRecord
	err := DB.Where("user_id = ? AND status = ?", userId, PlatformFeeStatusPending).
		Order("id asc").Find(&fees).Error
	return fees, err
}

// GetPlatformFeeByPeriod 获取某供应商某月入驻费记录
func GetPlatformFeeByPeriod(userId int, period string) (*PlatformFeeRecord, error) {
	var fee PlatformFeeRecord
	err := DB.Where("user_id = ? AND period = ?", userId, period).First(&fee).Error
	return &fee, err
}

// DeductPlatformFee 将入驻费标记为已扣除
func DeductPlatformFee(id int) error {
	return DB.Model(&PlatformFeeRecord{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      PlatformFeeStatusDeducted,
		"deducted_at": helper.GetTimestamp(),
	}).Error
}

// CalculateMonthlyPlatformFee 计算某供应商某月入驻费
// 月营业额 < ¥100 → 跳过，返回 false
// 否则按总营业额 × 5% 计算
func CalculateMonthlyPlatformFee(userId int, year int, month time.Month) (bool, error) {
	period := time.Date(year, month, 1, 0, 0, 0, 0, time.Local).Format("2006-01")

	// 查询当月已结算的入驻费记录
	existing, _ := GetPlatformFeeByPeriod(userId, period)
	if existing != nil {
		// 已有记录，不重复创建
		return existing.Status != PlatformFeeStatusSkipped, nil
	}

	// 统计当月该供应商所有已结算收益的总营业额
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

	if totalRevenue < 100 { // < ¥100 跳过
		err = CreatePlatformFeeRecord(userId, period, totalRevenue, 0, 0, PlatformFeeStatusSkipped)
		return false, err
	}

	// 入驻费 = 总营业额 × 配置费率
	feeRate := getPlatformFeeRate()
	feeAmount := int64(float64(totalRevenue) * feeRate / 100.0)
	err = CreatePlatformFeeRecord(userId, period, totalRevenue, feeRate, feeAmount, PlatformFeeStatusPending)
	return true, err
}

// AutoSettleMonthlyFees 自动结算所有供应商上月入驻费（定时任务调用）
// getPlatformFeeRate 从 platform_config 读取入驻费率（默认 5%）
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

	// 获取所有有收益的供应商
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
