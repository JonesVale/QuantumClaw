package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/model"
)

type SetupCheckResponse struct {
	SetupNeeded bool `json:"setup_needed"`
	IsAdmin     bool `json:"is_admin"`
}

// CheckSetup checks if setup is needed.
// If only the auto-created root user exists (no email set), setup is still considered needed.
func CheckSetup(c *gin.Context) {
	count, err := model.CountUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to check users: " + err.Error(),
		})
		return
	}

	// If exactly 1 user exists (auto-created root with no email), setup is still needed
	setupNeeded := count == 0
	if count == 1 {
		if user, err := model.GetRootUser(); err == nil && user.Email == "" {
			setupNeeded = true
		}
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
			SetupNeeded: setupNeeded,
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

	// Double-check no real users exist yet (allow re-setup for auto-created root)
	count, err := model.CountUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to check users: " + err.Error(),
		})
		return
	}
	canSetup := count == 0
	if count == 1 {
		// Allow setup to override auto-created root user with no email
		if root, err := model.GetRootUser(); err == nil && root.Email == "" {
			canSetup = true
		}
	}
	if !canSetup {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"message": "setup already completed",
		})
		return
	}

	// Create or update admin user
	hashedPassword, err := common.Password2Hash(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to hash password",
		})
		return
	}

	user := &model.User{
		Username:   req.Username,
		Password:   req.Password,
		Email:      req.Email,
		Role:       model.RoleRootUser,
		Status:     model.UserStatusEnabled,
		Group:      "default",
	}

	// Check if auto-created root user exists — update instead of create
	var existingUser *model.User
	if root, err := model.GetRootUser(); err == nil && root.Email == "" {
		existingUser = root
	}

	if existingUser != nil {
		// Update the auto-created root user
		existingUser.Username = req.Username
		existingUser.Email = req.Email
		existingUser.Password = string(hashedPassword)
		if err := model.DB.Save(existingUser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "failed to update admin user: " + err.Error(),
			})
			return
		}
	} else {
		// Fresh setup — create new admin user
		user.Password = string(hashedPassword)
		if err := user.Insert(c.Request.Context(), 0); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "failed to create admin user: " + err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "setup completed successfully",
	})
}
