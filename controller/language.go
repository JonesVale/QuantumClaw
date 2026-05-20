package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/model"
)

// GetLanguages 返回所有可用语言版本列表
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

// SeedTranslations — 从前端 JSON 导入翻译到 T_Languages
// POST /api/languages/seed  body: {"languages_type": "English", "entries": [{"lcode":"key","display":"val","fromname":"seed"}]}
func SeedTranslations(c *gin.Context) {
	var req struct {
		LanguagesType string `json:"languages_type"`
		Entries       []struct {
			LCode    string `json:"lcode"`
			Display  string `json:"display"`
			FromName string `json:"fromname,omitempty"`
		} `json:"entries"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid request"})
		return
	}
	if req.LanguagesType == "" || len(req.Entries) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "languages_type and entries required"})
		return
	}
	for _, e := range req.Entries {
		if e.LCode == "" {
			continue
		}
		model.DB.Create(&model.LanguageEntry{
			LanguagesType: req.LanguagesType,
			LCode:         e.LCode,
			Display:       e.Display,
		})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "seeded", "count": len(req.Entries)})
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
