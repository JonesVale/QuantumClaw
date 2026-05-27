package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/service"
	"github.com/gin-gonic/gin"
)

// GetSub2Providers lists all supported subscription services.
func GetSub2Providers(c *gin.Context) {
	providers := service.Sub2API.GetSupportedProviders()
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": providers})
}

type AddSub2CredentialRequest struct {
	Provider string `json:"provider"`
	Token    string `json:"token"`
	Label    string `json:"label"`
	DailyCap int64  `json:"daily_cap"`
}

// AddSub2Credential adds a new subscription credential for the current user.
func AddSub2Credential(c *gin.Context) {
	userId := c.GetInt("id")
	var req AddSub2CredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid request"})
		return
	}
	cred, err := service.Sub2API.CreateCredential(userId, model.Sub2APIProvider(req.Provider), req.Label, req.Token, req.DailyCap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "credential added", "data": cred})
}

// ListSub2Credentials lists the current user's subscription credentials.
func ListSub2Credentials(c *gin.Context) {
	userId := c.GetInt("id")
	creds, err := service.Sub2API.GetUserCredentials(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": creds})
}

// DeleteSub2Credential deletes a credential for the current user.
func DeleteSub2Credential(c *gin.Context) {
	userId := c.GetInt("id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid id"})
		return
	}
	if err := model.DeleteSub2Credential(id, userId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "credential deleted"})
}

// TestSub2Credential validates a credential's connectivity (placeholder - actual test needs adaptor).
func TestSub2Credential(c *gin.Context) {
	userId := c.GetInt("id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid id"})
		return
	}
	cred, err := service.Sub2API.ValidateCredential(id, userId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"id":            cred.Id,
			"provider":      cred.Provider,
			"label":         cred.Label,
			"status":        cred.Status,
			"daily_cap":     cred.DailyCap,
			"used_today":    cred.UsedToday,
			"last_health":   time.UnixMilli(cred.LastHealthAt).Format("2006-01-02 15:04:05"),
		},
	})
}

// UpdateSub2Credential updates credential settings (label, daily_cap, etc.).
type UpdateSub2CredentialRequest struct {
	Label    string `json:"label"`
	DailyCap int64  `json:"daily_cap"`
	Status   int    `json:"status"` // 1=active, 2=paused
}

func UpdateSub2Credential(c *gin.Context) {
	userId := c.GetInt("id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid id"})
		return
	}
	var req UpdateSub2CredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid request"})
		return
	}
	cred, err := model.GetSub2Credential(id, userId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "credential not found"})
		return
	}
	if req.Label != "" {
		cred.Label = req.Label
	}
	if req.DailyCap > 0 {
		cred.DailyCap = req.DailyCap
	}
	if req.Status >= 1 && req.Status <= 2 {
		cred.Status = req.Status
	}
	cred.UpdatedTime = time.Now().UnixMilli()
	if err := model.UpdateSub2Credential(cred); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	cred.Token = ""
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "credential updated", "data": cred})
}

// ── Admin ──

// AdminListAllSub2Credentials lists all credentials across all users (admin).
func AdminListAllSub2Credentials(c *gin.Context) {
	creds, err := model.ListAllSub2Credentials()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": creds})
}

// AdminDeleteSub2Credential deletes any credential (admin).
func AdminDeleteSub2Credential(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid id"})
		return
	}
	if err := model.DB.Delete(&model.Sub2APICredential{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "credential deleted"})
}
