package service

import (
	"fmt"

	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
	"gorm.io/gorm"
)

func RecordConsumeLog(userId int, channelId int, promptTokens int, completionTokens int, modelName string, tokenName string, quota int, content string) {
	model.RecordConsumeLog(nil, &model.Log{
		UserId:           userId,
		ChannelId:        channelId,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		ModelName:        modelName,
		TokenName:        tokenName,
		Quota:            quota,
		Content:          content,
	})
}

func UpdateUserUsedQuotaAndRequestCount(userId int, quota int64) {
	err := model.DB.Model(&model.User{}).Where("id = ?", userId).
		Updates(map[string]interface{}{
			"used_quota":    gorm.Expr("used_quota + ?", quota),
			"request_count": gorm.Expr("request_count + 1"),
		}).Error
	if err != nil {
		logger.SysError("failed to update user used quota: " + err.Error())
	}
}

func UpdateChannelUsedQuota(channelId int, quota int64) {
	err := model.DB.Model(&model.Channel{}).Where("id = ?", channelId).
		Update("used_quota", gorm.Expr("used_quota + ?", quota)).Error
	if err != nil {
		logger.SysError(fmt.Sprintf("failed to update channel used quota for channel %d: %s", channelId, err.Error()))
	}
}
