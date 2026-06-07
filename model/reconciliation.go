package model

import (
	"strconv"

	"github.com/quantumclaw/quantumclaw/common/helper"
)

const (
	ReconciliationStatusOpen     = "open"      // 待处理（不一致）
	ReconciliationStatusResolved = "resolved"  // 已处理
	ReconciliationStatusIgnored  = "ignored"   // 已忽略
)

// ReconciliationLog 对账记录表
// 每次计费完成后异步触发对账，记录用户扣款、渠道成本、平台收入是否一致
type ReconciliationLog struct {
	Id                  int    `json:"id"`
	UserId              int    `json:"user_id" gorm:"type:int;index"`
	Username            string `json:"username" gorm:"-"` // 不映射数据库列，仅用于 JSON 序列化
	ChannelId           int    `json:"channel_id" gorm:"type:int;index"`
	ChannelName         string `json:"channel_name" gorm:"-"` // 不映射数据库列，仅用于 JSON 序列化
	ConsumeLogId       int    `json:"consume_log_id" gorm:"type:int;default:0"` // 关联的 consume_log 主键
	UserDeductedCents int64  `json:"user_deducted_cents" gorm:"bigint"`     // 用户实际扣款（分）
	ChannelCostCents   int64  `json:"channel_cost_cents" gorm:"bigint"`      // 渠道上游成本估算（分）
	PlatformIncomeCents int64  `json:"platform_income_cents" gorm:"bigint"` // 平台收入（分）
	Status              string `json:"status" gorm:"type:varchar(32);index"` // matched / discrepancy / pending
	DiffCents          int64  `json:"diff_cents" gorm:"bigint"`           // 差额（分），正=多收，负=少收
	Remark              string `json:"remark" gorm:"type:varchar(512)"`
	CreatedAt           int64  `json:"created_at" gorm:"bigint"`
	UpdatedAt           int64  `json:"updated_at" gorm:"bigint"`
}

func (ReconciliationLog) TableName() string {
	return "reconciliation_logs"
}

// CreateReconciliationLog 写入一条对账记录
func CreateReconciliationLog(log *ReconciliationLog) error {
	log.CreatedAt = helper.GetTimestamp()
	log.UpdatedAt = log.CreatedAt
	return DB.Create(log).Error
}

// UpdateReconciliationStatus 更新对账状态
func UpdateReconciliationStatus(id int, status string, diffCents int64, remark string) error {
	return DB.Model(&ReconciliationLog{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"diff_cents": diffCents,
			"remark":     remark,
			"updated_at": helper.GetTimestamp(),
		}).Error
}

// ListReconciliationLogs 查询对账记录（供后台 API 使用）
// keyword 支持按 user_id 或 channel_id 模糊搜索
func ListReconciliationLogs(userId int, status string, keyword string, page int, pageSize int) ([]ReconciliationLog, int64, error) {
	var logs []ReconciliationLog
	var total int64

	query := DB.Model(&ReconciliationLog{})
	if userId > 0 {
		query = query.Where("user_id = ?", userId)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		// 支持按 user_id 或 channel_id 搜索（精确匹配数字 ID）
		kwInt, err := strconv.Atoi(keyword)
		if err == nil && kwInt > 0 {
			query = query.Where("user_id = ? OR channel_id = ?", kwInt, kwInt)
		} else {
			// 非数字关键字不返回结果（避免全表扫描）
			query = query.Where("1 = 0")
		}
	}

	query.Count(&total)
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&logs).Error
	if err != nil {
		return logs, total, err
	}

	// 批量填充 Username 和 ChannelName
	fillReconciliationNames(&logs)

	return logs, total, nil
}

// ListUserReconciliationLogs 查询指定用户的对账记录（供普通用户查看自己的对账明细）
func ListUserReconciliationLogs(userId int, status string, page int, pageSize int) ([]ReconciliationLog, int64, error) {
	var logs []ReconciliationLog
	var total int64

	if userId <= 0 {
		return logs, 0, nil
	}

	query := DB.Model(&ReconciliationLog{}).Where("user_id = ?", userId)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&logs).Error
	if err != nil {
		return logs, total, err
	}

	// 批量填充 Username 和 ChannelName
	fillReconciliationNames(&logs)

	return logs, total, nil
}

func fillReconciliationNames(logs *[]ReconciliationLog) {
	if len(*logs) == 0 {
		return
	}

	// 收集所有不重复的 user_id 和 channel_id
	userIds := make(map[int]bool)
	channelIds := make(map[int]bool)
	for _, log := range *logs {
		if log.UserId > 0 {
			userIds[log.UserId] = true
		}
		if log.ChannelId > 0 {
			channelIds[log.ChannelId] = true
		}
	}

	// 批量查询用户名
	if len(userIds) > 0 {
		var users []User
		uidList := make([]int, 0, len(userIds))
		for id := range userIds {
			uidList = append(uidList, id)
		}
		DB.Select("id, username").Where("id IN ?", uidList).Find(&users)
		userMap := make(map[int]string)
		for _, u := range users {
			userMap[u.Id] = u.Username
		}
		for i := range *logs {
			if name, ok := userMap[(*logs)[i].UserId]; ok {
				(*logs)[i].Username = name
			}
		}
	}

	// 批量查询渠道名
	if len(channelIds) > 0 {
		var channels []Channel
		cidList := make([]int, 0, len(channelIds))
		for id := range channelIds {
			cidList = append(cidList, id)
		}
		DB.Select("id, name").Where("id IN ?", cidList).Find(&channels)
		channelMap := make(map[int]string)
		for _, c := range channels {
			channelMap[c.Id] = c.Name
		}
		for i := range *logs {
			if name, ok := channelMap[(*logs)[i].ChannelId]; ok {
				(*logs)[i].ChannelName = name
			}
		}
	}
}
