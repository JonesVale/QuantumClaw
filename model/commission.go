package model

import (
	"time"

	"gorm.io/gorm"
)

// CommissionSetting — 推广佣金配置（全局）
type CommissionSetting struct {
	Id             int     `json:"id" gorm:"primaryKey;autoIncrement"`
	Enabled        bool    `json:"enabled" gorm:"default:true"`
	RegisterReward int64   `json:"register_reward" gorm:"default:0"`       // 好友注册奖励（固定额度）
	ConsumeRate    float64 `json:"consume_rate" gorm:"type:decimal(5,4);default:0.1"` // 消费返佣比例（0.1=10%）
	MinWithdraw    int64   `json:"min_withdraw" gorm:"default:10000"`       // 最低提现额度
	UpdatedAt      time.Time `json:"updated_at"`
}

// CommissionRecord — 佣金流水
type CommissionRecord struct {
	Id          int       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId      int       `json:"user_id" gorm:"index;not null"`      // 获得佣金的人
	FromUserId  int       `json:"from_user_id" gorm:"index"`           // 来源用户（谁消费了）
	Type        string    `json:"type" gorm:"type:varchar(20)"`        // register | consume
	Amount      int64     `json:"amount" gorm:"bigint;not null"`       // 佣金数额（微额度）
	Status      string    `json:"status" gorm:"type:varchar(20);default:'pending'"` // pending | settled | withdrawn
	Description string    `json:"description" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	SettledAt   *time.Time `json:"settled_at,omitempty"`
}

// WithdrawalRequest — 提现申请
type WithdrawalRequest struct {
	Id          int       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId      int       `json:"user_id" gorm:"index;not null"`
	Amount      int64     `json:"amount" gorm:"bigint;not null"`       // 提现额度
	Status      string    `json:"status" gorm:"type:varchar(20);default:'pending'"` // pending | approved | rejected
	AccountInfo string    `json:"account_info" gorm:"type:varchar(500)"` // 提现账户信息
	Remark      string    `json:"remark" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
}

func InitCommissionTables() {
	DB.AutoMigrate(&CommissionSetting{}, &CommissionRecord{}, &WithdrawalRequest{})
}

// GetCommissionSetting 获取佣金配置
func GetCommissionSetting() (*CommissionSetting, error) {
	var s CommissionSetting
	err := DB.First(&s).Error
	if err != nil {
		// 返回默认值
		return &CommissionSetting{Enabled: true, ConsumeRate: 0.1, MinWithdraw: 10000}, nil
	}
	return &s, nil
}

// SaveCommissionSetting 保存佣金配置
func SaveCommissionSetting(s *CommissionSetting) error {
	var existing CommissionSetting
	if DB.First(&existing).Error != nil {
		return DB.Create(s).Error
	}
	return DB.Model(&CommissionSetting{}).Where("id = ?", existing.Id).Updates(s).Error
}

// CreateCommissionRecord 创建佣金记录
func CreateCommissionRecord(record *CommissionRecord) error {
	return DB.Create(record).Error
}

// GetUserCommissionRecords 获取用户佣金记录
func GetUserCommissionRecords(userId int, limit, offset int) ([]CommissionRecord, int64, error) {
	var records []CommissionRecord
	var total int64
	DB.Model(&CommissionRecord{}).Where("user_id = ?", userId).Count(&total)
	err := DB.Where("user_id = ?", userId).Order("id desc").Limit(limit).Offset(offset).Find(&records).Error
	return records, total, err
}

// GetUserTotalCommission 获取用户佣金总额
func GetUserTotalCommission(userId int) (int64, error) {
	var total int64
	err := DB.Model(&CommissionRecord{}).Where("user_id = ? AND status != ?", userId, "withdrawn").Select("COALESCE(SUM(amount), 0)").Scan(&total).Error
	return total, err
}

// CreateWithdrawal 创建提现申请

// RewardInviterOnConsume — 用户消费时自动返佣给邀请人
func RewardInviterOnConsume(userId int, consumeAmount int64) {
	if consumeAmount <= 0 {
		return
	}
	var user User
	if err := DB.First(&user, userId).Error; err != nil || user.InviterId <= 0 {
		return
	}
	setting, err := GetCommissionSetting()
	if err != nil || !setting.Enabled || setting.ConsumeRate <= 0 {
		return
	}
	reward := int64(float64(consumeAmount) * setting.ConsumeRate)
	if reward <= 0 {
		return
	}
	CreateCommissionRecord(&CommissionRecord{
		UserId:      user.InviterId,
		FromUserId:  userId,
		Type:        "consume",
		Amount:      reward,
		Status:      "settled",
		Description: "referral consumption reward",
	})
	DB.Model(&User{}).Where("id = ?", user.InviterId).
		UpdateColumn("quota", gorm.Expr("quota + ?", reward))
}

func CreateWithdrawal(w *WithdrawalRequest) error {
	return DB.Create(w).Error
}

// GetWithdrawalsByUser 获取用户提现记录
func GetWithdrawalsByUser(userId int) ([]WithdrawalRequest, error) {
	var ws []WithdrawalRequest
	err := DB.Where("user_id = ?", userId).Order("id desc").Find(&ws).Error
	return ws, err
}

// GetAllWithdrawals 管理员获取所有提现申请
func GetAllWithdrawals(status string) ([]WithdrawalRequest, error) {
	var ws []WithdrawalRequest
	query := DB.Order("id desc")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Find(&ws).Error
	return ws, err
}

// ProcessWithdrawal 处理提现
func ProcessWithdrawal(id int, status, remark string) error {
	now := time.Now()
	return DB.Model(&WithdrawalRequest{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       status,
		"remark":       remark,
		"processed_at": now,
	}).Error
}
