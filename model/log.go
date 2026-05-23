package model

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/helper"
	"github.com/quantumclaw/quantumclaw/common/logger"
)

type Log struct {
	Id                int    `json:"id"`
	UserId            int    `json:"user_id" gorm:"index"`
	CreatedAt         int64  `json:"created_at" gorm:"bigint;index:idx_created_at_type"`
	Type              int    `json:"type" gorm:"index:idx_created_at_type"`
	Content           string `json:"content"`
	Username          string `json:"username" gorm:"index:index_username_model_name,priority:2;default:''"`
	TokenName         string `json:"token_name" gorm:"index;default:''"`
	ModelName         string `json:"model_name" gorm:"index;index:index_username_model_name,priority:1;default:''"`
	Quota             int    `json:"quota" gorm:"default:0"`
	PromptTokens      int    `json:"prompt_tokens" gorm:"default:0"`
	CompletionTokens  int    `json:"completion_tokens" gorm:"default:0"`
	ChannelId         int    `json:"channel" gorm:"index"`
	RequestId         string `json:"request_id" gorm:"default:''"`
	ElapsedTime       int64  `json:"elapsed_time" gorm:"default:0"` // unit is ms
	IsStream          bool   `json:"is_stream" gorm:"default:false"`
	SystemPromptReset bool   `json:"system_prompt_reset" gorm:"default:false"`

	// Settlement system fields
	PromoterId     int  `json:"promoter_id" gorm:"default:0;index"`
	ChannelOwnerId int  `json:"channel_owner_id" gorm:"default:0"`
	IsFallback     bool `json:"is_fallback" gorm:"default:false"`
}

const (
	LogTypeUnknown = iota
	LogTypeTopup
	LogTypeConsume
	LogTypeManage
	LogTypeSystem
	LogTypeTest
)

func recordLogHelper(ctx context.Context, log *Log) {
	requestId := helper.GetRequestID(ctx)
	log.RequestId = requestId
	err := LOG_DB.Create(log).Error
	if err != nil {
		logger.Error(ctx, "failed to record log: "+err.Error())
		return
	}
	logger.Infof(ctx, "record log: %+v", log)
}

func RecordLog(ctx context.Context, userId int, logType int, content string) {
	if logType == LogTypeConsume && !config.LogConsumeEnabled {
		return
	}
	log := &Log{
		UserId:    userId,
		Username:  GetUsernameById(userId),
		CreatedAt: helper.GetTimestamp(),
		Type:      logType,
		Content:   content,
	}
	recordLogHelper(ctx, log)
}

func RecordTopupLog(ctx context.Context, userId int, content string, quota int) {
	log := &Log{
		UserId:    userId,
		Username:  GetUsernameById(userId),
		CreatedAt: helper.GetTimestamp(),
		Type:      LogTypeTopup,
		Content:   content,
		Quota:     quota,
	}
	recordLogHelper(ctx, log)
}

func RecordConsumeLog(ctx context.Context, log *Log) {
	if !config.LogConsumeEnabled {
		return
	}
	log.Username = GetUsernameById(log.UserId)
	log.CreatedAt = helper.GetTimestamp()
	log.Type = LogTypeConsume
	recordLogHelper(ctx, log)

	// 创建结算系统交易流水（含精确 token 数）
	createTransactionFromLog(log)
}

func createTransactionFromLog(log *Log) {
	if log.PromptTokens == 0 && log.CompletionTokens == 0 {
		return // 无实际 token 消耗，跳过
	}

	totalTokens := log.PromptTokens + log.CompletionTokens
	tokenK := float64(totalTokens) / 1000.0

	// 获取结算配置
	cfg, _ := GetSettlementConfig(log.ModelName)

	// 获取 channel 信息以计算单价
	var ch Channel
	if err := DB.First(&ch, "id = ?", log.ChannelId).Error; err != nil {
		// channel 不存在或已删除，用默认价格
		ch.CostPerUnit = 0.001
		ch.SellPriceRate = 1.0
	}

	unitPrice := ch.CostPerUnit * ch.SellPriceRate * 1000 // /1K tokens

	tx := &TokenTransaction{
		LogId:            log.Id,
		UserId:           log.UserId,
		ModelName:        log.ModelName,
		PromptTokens:     log.PromptTokens,
		CompletionTokens: log.CompletionTokens,
		ChannelId:        log.ChannelId,
		ChannelOwnerId:   log.ChannelOwnerId,
		PromoterId:       log.PromoterId,
		IsFallback:       0,
		UnitPrice:        unitPrice,
		TotalAmount:      unitPrice * tokenK,
		UnifiedCost:      cfg.UnifiedCost * tokenK,
		CommissionAmount: (unitPrice * tokenK) * cfg.CommissionRate,
		PlatformFee:      (unitPrice * tokenK) * cfg.PlatformFeeRate,
	}
	if log.IsFallback {
		tx.IsFallback = 1
	}

	// 查询 channel 的 cost_price
	if ch.CostPrice > 0 {
		tx.KeyProviderCost = ch.CostPrice * tokenK
	}

	if err := CreateTransaction(tx); err != nil {
		logger.SysError(fmt.Sprintf("failed to create token_transaction from log %d: %v", log.Id, err))
	}
}

func RecordTestLog(ctx context.Context, log *Log) {
	log.CreatedAt = helper.GetTimestamp()
	log.Type = LogTypeTest
	recordLogHelper(ctx, log)
}

func GetAllLogs(logType int, startTimestamp int64, endTimestamp int64, modelName string, username string, tokenName string, startIdx int, num int, channel int) (logs []*Log, err error) {
	var tx *gorm.DB
	if logType == LogTypeUnknown {
		tx = LOG_DB
	} else {
		tx = LOG_DB.Where("type = ?", logType)
	}
	if modelName != "" {
		tx = tx.Where("model_name = ?", modelName)
	}
	if username != "" {
		tx = tx.Where("username = ?", username)
	}
	if tokenName != "" {
		tx = tx.Where("token_name = ?", tokenName)
	}
	if startTimestamp != 0 {
		tx = tx.Where("created_at >= ?", startTimestamp)
	}
	if endTimestamp != 0 {
		tx = tx.Where("created_at <= ?", endTimestamp)
	}
	if channel != 0 {
		tx = tx.Where("channel_id = ?", channel)
	}
	err = tx.Order("id desc").Limit(num).Offset(startIdx).Find(&logs).Error
	return logs, err
}

func GetUserLogs(userId int, logType int, startTimestamp int64, endTimestamp int64, modelName string, tokenName string, startIdx int, num int) (logs []*Log, err error) {
	var tx *gorm.DB
	if logType == LogTypeUnknown {
		tx = LOG_DB.Where("user_id = ?", userId)
	} else {
		tx = LOG_DB.Where("user_id = ? and type = ?", userId, logType)
	}
	if modelName != "" {
		tx = tx.Where("model_name = ?", modelName)
	}
	if tokenName != "" {
		tx = tx.Where("token_name = ?", tokenName)
	}
	if startTimestamp != 0 {
		tx = tx.Where("created_at >= ?", startTimestamp)
	}
	if endTimestamp != 0 {
		tx = tx.Where("created_at <= ?", endTimestamp)
	}
	err = tx.Order("id desc").Limit(num).Offset(startIdx).Omit("id").Find(&logs).Error
	return logs, err
}

func SearchAllLogs(keyword string) (logs []*Log, err error) {
	err = LOG_DB.Where("type = ? or content LIKE ?", keyword, keyword+"%").Order("id desc").Limit(config.MaxRecentItems).Find(&logs).Error
	return logs, err
}

func SearchUserLogs(userId int, keyword string) (logs []*Log, err error) {
	err = LOG_DB.Where("user_id = ? and type = ?", userId, keyword).Order("id desc").Limit(config.MaxRecentItems).Omit("id").Find(&logs).Error
	return logs, err
}

func SumUsedQuota(logType int, startTimestamp int64, endTimestamp int64, modelName string, username string, tokenName string, channel int) (quota int64) {
	ifnull := "ifnull"
	if common.UsingPostgreSQL {
		ifnull = "COALESCE"
	}
	tx := LOG_DB.Table("logs").Select(fmt.Sprintf("%s(sum(quota),0)", ifnull))
	if username != "" {
		tx = tx.Where("username = ?", username)
	}
	if tokenName != "" {
		tx = tx.Where("token_name = ?", tokenName)
	}
	if startTimestamp != 0 {
		tx = tx.Where("created_at >= ?", startTimestamp)
	}
	if endTimestamp != 0 {
		tx = tx.Where("created_at <= ?", endTimestamp)
	}
	if modelName != "" {
		tx = tx.Where("model_name = ?", modelName)
	}
	if channel != 0 {
		tx = tx.Where("channel_id = ?", channel)
	}
	tx.Where("type = ?", LogTypeConsume).Scan(&quota)
	return quota
}

func SumUsedToken(logType int, startTimestamp int64, endTimestamp int64, modelName string, username string, tokenName string) (token int) {
	ifnull := "ifnull"
	if common.UsingPostgreSQL {
		ifnull = "COALESCE"
	}
	tx := LOG_DB.Table("logs").Select(fmt.Sprintf("%s(sum(prompt_tokens),0) + %s(sum(completion_tokens),0)", ifnull, ifnull))
	if username != "" {
		tx = tx.Where("username = ?", username)
	}
	if tokenName != "" {
		tx = tx.Where("token_name = ?", tokenName)
	}
	if startTimestamp != 0 {
		tx = tx.Where("created_at >= ?", startTimestamp)
	}
	if endTimestamp != 0 {
		tx = tx.Where("created_at <= ?", endTimestamp)
	}
	if modelName != "" {
		tx = tx.Where("model_name = ?", modelName)
	}
	tx.Where("type = ?", LogTypeConsume).Scan(&token)
	return token
}

// ModelRankingItem — 模型排行榜原始统计数据
// 用于后续合并渠道定价和趋势计算
func GetModelRankings(startTs, endTs int64) ([]LogStatistic, error) {
	groupSelect := "DATE_FORMAT(FROM_UNIXTIME(created_at), '%Y-%m-%d') as day"
	if common.UsingPostgreSQL {
		groupSelect = "TO_CHAR(date_trunc('day', to_timestamp(created_at)), 'YYYY-MM-DD') as day"
	}
	if common.UsingSQLite {
		groupSelect = "strftime('%Y-%m-%d', datetime(created_at, 'unixepoch')) as day"
	}

	var stats []LogStatistic
	// 按 model_name+channel_id 汇总（不保留 day 维度，但复用 LogStatistic 结构体）
	err := LOG_DB.Raw(`
		SELECT model_name,
		count(1) as request_count,
		sum(quota) as quota,
		sum(prompt_tokens) as prompt_tokens,
		sum(completion_tokens) as completion_tokens,
		`+groupSelect+` as day
		FROM logs
		WHERE type=2
		AND created_at BETWEEN ? AND ?
		GROUP BY model_name
		ORDER BY request_count DESC
	`, startTs, endTs).Scan(&stats).Error
	return stats, err
}

// ModelChannelStat — 每个 model+channel 组合的统计
// 用于排行榜和仪表盘
func GetModelChannelStats(startTs, endTs int64) ([]*ModelChannelStat, error) {
	var results []*ModelChannelStat
	err := LOG_DB.Raw(`
		SELECT model_name,
		channel_id,
		count(1) as request_count,
		coalesce(sum(prompt_tokens + completion_tokens),0) as total_tokens,
		coalesce(avg(elapsed_time),0) as avg_speed_ms
		FROM logs
		WHERE type=2
		AND created_at BETWEEN ? AND ?
		GROUP BY model_name, channel_id
		ORDER BY request_count DESC
	`, startTs, endTs).Scan(&results).Error
	return results, err
}

func DeleteOldLog(targetTimestamp int64) (int64, error) {
	result := LOG_DB.Where("created_at < ?", targetTimestamp).Delete(&Log{})
	return result.RowsAffected, result.Error
}

type ModelChannelStat struct {
	ModelName    string  `gorm:"column:model_name"`
	ChannelId    int     `gorm:"column:channel_id"`
	RequestCount int64   `gorm:"column:request_count"`
	TotalTokens  int64   `gorm:"column:total_tokens"`
	AvgSpeedMs   float64 `gorm:"column:avg_speed_ms"`
}

type LogStatistic struct {
	Day              string `gorm:"column:day" json:"day"`
	ModelName        string `gorm:"column:model_name" json:"model_name"`
	RequestCount     int    `gorm:"column:request_count" json:"request_count"`
	Quota            int    `gorm:"column:quota" json:"quota"`
	PromptTokens     int    `gorm:"column:prompt_tokens" json:"prompt_tokens"`
	CompletionTokens int    `gorm:"column:completion_tokens" json:"completion_tokens"`
}

func SearchLogsByDayAndModel(userId, start, end int) (LogStatistics []*LogStatistic, err error) {
	groupSelect := "DATE_FORMAT(FROM_UNIXTIME(created_at), '%Y-%m-%d') as day"

	if common.UsingPostgreSQL {
		groupSelect = "TO_CHAR(date_trunc('day', to_timestamp(created_at)), 'YYYY-MM-DD') as day"
	}

	if common.UsingSQLite {
		groupSelect = "strftime('%Y-%m-%d', datetime(created_at, 'unixepoch')) as day"
	}

	err = LOG_DB.Raw(`
		SELECT `+groupSelect+`,
		model_name, count(1) as request_count,
		sum(quota) as quota,
		sum(prompt_tokens) as prompt_tokens,
		sum(completion_tokens) as completion_tokens
		FROM logs
		WHERE type=2
		AND user_id= ?
		AND created_at BETWEEN ? AND ?
		GROUP BY day, model_name
		ORDER BY day, model_name
	`, userId, start, end).Scan(&LogStatistics).Error

	return LogStatistics, err
}
