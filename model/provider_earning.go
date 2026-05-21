package model

import "github.com/quantumclaw/quantumclaw/common/helper"

const (
	EarningStatusPending   = "pending"   // 待结算
	EarningStatusSettled   = "settled"   // 已结算
	EarningStatusWithdrawn = "withdrawn" // 已提现
)

// ProviderEarning 供应商收益记录
type ProviderEarning struct {
	Id                int    `json:"id"`
	UserId            int    `json:"user_id" gorm:"type:int;index"`                       // 供应商用户 ID
	ChannelId         int    `json:"channel_id" gorm:"type:int;index"`                     // 产生收益的渠道
	ConsumerId        int    `json:"consumer_id" gorm:"type:int;default:0"`                // 消费者用户 ID
	TotalAmount       int64  `json:"total_amount" gorm:"bigint"`                           // 消费者支付总额（分）
	CommissionAmount  int64  `json:"commission_amount" gorm:"bigint"`                      // 平台交易抽成（分）
	NetAmount         int64  `json:"net_amount" gorm:"bigint"`                             // 供应商净得（分）
	Period            string `json:"period" gorm:"type:varchar(7)"`                        // 所属月份 YYYY-MM
	Status            string `json:"status" gorm:"type:varchar(32);default:'pending'"`    // pending/settled/withdrawn
	CreatedAt         int64  `json:"created_at" gorm:"bigint"`
	SettledAt         int64  `json:"settled_at" gorm:"bigint;default:0"`
}

func (ProviderEarning) TableName() string {
	return "provider_earnings"
}

// CreateProviderEarning 创建供应商收益记录
func CreateProviderEarning(userId, channelId, consumerId, totalAmount, commissionAmount, netAmount int64, period string, status string) error {
	if status == "" {
		status = EarningStatusPending
	}
	return DB.Create(&ProviderEarning{
		UserId:           int(userId),
		ChannelId:        int(channelId),
		ConsumerId:       int(consumerId),
		TotalAmount:      totalAmount,
		CommissionAmount: commissionAmount,
		NetAmount:        netAmount,
		Period:           period,
		Status:           status,
		CreatedAt:        helper.GetTimestamp(),
	}).Error
}

// GetUserEarningsSum 获取供应商收益汇总（按状态）
func GetUserEarningsSum(userId int, status string) (int64, error) {
	var sum struct {
		Total int64
	}
	err := DB.Model(&ProviderEarning{}).
		Select("COALESCE(SUM(net_amount), 0) as total").
		Where("user_id = ? AND status = ?", userId, status).
		Scan(&sum).Error
	return sum.Total, err
}

// GetUserEarnings 获取供应商收益明细
func GetUserEarnings(userId int, limit int) ([]ProviderEarning, error) {
	var earnings []ProviderEarning
	err := DB.Where("user_id = ?", userId).Order("id desc").Limit(limit).Find(&earnings).Error
	return earnings, err
}

// GetUserEarningsByChannel 获取供应商各渠道收益汇总
type ChannelEarningSummary struct {
	ChannelId    int    `json:"channel_id"`
	ChannelName  string `json:"channel_name"`
	TotalAmount  int64  `json:"total_amount"`
	NetAmount    int64  `json:"net_amount"`
}

func GetUserEarningsByChannel(userId int) ([]ChannelEarningSummary, error) {
	var summaries []ChannelEarningSummary
	err := DB.Model(&ProviderEarning{}).
		Select("channel_id, '' as channel_name, SUM(total_amount) as total_amount, SUM(net_amount) as net_amount").
		Where("user_id = ? AND status = ?", userId, EarningStatusSettled).
		Group("channel_id").
		Scan(&summaries).Error
	if err != nil {
		return nil, err
	}
	// 填充渠道名称
	for i := range summaries {
		ch, err := GetChannelById(summaries[i].ChannelId, false)
		if err == nil {
			summaries[i].ChannelName = ch.Name
		}
	}
	return summaries, nil
}
