package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/model"
)

// GetPromoAds returns enabled promo ads, optionally filtered by page_key.
// Public endpoint (no auth required).
func GetPromoAds(c *gin.Context) {
	pageKey := c.Query("page_key")
	ads, err := model.GetEnabledPromoAds(pageKey)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": ads})
}

// AdminGetAllPromoAds returns all promo ads (including disabled) for admin management.
func AdminGetAllPromoAds(c *gin.Context) {
	ads, err := model.GetAllPromoAds()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": ads})
}

// AdminCreatePromoAd creates a new promo ad.
func AdminCreatePromoAd(c *gin.Context) {
	var ad model.PromoAd
	if err := c.ShouldBindJSON(&ad); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid request: " + err.Error()})
		return
	}
	if err := model.CreatePromoAd(&ad); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": ad})
}

// AdminUpdatePromoAd updates an existing promo ad.
func AdminUpdatePromoAd(c *gin.Context) {
	var ad model.PromoAd
	if err := c.ShouldBindJSON(&ad); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid request: " + err.Error()})
		return
	}
	if ad.ID == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "id is required"})
		return
	}
	if err := model.UpdatePromoAd(&ad); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": ad})
}

// AdminDeletePromoAd deletes a promo ad by ID.
func AdminDeletePromoAd(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid id"})
		return
	}
	if err := model.DeletePromoAd(uint(id)); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "deleted"})
}
