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

const (
	WithdrawStatusPending   = "pending"
	WithdrawStatusApproved  = "approved"
	WithdrawStatusRejected  = "rejected"
	WithdrawStatusCompleted = "completed"
	WithdrawMinAmount       = 100 // 最低提现金额（分）
)

// WithdrawalRequest — 提现申请
type WithdrawalRequest struct {
	Id                int        `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId            int        `json:"user_id" gorm:"index;not null"`
	Amount            int64      `json:"amount" gorm:"bigint;not null"`              // 提现金额（分）
	PlatformFeeAmount int64      `json:"platform_fee_amount" gorm:"bigint;default:0"` // 扣除入驻费（分）
	NetAmount         int64      `json:"net_amount" gorm:"bigint;default:0"`          // 实际到账（分）
	Status            string     `json:"status" gorm:"type:varchar(20);default:'pending'"`
	AccountInfo       string     `json:"account_info" gorm:"type:varchar(500)"`       // 收款账号
	Remark            string     `json:"remark" gorm:"type:text"`
	CreatedAt         time.Time  `json:"created_at"`
	ProcessedAt       *time.Time `json:"processed_at,omitempty"`
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

func CreateWithdrawal(w *WithdrawalRequest) error {
	// 扣除 pending 入驻费
	pendingFees, _ := GetPendingPlatformFees(w.UserId)
	var totalPending int64
	for _, f := range pendingFees {
		totalPending += f.FeeAmount
	}
	w.PlatformFeeAmount = totalPending
	w.NetAmount = w.Amount - totalPending
	if w.NetAmount < 0 {
		w.NetAmount = 0
	}
	if err := DB.Create(w).Error; err != nil {
		return err
	}
	// 标记入驻费为已扣除
	for _, f := range pendingFees {
		_ = DeductPlatformFee(f.Id)
	}
	return nil
}

func GetWithdrawalById(id int) (*WithdrawalRequest, error) {
	var w WithdrawalRequest
	err := DB.First(&w, "id = ?", id).Error
	return &w, err
}

func GetPendingWithdrawals(limit int) ([]WithdrawalRequest, error) {
	var list []WithdrawalRequest
	err := DB.Where("status = ?", WithdrawStatusPending).Order("id asc").Limit(limit).Find(&list).Error
	return list, err
}

func GetUserWithdrawableBalance(userId int) (int64, error) {
	var totalEarned int64
	DB.Model(&ProviderEarning{}).Where("user_id = ? AND status = ?", userId, EarningStatusSettled).
		Select("COALESCE(SUM(net_amount), 0)").Scan(&totalEarned)
	var withdrawn int64
	DB.Model(&WithdrawalRequest{}).Where("user_id = ? AND status IN ?",
		userId, []string{WithdrawStatusApproved, WithdrawStatusCompleted}).
		Select("COALESCE(SUM(net_amount), 0)").Scan(&withdrawn)
	var pendingFee int64
	DB.Model(&PlatformFeeRecord{}).Where("user_id = ? AND status = ?",
		userId, PlatformFeeStatusPending).
		Select("COALESCE(SUM(fee_amount), 0)").Scan(&pendingFee)
	available := totalEarned - withdrawn - pendingFee
	if available < 0 {
		available = 0
	}
	return available, nil
}

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

func GetWithdrawalByUser(userId int, limit int) ([]WithdrawalRequest, error) {
	var ws []WithdrawalRequest
	err := DB.Where("user_id = ?", userId).Order("id desc").Limit(limit).Find(&ws).Error
	return ws, err
}

func GetAllWithdrawals(status string, page, pageSize int) ([]WithdrawalRequest, int64, error) {
	var ws []WithdrawalRequest
	var total int64
	query := DB.Model(&WithdrawalRequest{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	err := query.Order("id desc").Offset(page * pageSize).Limit(pageSize).Find(&ws).Error
	return ws, total, err
}

func ApproveWithdrawal(id int, remark string) error {
	now := time.Now()
	return DB.Model(&WithdrawalRequest{}).Where("id = ? AND status = ?",
		id, WithdrawStatusPending).Updates(map[string]interface{}{
		"status":       WithdrawStatusApproved,
		"remark":       remark,
		"processed_at": now,
	}).Error
}

func CompleteWithdrawal(id int, remark string) error {
	now := time.Now()
	return DB.Model(&WithdrawalRequest{}).Where("id = ? AND status = ?",
		id, WithdrawStatusApproved).Updates(map[string]interface{}{
		"status":       WithdrawStatusCompleted,
		"remark":       remark,
		"processed_at": now,
	}).Error
}

func RejectWithdrawal(id int, remark string) error {
	now := time.Now()
	return DB.Model(&WithdrawalRequest{}).Where("id = ? AND status = ?",
		id, WithdrawStatusPending).Updates(map[string]interface{}{
		"status":       WithdrawStatusRejected,
		"remark":       remark,
		"processed_at": now,
	}).Error
}
