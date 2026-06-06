package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/model"
)

// GetTranslations 获取指定语言的所有翻译（公开）
// GET /api/translations/:langCode
func GetTranslations(c *gin.Context) {
	langCode := c.Param("langCode")
	if langCode == "" {
		langCode = "zh-CN"
	}

	translations, err := model.GetTranslationsByCode(langCode)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "获取翻译失败"})
		return
	}

	result := make(map[string]string)
	for _, t := range translations {
		result[t.LangKey] = t.Value
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// AdminGetTranslations 管理员搜索翻译
// GET /api/admin/translations?query=&langCode=&offset=0&limit=50
func AdminGetTranslations(c *gin.Context) {
	role := c.GetInt("role")
	if role < 10 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "权限不足"})
		return
	}

	query := c.DefaultQuery("query", "")
	langCode := c.DefaultQuery("langCode", "")
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	list, total, err := model.SearchTranslations(query, langCode, offset, limit)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "搜索失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    list,
		"total":   total,
	})
}

// AdminUpsertTranslation 管理员创建/更新翻译
// POST /api/admin/translations
func AdminUpsertTranslation(c *gin.Context) {
	role := c.GetInt("role")
	if role < 10 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "权限不足"})
		return
	}

	var req struct {
		LangKey  string `json:"lang_key" binding:"required"`
		LangCode string `json:"lang_code" binding:"required"`
		Value    string `json:"value" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误"})
		return
	}

	if err := model.UpsertTranslation(req.LangKey, req.LangCode, req.Value); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "保存翻译失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": ""})
}

// AdminBatchImportTranslations 管理员批量导入翻译
// POST /api/admin/translations/batch
func AdminBatchImportTranslations(c *gin.Context) {
	role := c.GetInt("role")
	if role < 10 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "权限不足"})
		return
	}

	var req struct {
		Translations []struct {
			LangKey  string `json:"lang_key"`
			LangCode string `json:"lang_code"`
			Value    string `json:"value"`
		} `json:"translations"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误"})
		return
	}

	var models []model.Translation
	for _, t := range req.Translations {
		models = append(models, model.Translation{
			LangKey:  t.LangKey,
			LangCode: t.LangCode,
			Value:    t.Value,
		})
	}

	count, err := model.BatchUpsertTranslations(models)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "批量导入失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": gin.H{"imported": count}})
}

// AdminDeleteTranslation 管理员删除翻译
// DELETE /api/admin/translations/:id
func AdminDeleteTranslation(c *gin.Context) {
	role := c.GetInt("role")
	if role < 10 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "权限不足"})
		return
	}

	id, _ := strconv.Atoi(c.Param("id"))
	if err := model.DeleteTranslation(id); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": ""})
}
