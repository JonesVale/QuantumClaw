package controller

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/relay/channeltype"
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

		// Official reference price from reference_prices table
		// Seeded from ModelRatio on startup; updated monthly from official pricing pages.
		if rp := model.GetReferencePrice(m.ModelName); rp != nil {
			resp.InputPrice = rp.InputPrice
			resp.OutputPrice = rp.OutputPrice
		}

		// Check if any channel has this model (for Group/Status metadata)
		key := strings.ToLower(m.ModelName)
		if ch, ok := channelMap[key]; ok {
			resp.ChannelID = ch.Id
			resp.ChannelName = ch.Name
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
	// Collect all model names from all enabled channels
	var channels []model.Channel
	model.DB.Where("status = ?", model.ChannelStatusEnabled).Find(&channels)

	typeNames := channeltype.ChannelTypeNames

	seen := make(map[string]int) // model_name → channelType
	for _, ch := range channels {
		models := strings.Split(ch.Models, ",")
		for _, m := range models {
			m = strings.TrimSpace(m)
			if m != "" {
				seen[strings.ToLower(m)] = ch.Type
			}
		}
	}

	languages := []string{"中文简体", "中文繁体", "English", "Français", "日本語", "Русский", "Tiếng Việt"}
	var existing []model.ModelMetadata
	model.DB.Find(&existing)
	existingMap := make(map[string]bool)
	for _, e := range existing {
		existingMap[strings.ToLower(e.ModelName)+":"+e.LanguagesType] = true
	}

	now := now()
	count := 0
	type channelInfo struct {
		provider string
		chType   int
	}
	modelToChannel := make(map[string]channelInfo)
	for _, ch := range channels {
		provider := ""
		if p, ok := typeNames[ch.Type]; ok {
			provider = p
		}
		for _, m := range strings.Split(ch.Models, ",") {
			m = strings.TrimSpace(m)
			if m != "" {
				key := strings.ToLower(m)
				if _, exists := modelToChannel[key]; !exists {
					modelToChannel[key] = channelInfo{provider: provider, chType: ch.Type}
				}
			}
		}
	}

	for name, chType := range seen {
		hasAny := false
		for _, lang := range languages {
			if existingMap[name+":"+lang] {
				hasAny = true
				break
			}
		}
		if hasAny {
			continue
		}

		// 判断是否为量子模型
		isQuantum := chType >= 100 && chType < channeltype.QuantumDummy
		useCase := "chat"
		if isQuantum {
			useCase = "quantum"
		}

		info, _ := modelToChannel[strings.ToLower(name)]

		// Auto-generate description
		parts := strings.SplitN(name, "/", 2)
		shortName := name
		provider := info.provider
		if provider == "" {
			provider = "Unknown"
		}
		if len(parts) == 2 {
			shortName = parts[1]
			if provider == "Unknown" {
				provider = strings.Title(parts[0])
			}
		}
		shortName = strings.ReplaceAll(shortName, "-", " ")
		shortName = strings.ReplaceAll(shortName, "_", " ")
		shortName = strings.Title(shortName)

		var enDesc, cnDesc string
		if isQuantum {
			enDesc = shortName + " is a quantum computing model from " + provider + ", supporting quantum circuit execution and hybrid quantum-classical computation."
			cnDesc = shortName + " 是 " + provider + " 的量子计算模型，支持量子线路执行和混合量子-经典计算。"
		} else {
			enDesc = shortName + " is a model from " + provider + "."
			cnDesc = provider + " 的 " + shortName + " 模型。"
		}

		contextWin := 128000
		if isQuantum {
			contextWin = 1000000 // Quantum models have million-scale context
		}

		series := provider
		if isQuantum {
			series = "Quantum"
		}

		for _, lang := range languages {
			desc := enDesc
			switch lang {
			case "中文简体", "中文繁体":
				desc = cnDesc
			}
			m := &model.ModelMetadata{
				ModelName:       name,
				LanguagesType:   lang,
				DisplayName:     name,
				Description:     desc,
				UseCase:         useCase,
				ContextWindow:   contextWin,
				InputModalities: `["Text"]`,
				Series:          series,
				Provider:        provider,
				CreatedTime:     now,
				UpdatedTime:     now,
			}
			if err := model.DB.Create(m).Error; err == nil {
				count++
			}
		}
	}

	// 4. 更新已有模型：量子渠道的标记 use_case="quantum"
	updateCount := 0
	for _, ch := range channels {
		isQuantum := ch.Type >= 100 && ch.Type < channeltype.QuantumDummy
		if !isQuantum {
			continue
		}
		for _, m := range strings.Split(ch.Models, ",") {
			m = strings.TrimSpace(m)
			if m == "" {
				continue
			}
			result := model.DB.Model(&model.ModelMetadata{}).
				Where("model_name = ? AND use_case != 'quantum'", m).
				Update("use_case", "quantum")
			if result.Error == nil {
				updateCount += int(result.RowsAffected)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  "sync complete",
		"inserted": count,
		"updated":  updateCount,
	})
}

func now() int64 {
	return time.Now().Unix()
}
