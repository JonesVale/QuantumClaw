package model

import (
	"fmt"
	"time"

	"github.com/quantumclaw/quantumclaw/common/helper"
	"github.com/quantumclaw/quantumclaw/common/logger"
)

// ==================== 平台收入账本 ====================
// PlatformIncome 记录每笔平台收益（自动入账流水）
type PlatformIncome struct {
	Id              int64   `json:"id" gorm:"primaryKey;autoIncrement"`
	ChannelId       int     `json:"channel_id" gorm:"default:0;index"`         // 渠道 ID
	ConsumerUserId  int     `json:"consumer_user_id" gorm:"index"`             // 消费者用户 ID
	TotalAmount     int64   `json:"total_amount" gorm:"bigint;default:0"`      // 用户实付（分）
	CommissionAmount int64  `json:"commission_amount" gorm:"bigint;default:0"`  // 平台抽成（分）
	Source          string  `json:"source" gorm:"type:varchar(32);default:''"` // relay_billing / quantum / monthly_fee 等
	ModelName       string  `json:"model_name" gorm:"type:varchar(255)"`
	Status          string  `json:"status" gorm:"type:varchar(20);default:'settled'"` // settled / pending / withdrawn
	CreatedAt       int64   `json:"created_at" gorm:"bigint"`
}

func (PlatformIncome) TableName() string {
	return "platform_income"
}

// CreatePlatformIncomeRecord 写入平台收入流水
func CreatePlatformIncomeRecord(income *PlatformIncome) error {
	if income.CreatedAt == 0 {
		income.CreatedAt = helper.GetTimestamp()
	}
	return DB.Create(income).Error
}

// ==================== 平台余额管理 ====================
// 平台余额存储在 platform_config 表中，key = "platform_balance_cents"

const (
	PlatformConfigKeyBalance = "platform_balance_cents" // 平台可提现余额（分）
	PlatformConfigKeyTotalIncome = "platform_total_income_cents" // 历史累计收入（分）
)

// AddPlatformIncome 累加平台收入到余额（原子操作）
func AddPlatformIncome(amount int64) error {
	if amount <= 0 {
		return nil
	}
	err := DB.Exec(
		"INSERT INTO platform_config (`key`, value, updated_time) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE value = CAST(value AS SIGNED) + ?, updated_time = ?",
		PlatformConfigKeyBalance, amount, time.Now().Unix(), amount, time.Now().Unix(),
	).Error
	if err != nil {
		return err
	}

	// 同时累加历史总收入
	return DB.Exec(
		"INSERT INTO platform_config (`key`, value, updated_time) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE value = CAST(value AS SIGNED) + ?, updated_time = ?",
		PlatformConfigKeyTotalIncome, amount, time.Now().Unix(), amount, time.Now().Unix(),
	).Error
}

// AddPlatformBalance 平台余额增加（供 service 层调用）
func AddPlatformBalance(amount int64) error {
	return AddPlatformIncome(amount)
}

// GetPlatformBalance 获取当前平台可提现余额（分）
func GetPlatformBalance() int64 {
	var cfg PlatformConfig
	if err := DB.Where("`key` = ?", PlatformConfigKeyBalance).First(&cfg).Error; err != nil {
		return 0
	}
	var val int64
	if _, err := fmt.Sscanf(cfg.Value, "%d", &val); err != nil {
		return 0
	}
	return val
}

// GetPlatformTotalIncome 获取历史累计收入（分）
func GetPlatformTotalIncome() int64 {
	var cfg PlatformConfig
	if err := DB.Where("`key` = ?", PlatformConfigKeyTotalIncome).First(&cfg).Error; err != nil {
		return 0
	}
	var val int64
	if _, err := fmt.Sscanf(cfg.Value, "%d", &val); err != nil {
		return 0
	}
	return val
}

// GetPlatformIncomeHistory 获取平台收入流水（分页）
func GetPlatformIncomeHistory(page, pageSize int) ([]PlatformIncome, int64, error) {
	var list []PlatformIncome
	var total int64
	DB.Model(&PlatformIncome{}).Count(&total)
	err := DB.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// GetPlatformIncomeSummary 获取平台收入汇总
// 返回：(今日收入, 本月收入, 总余额, 累计总收入)
func GetPlatformIncomeSummary() (today int64, thisMonth int64, balance int64, totalIncome int64) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Unix()

	DB.Model(&PlatformIncome{}).Where("created_at >= ?", startOfDay).Select("COALESCE(SUM(commission_amount), 0)").Scan(&today)
	DB.Model(&PlatformIncome{}).Where("created_at >= ?", startOfMonth).Select("COALESCE(SUM(commission_amount), 0)").Scan(&thisMonth)
	balance = GetPlatformBalance()
	totalIncome = GetPlatformTotalIncome()

	return
}

// initPlatformConfig 初始化平台配置表（确保 balance key 存在）
func initPlatformConfig() {
	keys := []string{PlatformConfigKeyBalance, PlatformConfigKeyTotalIncome}
	for _, key := range keys {
		var count int64
		DB.Model(&PlatformConfig{}).Where("`key` = ?", key).Count(&count)
		if count == 0 {
			DB.Create(&PlatformConfig{Key: key, Value: "0", UpdatedTime: time.Now().Unix()})
			logger.SysLogf("platform_config init: %s = 0", key)
		}
	}
}
