package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
)

func GetLatestPoolAgreement(c *gin.Context) {
	a, err := model.GetLatestAgreement()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "agreement not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": a})
}

func GetMyPoolConsent(c *gin.Context) {
	userID := c.GetInt("id")
	consent, err := model.GetUserPoolConsent(userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
			"agreed": false, "agreed_version": 0, "need_reconsent": true,
		}})
		return
	}
	latest, _ := model.GetLatestAgreement()
	needReconsent := latest != nil && consent.AgreedVersion < latest.Version
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
		"agreed":         consent.Agreed,
		"agreed_version": consent.AgreedVersion,
		"need_reconsent": needReconsent,
	}})
}

func SetMyPoolConsent(c *gin.Context) {
	userID := c.GetInt("id")
	var req struct {
		Agreed bool `json:"agreed"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid params"})
		return
	}
	if req.Agreed {
		if _, err := model.GetLatestAgreement(); err != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "no agreement published yet"})
			return
		}
	}
	if err := model.UpsertUserPoolConsent(userID, req.Agreed); err != nil {
		logger.Errorf(c.Request.Context(), "save pool consent: %v", err)
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "save failed"})
		return
	}
	msg := "agreed"
	if !req.Agreed {
		msg = "declined"
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": msg})
}

func AdminGetPoolAgreements(c *gin.Context) {
	list, err := model.GetAllAgreements()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	if list == nil {
		list = []model.PlatformPoolAgreement{}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": list})
}

func AdminPublishPoolAgreement(c *gin.Context) {
	var req struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "title and content required"})
		return
	}
	a, err := model.PublishNewAgreement(req.Title, req.Content)
	if err != nil {
		logger.Errorf(c.Request.Context(), "publish agreement: %v", err)
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "publish failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "published", "data": a})
}
