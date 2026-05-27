package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/i18n"
	"github.com/quantumclaw/quantumclaw/model"
)

// ListApps 公开应用列表 — GET /api/apps
func ListApps(c *gin.Context) {
	ctx := c.Request.Context()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	category := c.Query("category")

	items, total, err := model.GetPublishedApps(ctx, category, page, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": items, "total": total})
}

// GetApp 应用详情 — GET /api/apps/:id
func GetApp(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Translate(c, "invalid_parameter")})
		return
	}
	app, err := model.GetAppByID(ctx, uint(id))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": app})
}

// SubmitApp 用户提交应用 — POST /api/user/apps
func SubmitApp(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetInt("id")

	var app model.AppMarket
	if err := c.ShouldBindJSON(&app); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Translate(c, "invalid_input")})
		return
	}
	if app.Name == "" || app.Description == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Translate(c, "invalid_input")})
		return
	}
	app.UserID = uint(userID)
	app.Status = "draft"

	if err := model.InsertApp(ctx, &app); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": app})
}

// GetMyApps 用户查看自己的应用 — GET /api/user/apps
func GetMyApps(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetInt("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	items, total, err := model.GetUserApps(ctx, uint(userID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": items, "total": total})
}

// AdminListApps 管理员应用列表 — GET /api/admin/apps
func AdminListApps(c *gin.Context) {
	ctx := c.Request.Context()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")

	items, total, err := model.GetAllAppsPaginated(ctx, page, pageSize, status)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": items, "total": total})
}

// AdminUpdateAppStatus 管理员审核应用 — POST /api/admin/apps/:id/status
func AdminUpdateAppStatus(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Translate(c, "invalid_parameter")})
		return
	}
	var req struct {
		Status string `json:"status" validate:"required,oneof=draft published rejected"`
	}
	if err := common.Validate.Struct(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Translate(c, "invalid_input")})
		return
	}
	if err := model.UpdateAppStatus(ctx, uint(id), req.Status); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": ""})
}
