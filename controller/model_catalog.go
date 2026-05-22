package controller

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/model"
)

// GetModelCatalog returns all model metadata for the given language,
// merged with channel pricing/status where available.
// GET /api/model-catalog?lang=English&search=&use_case=&series=
func GetModelCatalog(c *gin.Context) {
	lang := c.DefaultQuery("lang", "English")
	search := c.Query("search")
	useCase := c.Query("use_case")
	series := c.Query("series")
	modality := c.Query("modality")

	// 1. Fetch model_metadata for the requested language
	var metadata []model.ModelMetadata
	query := model.DB.Where("languages_type = ?", lang)

	if search != "" {
		q := "%" + search + "%"
		query = query.Where("model_name LIKE ? OR description LIKE ?", q, q)
	}
	if useCase != "" {
		query = query.Where("use_case = ?", useCase)
	}
	if series != "" {
		query = query.Where("series LIKE ?", "%"+series+"%")
	}

	if err := query.Find(&metadata).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 2. Fetch all enabled channels for pricing/status
	var channels []model.Channel
	model.DB.Where("status = ?", model.ChannelStatusEnabled).Find(&channels)
	channelMap := make(map[string]*model.Channel)
	for _, ch := range channels {
		models := strings.Split(ch.Models, ",")
		for _, m := range models {
			m = strings.TrimSpace(m)
			if m != "" {
				channelMap[strings.ToLower(m)] = &ch
			}
		}
	}

	// 3. Build response
	var result []model.CatalogModelResponse
	for _, m := range metadata {
		resp := model.CatalogModelResponse{
			Name:            m.ModelName,
			DisplayName:     m.DisplayName,
			Description:     m.Description,
			UseCase:         m.UseCase,
			ContextWindow:   m.ContextWindow,
			InputModalities: m.Modalities(),
			Series:          m.Series,
			Provider:        m.Provider,
			Status:          0, // unconfigured by default
		}

		// Check if any channel has this model
		key := strings.ToLower(m.ModelName)
		if ch, ok := channelMap[key]; ok {
			resp.ChannelID = ch.Id
			resp.ChannelName = ch.Name
			perTokenCost := ch.CostPerUnit / 1000.0
			resp.InputPrice = perTokenCost
			resp.OutputPrice = perTokenCost * ch.SellPriceRate
			resp.Status = 1
			resp.Group = ch.Group
		}

		result = append(result, resp)
	}

	// Filter by modality if requested
	if modality != "" {
		var filtered []model.CatalogModelResponse
		for _, r := range result {
			for _, m := range r.InputModalities {
				if strings.EqualFold(m, modality) {
					filtered = append(filtered, r)
					break
				}
			}
		}
		result = filtered
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetModelDetail returns metadata for a single model.
// GET /api/model-catalog/:model_name?lang=English
func GetModelDetail(c *gin.Context) {
	lang := c.DefaultQuery("lang", "English")
	modelName := c.Param("model_name")

	var meta model.ModelMetadata
	if err := model.DB.Where("model_name = ? AND languages_type = ?", modelName, lang).First(&meta).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "model not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    meta,
	})
}

// SyncModelMetadata detects new models from channels and inserts metadata rows.
// POST /api/models/sync
func SyncModelMetadata(c *gin.Context) {
	// 1. Collect all model names from all enabled channels
	var channels []model.Channel
	model.DB.Where("status = ?", model.ChannelStatusEnabled).Find(&channels)

	seen := make(map[string]bool)
	for _, ch := range channels {
		models := strings.Split(ch.Models, ",")
		for _, m := range models {
			m = strings.TrimSpace(m)
			if m != "" {
				seen[strings.ToLower(m)] = true
			}
		}
	}

	// 2. Check which model names already have metadata
	languages := []string{"中文简体", "中文繁体", "English", "Français", "日本語", "Русский", "Tiếng Việt"}
	var existing []model.ModelMetadata
	model.DB.Find(&existing)
	existingMap := make(map[string]bool)
	for _, e := range existing {
		existingMap[strings.ToLower(e.ModelName)+":"+e.LanguagesType] = true
	}

	// 3. Insert new models that don't have metadata yet
	now := now()
	count := 0
	for name := range seen {
		hasAny := false
		for _, lang := range languages {
			if existingMap[name+":"+lang] {
				hasAny = true
				break
			}
		}
		if !hasAny {
			for _, lang := range languages {
				m := &model.ModelMetadata{
					ModelName:       name,
					LanguagesType:   lang,
					DisplayName:     name,
					Description:     "New model — description pending.",
					UseCase:         "chat",
					ContextWindow:   128000,
					InputModalities: `["Text"]`,
					Series:          "Other",
					Provider:        "Unknown",
					CreatedTime:     now,
					UpdatedTime:     now,
				}
				if err := model.DB.Create(m).Error; err == nil {
					count++
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  "sync complete",
		"inserted": count,
	})
}

func now() int64 {
	return time.Now().Unix()
}
