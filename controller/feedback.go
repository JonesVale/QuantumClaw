package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/i18n"
	"github.com/quantumclaw/quantumclaw/model"
)

// SubmitFeedback 提交用户反馈 — POST /api/user/feedback
func SubmitFeedback(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetInt("id")

	var req struct {
		Title   string `json:"title" validate:"required,max=200"`
		Content string `json:"content" validate:"required"`
		Type    string `json:"type" validate:"required,oneof=bug feature question"`
		Email   string `json:"email"`
	}
	if err := common.Validate.Struct(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Translate(c, "invalid_input")})
		return
	}

	fb := &model.Feedback{
		UserID:  uint(userID),
		Title:   req.Title,
		Content: req.Content,
		Type:    req.Type,
		Email:   req.Email,
	}
	if err := model.InsertFeedback(ctx, fb); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Translate(c, "submit_failed")})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": fb})
}

// ListFeedback 管理员查看反馈 — GET /api/admin/feedback
func ListFeedback(c *gin.Context) {
	ctx := c.Request.Context()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")
	fbType := c.Query("type")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	items, total, err := model.GetFeedbackPaginated(ctx, page, pageSize, status, fbType)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    items,
		"total":   total,
	})
}

// RespondFeedback 管理员回复反馈 — POST /api/admin/feedback/:id/respond
func RespondFeedback(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Translate(c, "invalid_parameter")})
		return
	}

	var req struct {
		Response string `json:"response" validate:"required"`
	}
	if err := common.Validate.Struct(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Translate(c, "invalid_input")})
		return
	}

	if err := model.RespondToFeedback(ctx, uint(id), req.Response); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Translate(c, "operation_failed")})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": ""})
}

// UpdateFeedbackStatus 更新反馈状态 — POST /api/admin/feedback/:id/status
func UpdateFeedbackStatus(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Translate(c, "invalid_parameter")})
		return
	}
	var req struct {
		Status string `json:"status" validate:"required,oneof=pending resolved closed"`
	}
	if err := common.Validate.Struct(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Translate(c, "invalid_input")})
		return
	}
	if err := model.UpdateFeedbackStatus(ctx, uint(id), req.Status); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Translate(c, "operation_failed")})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": ""})
}

// GetMyFeedback 用户查看自己的反馈 — GET /api/user/feedback
func GetMyFeedback(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetInt("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	items, total, err := model.GetUserFeedback(ctx, uint(userID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": items, "total": total})
}

// ========== FAQ ==========

// GetFAQs 公开 FAQ 列表 — GET /api/faq
func GetFAQs(c *gin.Context) {
	ctx := c.Request.Context()
	category := c.Query("category")
	items, err := model.GetFAQs(ctx, category)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": items})
}

// GetFAQCategories 公开 FAQ 分类 — GET /api/faq/categories
func GetFAQCategories(c *gin.Context) {
	ctx := c.Request.Context()
	cats, err := model.GetFAQCategories(ctx)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": cats})
}

// CreateFAQ 管理员创建 FAQ — POST /api/admin/faq
func CreateFAQ(c *gin.Context) {
	ctx := c.Request.Context()
	var faq model.FAQ
	if err := c.ShouldBindJSON(&faq); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Translate(c, "invalid_input")})
		return
	}
	if faq.Question == "" || faq.Answer == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Translate(c, "invalid_input")})
		return
	}
	if err := model.InsertFAQ(ctx, &faq); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": faq})
}

// UpdateFAQ 管理员更新 FAQ — PUT /api/admin/faq/:id
func UpdateFAQ(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Translate(c, "invalid_parameter")})
		return
	}
	var faq model.FAQ
	if err := c.ShouldBindJSON(&faq); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Translate(c, "invalid_input")})
		return
	}
	faq.ID = uint(id)
	if err := model.UpdateFAQ(ctx, &faq); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": ""})
}

// DeleteFAQ 管理员删除 FAQ — DELETE /api/admin/faq/:id
func DeleteFAQ(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": i18n.Translate(c, "invalid_parameter")})
		return
	}
	if err := model.DeleteFAQ(ctx, uint(id)); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": ""})
}
