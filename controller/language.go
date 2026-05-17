package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/quantumclaw/quantumclaw/model"
)

type LanguageController struct{}

func NewLanguageController() *LanguageController {
	return &LanguageController{}
}

func (lc *LanguageController) RegisterRoutes(r *gin.RouterGroup) {
	langGroup := r.Group("/languages")
	{
		langGroup.GET("/types", lc.GetAllLanguageTypes)
		langGroup.POST("/types", lc.AddLanguageType)
		langGroup.DELETE("/types/:code", lc.DeleteLanguageType)
		langGroup.GET("/current", lc.GetCurrentLanguage)
		langGroup.POST("/current", lc.SetCurrentLanguage)
		langGroup.GET("/resources", lc.GetLanguageResources)
		langGroup.GET("/resources/search", lc.SearchLanguageResource)
		langGroup.POST("/resources", lc.AddLanguageResource)
		langGroup.PUT("/resources/:items", lc.UpdateLanguageResourceDisplay)
	}
}

func (lc *LanguageController) GetAllLanguageTypes(c *gin.Context) {
	langs, err := model.GetAllLanguageTypes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": langs})
}

func (lc *LanguageController) AddLanguageType(c *gin.Context) {
	var req struct {
		LanguageCode string `json:"language_code" binding:"required"`
		LanguageName string `json:"language_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := model.AddLanguageType(req.LanguageCode, req.LanguageName); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "language type added successfully"})
}

func (lc *LanguageController) DeleteLanguageType(c *gin.Context) {
	languageCode := c.Param("code")

	if err := model.DeleteLanguageType(languageCode); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "language type deleted successfully"})
}

func (lc *LanguageController) GetCurrentLanguage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"current_language": model.GetCurrentLanguage()})
}

func (lc *LanguageController) SetCurrentLanguage(c *gin.Context) {
	var req struct {
		LanguageCode string `json:"language_code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := model.SetCurrentLanguage(req.LanguageCode); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "language switched successfully", "current_language": req.LanguageCode})
}

func (lc *LanguageController) GetLanguageResources(c *gin.Context) {
	languageType := c.Query("language_type")
	if languageType == "" {
		languageType = model.GetCurrentLanguage()
	}

	resources, err := model.GetLanguageResources(languageType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": resources})
}

func (lc *LanguageController) SearchLanguageResource(c *gin.Context) {
	languageType := c.Query("language_type")
	if languageType == "" {
		languageType = model.GetCurrentLanguage()
	}

	lcode := c.Query("lcode")
	fromName := c.Query("from_name")

	if lcode != "" {
		resource, err := model.GetLanguageResourceByLCode(languageType, lcode)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "resource not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": resource})
		return
	}

	if fromName != "" {
		resources, err := model.GetLanguageResources(languageType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var filtered []model.LanguageResource
		for _, r := range resources {
			if r.FromName == fromName {
				filtered = append(filtered, r)
			}
		}
		c.JSON(http.StatusOK, gin.H{"data": filtered})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "please provide lcode or from_name parameter"})
}

func (lc *LanguageController) AddLanguageResource(c *gin.Context) {
	var resource model.LanguageResource
	if err := c.ShouldBindJSON(&resource); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := model.AddLanguageResource(resource); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "resource added successfully"})
}

func (lc *LanguageController) UpdateLanguageResourceDisplay(c *gin.Context) {
	var req struct {
		Display string `json:"display" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	itemsStr := c.Param("items")
	items := 0
	if itemsStr != "" {
		fmt.Sscanf(itemsStr, "%d", &items)
	}

	if err := model.UpdateLanguageResourceDisplay(items, req.Display); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "resource updated successfully"})
}