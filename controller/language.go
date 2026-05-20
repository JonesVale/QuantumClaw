package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/model"
)

// ==================== 语言控制器（原有 LanguageController） ====================

type LanguageController struct{}

func NewLanguageController() *LanguageController {
	return &LanguageController{}
}

func (ctl *LanguageController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/languages", ctl.GetLanguages)
	rg.GET("/translations", ctl.GetTranslations)
}

func (ctl *LanguageController) GetLanguages(c *gin.Context) {
	langs, err := model.GetLanguageTypes()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	type langItem struct {
		LanguagesType string `json:"languages_type"`
	}
	var result []langItem
	for _, l := range langs {
		result = append(result, langItem{LanguagesType: l.LanguagesType})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func (ctl *LanguageController) GetTranslations(c *gin.Context) {
	lang := c.Query("lang")
	if lang == "" {
		lang = "中文简体"
	}
	translations, err := model.GetTranslationsByLanguage(lang)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": translations})
}

// ==================== 新增无控制器 API（直接 handler，供无状态调用） ====================

// GetLanguages 直接返回所有可用语言版本列表
func GetLanguages(c *gin.Context) {
	langs, err := model.GetLanguageTypes()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": langs})
}

// GetTranslations 返回指定语言的所有翻译
// GET /api/translations?lang=中文简体
func GetTranslations(c *gin.Context) {
	lang := c.Query("lang")
	if lang == "" {
		lang = "中文简体"
	}
	translations, err := model.GetTranslationsByLanguage(lang)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": translations})
}

// UseLanguage 切换当前用户语言
// POST /api/language/switch  body: {"lang": "English"}
func UseLanguage(c *gin.Context) {
	var req struct {
		Lang string `json:"lang"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid request"})
		return
	}
	if req.Lang == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "lang is required"})
		return
	}
	// 校验语言是否存在
	langs, err := model.GetLanguageTypes()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	valid := false
	for _, l := range langs {
		if l.LanguagesType == req.Lang {
			valid = true
			break
		}
	}
	if !valid {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": fmt.Sprintf("unsupported language: %s", req.Lang)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": fmt.Sprintf("language switched to %s", req.Lang)})
}
