package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/model"
)

// ── Billing Audit API ──

// BillingRecord represents a single billing record for audit purposes.
type BillingRecord struct {
	ID          int     `json:"id"`
	UserId      int     `json:"user_id"`
	TokenId     int     `json:"token_id"`
	ModelName   string  `json:"model_name"`
	ChannelID   int     `json:"channel_id"`
	InputTokens int     `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
	BaseCost    float64 `json:"base_cost"`
	MarkupRate  float64 `json:"markup_rate"`
	FinalCost   float64 `json:"final_cost"`
	TierRate    float64 `json:"tier_rate"`
	CreatedAt   int64   `json:"created_at"`
}

// GetBillingAudit returns detailed billing records for audit (admin).
func GetBillingAudit(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	userId, _ := strconv.Atoi(c.DefaultQuery("user_id", "0"))
	modelName := c.Query("model")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 50
	}

	query := model.DB.Model(&model.Log{}).
		Order("created_at desc")

	if userId > 0 {
		query = query.Where("user_id = ?", userId)
	}
	if modelName != "" {
		query = query.Where("model_name LIKE ?", "%"+modelName+"%")
	}

	var total int64
	query.Count(&total)

	var logs []model.Log
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	records := make([]BillingRecord, 0, len(logs))
	for _, l := range logs {
		baseCost := float64(l.Quota) // quota is in micro-units
		records = append(records, BillingRecord{
			ID:           l.Id,
			UserId:       l.UserId,
			TokenId:      0,
			ModelName:    l.ModelName,
			ChannelID:    l.ChannelId,
			InputTokens:  l.PromptTokens,
			OutputTokens: l.CompletionTokens,
			BaseCost:     baseCost / 1000000.0,
			MarkupRate:   1.0,
			FinalCost:    baseCost / 1000000.0,
			TierRate:     1.0,
			CreatedAt:    l.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "",
		"data":      records,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ── Subscription Tier Info ──

// GetSubscriptionTierInfo returns tiered pricing info for a plan.
func GetSubscriptionTierInfo(c *gin.Context) {
	planId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid plan id"})
		return
	}

	var plan model.SubscriptionPlan
	if err := model.DB.First(&plan, planId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "plan not found"})
		return
	}

	var tiers []map[string]interface{}
	if plan.TiersJSON != "" {
		if err := json.Unmarshal([]byte(plan.TiersJSON), &tiers); err != nil {
			tiers = nil
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"plan_id":      plan.Id,
			"plan_title":   plan.Title,
			"has_tiers":    plan.TiersJSON != "",
			"tiers":        tiers,
			"base_price":   plan.PriceCents,
			"total_amount": plan.TotalAmount,
		},
	})
}

// ── Price Calculation Preview ──

// PreviewBilling shows how a request would be billed for preview purposes.
func PreviewBilling(c *gin.Context) {
	var req struct {
		ModelName    string `json:"model_name"`
		InputTokens  int    `json:"input_tokens"`
		OutputTokens int    `json:"output_tokens"`
		ChannelID    int    `json:"channel_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid request"})
		return
	}

	// Get base price from settlement config
	settleCfg, _ := model.GetSettlementConfig(req.ModelName)
	basePrice := settleCfg.UnifiedCost

	// Calculate token cost
	tokenCost := basePrice * float64(req.InputTokens+req.OutputTokens)

	// Apply channel markup
	markupRate := 1.0
	if req.ChannelID > 0 {
		var ch model.Channel
		if err := model.DB.First(&ch, req.ChannelID).Error; err == nil {
			markupRate = ch.ChannelMarkup
			if markupRate <= 0 {
				markupRate = 1.0
			}
		}
	}
	markedUpCost := tokenCost * markupRate

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"model":                req.ModelName,
			"input_tokens":         req.InputTokens,
			"output_tokens":        req.OutputTokens,
			"total_tokens":         req.InputTokens + req.OutputTokens,
			"base_price_per_token": basePrice,
			"base_cost":            tokenCost,
			"channel_markup_rate":  markupRate,
			"final_cost":           markedUpCost,
			"currency":             "USD",
			"timestamp":            time.Now().UnixMilli(),
		},
	})
}
