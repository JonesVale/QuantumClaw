package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/ctxkey"
	"github.com/quantumclaw/quantumclaw/model"
	relaymodel "github.com/quantumclaw/quantumclaw/relay/model"
)

func GetSubscription(c *gin.Context) {
	var remainQuota int64
	var usedQuota int64
	var err error
	var token *model.Token
	var expiredTime int64
	if config.DisplayTokenStatEnabled {
		tokenId := c.GetInt(ctxkey.TokenId)
		token, err = model.GetTokenById(tokenId)
		if err == nil {
			expiredTime = token.ExpiredTime
			remainQuota = token.RemainQuota
			usedQuota = token.UsedQuota
		}
	} else {
		userId := c.GetInt(ctxkey.Id)
		remainQuota, err = model.GetUserQuota(userId)
		if err != nil {
			usedQuota, err = model.GetUserUsedQuota(userId)
		}
	}
	if expiredTime <= 0 {
		expiredTime = 0
	}
	if err != nil {
		Error := relaymodel.Error{
			Message: err.Error(),
			Type:    "upstream_error",
		}
		c.JSON(200, gin.H{
			"error": Error,
		})
		return
	}
	quota := remainQuota + usedQuota
	amount := float64(quota)
	if config.DisplayInCurrencyEnabled {
		amount /= config.QuotaPerUnit
	}
	if token != nil && token.UnlimitedQuota {
		amount = 100000000
	}
	subscription := OpenAISubscriptionResponse{
		Object:             "billing_subscription",
		HasPaymentMethod:   true,
		SoftLimitUSD:       amount,
		HardLimitUSD:       amount,
		SystemHardLimitUSD: amount,
		AccessUntil:        expiredTime,
	}
	c.JSON(200, subscription)
	return
}

// GetBillingStats 获取用户消费统计数据
func GetBillingStats(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	usedQuota, _ := model.GetUserUsedQuota(userId)
	remainQuota, _ := model.GetUserQuota(userId)
	var logCount int64
	model.DB.Model(&model.Log{}).Where("user_id = ?", userId).Count(&logCount)

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"total_quota":     remainQuota + usedQuota,
			"used_quota":      usedQuota,
			"remain_quota":    remainQuota,
			"request_count":   logCount,
			"display_in_currency": config.DisplayInCurrencyEnabled,
			"quota_per_unit":  config.QuotaPerUnit,
		},
	})
}

// GetBillingRecords 获取用户消费记录
func GetBillingRecords(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	records, _, err := model.GetUserTransactionLogs(userId, 1, 50)
	if err != nil || records == nil {
		records = []model.TransactionLog{}
	}
	data := make([]gin.H, 0, len(records))
	for _, r := range records {
		data = append(data, gin.H{
			"id":           r.Id,
			"amount":       r.Amount,
			"action":       r.Action,
			"before_quota": r.BeforeQuota,
			"after_quota":  r.AfterQuota,
			"status":       r.Status,
			"remark":       r.Remark,
			"created_at":   r.CreatedAt,
		})
	}
	c.JSON(200, gin.H{"success": true, "data": data})
}

func GetUsage(c *gin.Context) {
	var quota int64
	var err error
	var token *model.Token
	if config.DisplayTokenStatEnabled {
		tokenId := c.GetInt(ctxkey.TokenId)
		token, err = model.GetTokenById(tokenId)
		quota = token.UsedQuota
	} else {
		userId := c.GetInt(ctxkey.Id)
		quota, err = model.GetUserUsedQuota(userId)
	}
	if err != nil {
		Error := relaymodel.Error{
			Message: err.Error(),
			Type:    "quantumclaw_error",
		}
		c.JSON(200, gin.H{
			"error": Error,
		})
		return
	}
	amount := float64(quota)
	if config.DisplayInCurrencyEnabled {
		amount /= config.QuotaPerUnit
	}
	usage := OpenAIUsageResponse{
		Object:     "list",
		TotalUsage: amount * 100,
	}
	c.JSON(200, usage)
	return
}
