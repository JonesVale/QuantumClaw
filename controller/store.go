package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/helper"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/common/ctxkey"
	"github.com/quantumclaw/quantumclaw/model"
)

// RegisterStore - open a new store
func RegisterStore(c *gin.Context) {
	userID := c.GetInt(ctxkey.Id)

	existing, _ := model.GetStoreByUserID(userID)
	if existing != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "store already exists"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "store name required"})
		return
	}

	store, err := model.CreateStore(userID, req.Name)
	if err != nil {
		logger.Errorf(c.Request.Context(), "create store failed: %v", err)
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "create store failed"})
		return
	}

	// auto-create listings from existing channels
	channels, _ := model.GetAllChannels(0, 0, "all")
	autoListings := 0
	for _, ch := range channels {
		if ch.UserId == userID && ch.Key != "" && !strings.HasPrefix(ch.Key, "PUT_YOUR") {
			if ch.StoreID == 0 {
				model.DB.Model(ch).Update("store_id", store.ID)
			}
			for _, modelName := range splitModels(ch.Models) {
				listing := &model.Listing{
					ID:           fmt.Sprintf("lst_%d_%s", helper.GetTimestamp(), generateRandomString(8)),
					StoreID:      store.ID,
					ProviderID:   userID,
					ChannelID:    ch.Id,
					ModelName:    modelName,
					Region:       ch.Region,
					PricePerUnit: int64(ch.CostPrice * float64(ch.ChannelMarkup) * 100),
					Status:       model.ListingStatusActive,
				}
				if listing.PricePerUnit <= 0 {
					listing.PricePerUnit = 100
				}
				if err := model.CreateListing(listing); err != nil {
					logger.Warnf(c.Request.Context(), "auto listing failed: %v", err)
				} else {
					autoListings++
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"store":         store,
			"auto_listings": autoListings,
		},
	})
}

// GetStoreProfile - get my store info
func GetStoreProfile(c *gin.Context) {
	userID := c.GetInt(ctxkey.Id)
	store, err := model.GetStoreByUserID(userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "no store"})
		return
	}
	feeCfg, _ := model.GetFeeConfig(store.Tier)
	rate := 10.0
	if feeCfg != nil {
		rate = feeCfg.Rate
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"store":       store,
			"fee_rate":    rate,
			"fee_rate_pct": fmt.Sprintf("%.0f%%", rate),
		},
	})
}

// UpdateStoreProfile - update store info
func UpdateStoreProfile(c *gin.Context) {
	userID := c.GetInt(ctxkey.Id)
	store, err := model.GetStoreByUserID(userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "no store"})
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid params"})
		return
	}
	if req.Name != "" {
		store.Name = req.Name
	}
	if err := model.UpdateStoreInfo(store); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "update failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "updated"})
}

// CreateListing - list a resource for sale
func CreateListing(c *gin.Context) {
	userID := c.GetInt(ctxkey.Id)
	store, err := model.GetStoreByUserID(userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "no store"})
		return
	}

	var req struct {
		ChannelID    int    `json:"channel_id" binding:"required"`
		ModelName    string `json:"model_name" binding:"required"`
		PricePerUnit int64  `json:"price_per_unit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid params"})
		return
	}

	channel, err := model.GetChannelById(req.ChannelID, true)
	if err != nil || channel.UserId != userID {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "channel not found or not yours"})
		return
	}
	if channel.StoreID == 0 {
		model.DB.Model(channel).Update("store_id", store.ID)
	}

	if req.PricePerUnit <= 0 {
		req.PricePerUnit = int64(channel.CostPrice * float64(channel.ChannelMarkup) * 100)
	}
	if req.PricePerUnit <= 0 {
		req.PricePerUnit = 100
	}

	listing := &model.Listing{
		ID:           fmt.Sprintf("lst_%d_%s", helper.GetTimestamp(), generateRandomString(8)),
		StoreID:      store.ID,
		ProviderID:   userID,
		ChannelID:    req.ChannelID,
		ModelName:    req.ModelName,
		PricePerUnit: req.PricePerUnit,
		Region:       channel.Region,
		Status:       model.ListingStatusActive,
	}
	if err := model.CreateListing(listing); err != nil {
		logger.Errorf(c.Request.Context(), "create listing failed: %v", err)
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "list failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": listing})
}

// GetMyListings - my listed resources
func GetMyListings(c *gin.Context) {
	userID := c.GetInt(ctxkey.Id)
	store, err := model.GetStoreByUserID(userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "data": []model.Listing{}})
		return
	}
	listings, err := model.GetListingsByStoreID(store.ID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "data": []model.Listing{}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": listings})
}

// UpdateListingPrice - change price
func UpdateListingPrice(c *gin.Context) {
	userID := c.GetInt(ctxkey.Id)
	id := c.Param("id")

	listing, err := model.GetListingByID(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "listing not found"})
		return
	}
	store, _ := model.GetStoreByUserID(userID)
	if store == nil || listing.StoreID != store.ID {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "not your listing"})
		return
	}

	var req struct {
		PricePerUnit int64 `json:"price_per_unit" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.PricePerUnit <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "price must be > 0"})
		return
	}
	listing.PricePerUnit = req.PricePerUnit
	if err := model.UpdateListing(listing); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "update failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "price updated"})
}

// UpdateListingStatus - pause/resume/archive
func UpdateListingStatus(c *gin.Context) {
	userID := c.GetInt(ctxkey.Id)
	id := c.Param("id")

	listing, err := model.GetListingByID(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "listing not found"})
		return
	}
	store, _ := model.GetStoreByUserID(userID)
	if store == nil || listing.StoreID != store.ID {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "not your listing"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid params"})
		return
	}
	switch req.Status {
	case "active", "paused", "archived":
		listing.Status = model.ListingStatus(req.Status)
	default:
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid status"})
		return
	}
	if err := model.UpdateListing(listing); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "update failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "status updated"})
}

// GetStoreEarnings - earnings dashboard
func GetStoreEarnings(c *gin.Context) {
	userID := c.GetInt(ctxkey.Id)
	store, err := model.GetStoreByUserID(userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "no store"})
		return
	}

	var totalEarned int64
	model.DB.Model(&model.ProviderEarning{}).Where("user_id = ? AND status = ?", userID, model.EarningStatusSettled).
		Select("COALESCE(SUM(net_amount), 0)").Scan(&totalEarned)

	var totalWithdrawn int64
	model.DB.Model(&model.WithdrawalRequest{}).Where("user_id = ? AND status IN ?",
		userID, []string{model.WithdrawStatusApproved, model.WithdrawStatusCompleted}).
		Select("COALESCE(SUM(net_amount), 0)").Scan(&totalWithdrawn)

	var pendingFee int64
	model.DB.Model(&model.PlatformFeeRecord{}).Where("store_id = ? AND status = ?",
		store.ID, model.PlatformFeeStatusPending).
		Select("COALESCE(SUM(fee_amount), 0)").Scan(&pendingFee)

	available := totalEarned - totalWithdrawn - pendingFee

	var periodStats []struct {
		Period string `json:"period"`
		Amount int64  `json:"amount"`
	}
	model.DB.Model(&model.ProviderEarning{}).Where("user_id = ? AND status = ?", userID, model.EarningStatusSettled).
		Select("strftime('%Y-%m', created_at, 'unixepoch') as period, SUM(net_amount) as amount").
		Group("period").Order("period DESC").Limit(12).Scan(&periodStats)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"store":           store,
			"total_earned":    totalEarned,
			"total_withdrawn": totalWithdrawn,
			"pending_fee":     pendingFee,
			"available":       available,
			"period_stats":    periodStats,
		},
	})
}

// GetStoreFees - platform fee records
func GetStoreFees(c *gin.Context) {
	userID := c.GetInt(ctxkey.Id)
	store, err := model.GetStoreByUserID(userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "data": []model.PlatformFeeRecord{}})
		return
	}
	fees, err := model.GetPendingPlatformFees(store.UserID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": fees})
}

func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	now := helper.GetTimestamp()
	for i := range b {
		b[i] = letters[now%int64(len(letters))]
		now++
	}
	return string(b)
}
