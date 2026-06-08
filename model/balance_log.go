package model

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/quantumclaw/quantumclaw/common/helper"
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
	Id           int    `json:"id"`
	UserId       int    `json:"user_id" gorm:"type:int;index"`
	Type         string `json:"type" gorm:"type:varchar(32)"`             // recharge/consume/refund/admin
	Amount       int64  `json:"amount" gorm:"bigint"`                    // 变动金额（分），正=增加，负=减少
	Balance      int64  `json:"balance" gorm:"bigint"`                   // 变动后余额
	ChannelId    int    `json:"channel_id" gorm:"type:int;default:0"`    // 消费时关联的渠道
	Remark       string `json:"remark" gorm:"type:varchar(255)"`         // 备注
	RelatedLogId int    `json:"related_log_id" gorm:"type:int;default:0;index"` // 关联的原交易ID（用于回滚），0=非回滚记录
	CreatedAt    int64  `json:"created_at" gorm:"bigint"`
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

// ==================== Audit / Rollback ====================

// GetAllBalanceLogs 管理员查询所有用户余额流水（分页），支持按 user_id 过滤
func GetAllBalanceLogs(userId int, page int, pageSize int) ([]BalanceLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var logs []BalanceLog
	var total int64
	q := DB.Model(&BalanceLog{})
	if userId > 0 {
		q = q.Where("user_id = ?", userId)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}

// GetBalanceLogByID 根据 ID 获取单条流水记录
func GetBalanceLogByID(id int) (*BalanceLog, error) {
	var log BalanceLog
	if err := DB.Where("id = ?", id).First(&log).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

// RollbackLogAlreadyExists 检查指定流水是否已被回滚
func RollbackLogAlreadyExists(relatedLogId int) (bool, error) {
	var count int64
	err := DB.Model(&BalanceLog{}).
		Where("related_log_id = ? AND type = ?", relatedLogId, BalanceLogTypeRefund).
		Count(&count).Error
	return count > 0, err
}

// RollbackBalanceLog 回滚一笔余额流水
// 仅支持回滚 consume/admin/topup 类型的流水
// 回滚时：创建一条 type=refund 的新记录，金额取反，并自动充值到用户余额
func RollbackBalanceLog(originalLogId int, operatorUserId int, remark string) (*BalanceLog, error) {
	// 1. 读取原流水
	original, err := GetBalanceLogByID(originalLogId)
	if err != nil {
		return nil, fmt.Errorf("original log %d not found: %w", originalLogId, err)
	}

	// 2. 类型校验：不能回滚 refund/debt_recovery/debt_deduct 类型
	switch original.Type {
	case BalanceLogTypeRefund:
		return nil, errors.New("cannot rollback a refund record")
	case BalanceLogTypeDebtRecover:
		return nil, errors.New("cannot rollback a debt recovery record")
	case BalanceLogTypeDebtDeduct:
		return nil, errors.New("cannot rollback a debt deduction record")
	}

	// 3. 检查是否已经被回滚过
	already, err := RollbackLogAlreadyExists(originalLogId)
	if err != nil {
		return nil, fmt.Errorf("check rollback: %w", err)
	}
	if already {
		return nil, errors.New("this transaction has already been rolled back")
	}

	// 4. 计算反向金额和回滚后余额
	rollbackAmount := -original.Amount
	currentBalance, err := GetUserCashBalance(original.UserId)
	if err != nil {
		return nil, fmt.Errorf("get user balance: %w", err)
	}

	// 只有 amount 为正（即回滚增加余额）时才执行加余额操作
	newBalance := currentBalance
	if rollbackAmount > 0 {
		if err := PlusUserCashBalance(original.UserId, rollbackAmount); err != nil {
			return nil, fmt.Errorf("plus balance: %w", err)
		}
		newBalance = currentBalance + rollbackAmount
	} else if rollbackAmount < 0 {
		// 回滚负金额 = 从用户余额中扣减
		if err := MinusUserCashBalance(original.UserId, -rollbackAmount); err != nil {
			return nil, fmt.Errorf("minus balance: %w", err)
		}
		newBalance = currentBalance - (-rollbackAmount)
		if newBalance < 0 {
			newBalance = 0
		}
	}

	// 5. 创建回滚流水记录
	finalRemark := fmt.Sprintf("回滚交易 #%d (%s: %d分)", originalLogId, original.Type, original.Amount)
	if remark != "" {
		finalRemark = remark + " | " + finalRemark
	}

	rollbackLog := &BalanceLog{
		UserId:       original.UserId,
		Type:         BalanceLogTypeRefund,
		Amount:       rollbackAmount,
		Balance:      newBalance,
		ChannelId:    original.ChannelId,
		Remark:       finalRemark,
		RelatedLogId: originalLogId,
		CreatedAt:    helper.GetTimestamp(),
	}

	if err := DB.Create(rollbackLog).Error; err != nil {
		return nil, fmt.Errorf("create rollback log: %w", err)
	}

	return rollbackLog, nil
}
