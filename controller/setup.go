package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/model"
)

type SetupCheckResponse struct {
	SetupNeeded bool `json:"setup_needed"`
	IsAdmin     bool `json:"is_admin"`
}

// CheckSetup checks if setup is needed
func CheckSetup(c *gin.Context) {
	count, err := model.CountUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to check users: " + err.Error(),
		})
		return
	}
	isAdmin := false
	if id, exists := c.Get("id"); exists {
		if idInt, ok := id.(int); ok {
			isAdmin = model.IsAdmin(idInt)
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"data": SetupCheckResponse{
			SetupNeeded: count == 0,
			IsAdmin:     isAdmin,
		},
	})
}

type SetupCompleteRequest struct {
	Username         string `json:"username" binding:"required"`
	Password         string `json:"password" binding:"required,min=8"`
	Email            string `json:"email" binding:"required,email"`
	SiteName         string `json:"site_name"`
	RegisterEnabled  *bool  `json:"register_enabled"`
	PasswordRegister *bool  `json:"password_register_enabled"`
}

// CompleteSetup creates the initial admin user
func CompleteSetup(c *gin.Context) {
	var req SetupCompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid request: " + err.Error(),
		})
		return
	}

	// Double-check no users exist yet
	count, err := model.CountUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to check users: " + err.Error(),
		})
		return
	}
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"message": "setup already completed",
		})
		return
	}

	// Create admin user
	user := &model.User{
		Username:   req.Username,
		Password:   req.Password,
		Email:      req.Email,
		Role:       model.RoleRootUser,
		Status:     model.UserStatusEnabled,
		Group:      "default",
	}

	if err := user.Insert(c.Request.Context(), 0); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to create admin user: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "setup completed successfully",
	})
}
