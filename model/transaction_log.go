package model

import "time"

type TransactionLog struct {
	Id          int       `json:"id" gorm:"primaryKey"`
	UserId      int       `json:"user_id" gorm:"index"`
	Action      string    `json:"action" gorm:"type:varchar(100)"` // topup, refund, admin_adjust, withdrawal
	Amount      int64     `json:"amount"`
	BeforeQuota int64     `json:"before_quota"`
	AfterQuota  int64     `json:"after_quota"`
	IP          string    `json:"ip" gorm:"type:varchar(45)"`
	UserAgent   string    `json:"user_agent" gorm:"type:varchar(500)"`
	TradeNo     string    `json:"trade_no" gorm:"type:varchar(100)"`
	Status      string    `json:"status" gorm:"type:varchar(20)"`
	AdminId     int       `json:"admin_id"` // who performed the action (0 = system/user self)
	Remark      string    `json:"remark" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
}

func (TransactionLog) TableName() string {
	return "transaction_logs"
}

// InsertTransactionLog creates a new transaction log entry
func InsertTransactionLog(log *TransactionLog) error {
	return DB.Create(log).Error
}

// GetUserTransactionLogs returns paginated transaction logs for a user
func GetUserTransactionLogs(userId int, page int, pageSize int) ([]TransactionLog, int64, error) {
	var logs []TransactionLog
	var total int64

	if err := DB.Model(&TransactionLog{}).Where("user_id = ?", userId).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := DB.Where("user_id = ?", userId).Order("id DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetRecentTransactionLogs returns the most recent N transaction logs for a user
func GetRecentTransactionLogs(userId int, limit int) ([]TransactionLog, error) {
	var logs []TransactionLog
	err := DB.Where("user_id = ?", userId).Order("id DESC").Limit(limit).Find(&logs).Error
	return logs, err
}
