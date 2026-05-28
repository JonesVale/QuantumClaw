package model

import (
	"github.com/quantumclaw/quantumclaw/common/helper"
	"gorm.io/gorm"
)

const (
	BalanceLogTypeRecharge    = "recharge"      // 充值
	BalanceLogTypeConsume     = "consume"       // 消费
	BalanceLogTypeRefund      = "refund"        // 退款
	BalanceLogTypeAdmin       = "admin"         // 管理员调整
	BalanceLogTypeTopUp       = "topup"         // 充值入账(佣金相关)
	BalanceLogTypeDebtRecover = "debt_recovery" // 对账追偿扣款
	BalanceLogTypeDebtDeduct  = "debt_deduct"   // 自动抵扣欠费
)

// BalanceLog 余额流水记录
type BalanceLog struct {
	Id        int    `json:"id"`
	UserId    int    `json:"user_id" gorm:"type:int;index"`
	Type      string `json:"type" gorm:"type:varchar(32)"`            // recharge/consume/refund/admin
	Amount    int64  `json:"amount" gorm:"bigint"`                    // 变动金额（分），正=增加，负=减少
	Balance   int64  `json:"balance" gorm:"bigint"`                   // 变动后余额
	ChannelId int    `json:"channel_id" gorm:"type:int;default:0"`    // 消费时关联的渠道
	Remark    string `json:"remark" gorm:"type:varchar(255)"`         // 备注
	CreatedAt int64  `json:"created_at" gorm:"bigint"`
}

func (BalanceLog) TableName() string {
	return "balance_logs"
}

// CreateBalanceLog 创建余额流水（使用默认 DB）
func CreateBalanceLog(userId int, logType string, amount int64, balance int64, channelId int, remark string) error {
	return createBalanceLogWithDB(DB, userId, logType, amount, balance, channelId, remark)
}

// CreateBalanceLogTx 在指定事务内创建余额流水（避免 SQLITE_BUSY）
func CreateBalanceLogTx(tx *gorm.DB, userId int, logType string, amount int64, balance int64, channelId int, remark string) error {
	return createBalanceLogWithDB(tx, userId, logType, amount, balance, channelId, remark)
}

func createBalanceLogWithDB(d *gorm.DB, userId int, logType string, amount int64, balance int64, channelId int, remark string) error {
	return d.Create(&BalanceLog{
		UserId:    userId,
		Type:      logType,
		Amount:    amount,
		Balance:   balance,
		ChannelId: channelId,
		Remark:    remark,
		CreatedAt: helper.GetTimestamp(),
	}).Error
}

// GetUserBalanceLogs 获取用户余额流水
func GetUserBalanceLogs(userId int, limit int) ([]BalanceLog, error) {
	var logs []BalanceLog
	err := DB.Where("user_id = ?", userId).Order("id desc").Limit(limit).Find(&logs).Error
	return logs, err
}
