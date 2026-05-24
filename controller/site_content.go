package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/model"
	"gorm.io/gorm"
)

// ── Data structures ──

type SiteStat struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Label  string `json:"label"`
	Detail string `json:"detail"`
}

type SiteFeature struct {
	Key      string `json:"key"`
	Title    string `json:"title"`
	Desc     string `json:"desc"`
	IconName string `json:"icon_name"`
	Color    string `json:"color"`
}

type SiteProvider struct {
	Name   string `json:"name"`
	Models string `json:"models"`
}

type SiteContent struct {
	Stats     []SiteStat     `json:"stats"`
	Features  []SiteFeature  `json:"features"`
	Providers []SiteProvider `json:"providers"`
}

// ── Handlers ──

func GetSiteContent(c *gin.Context) {
	db := model.DB
	if db == nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": SiteContent{
			Stats:     getDefaultStats(),
			Features:  getDefaultFeatures(),
			Providers: getDefaultProviders(),
		}})
		return
	}

	stats := fetchStats(db)
	features := fetchFeatures(db)
	providers := fetchProviders(db)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": SiteContent{
			Stats:     stats,
			Features:  features,
			Providers: providers,
		},
	})
}

func GetSiteStats(c *gin.Context) {
	db := model.DB
	if db == nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": getDefaultStats()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": fetchStats(db)})
}

func GetSiteFeatures(c *gin.Context) {
	db := model.DB
	if db == nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": getDefaultFeatures()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": fetchFeatures(db)})
}

func GetSiteProviders(c *gin.Context) {
	db := model.DB
	if db == nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": getDefaultProviders()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": fetchProviders(db)})
}

// ── Database queries ──

func fetchStats(db *gorm.DB) []SiteStat {
	var rows []struct {
		StatKey        string `gorm:"column:stat_key"`
		StatValue      string `gorm:"column:stat_value"`
		StatLabelKey   string `gorm:"column:stat_label_key"`
		StatDetail     string `gorm:"column:stat_detail"`
		SortOrder      int    `gorm:"column:sort_order"`
	}
	db.Table("site_stats").Order("sort_order ASC").Find(&rows)
	if len(rows) == 0 {
		return getDefaultStats()
	}
	stats := make([]SiteStat, len(rows))
	for i, r := range rows {
		stats[i] = SiteStat{
			Key:    r.StatKey,
			Value:  r.StatValue,
			Label:  r.StatLabelKey,
			Detail: r.StatDetail,
		}
	}
	return stats
}

func fetchFeatures(db *gorm.DB) []SiteFeature {
	var rows []struct {
		FeatureKey string `gorm:"column:feature_key"`
		TitleKey   string `gorm:"column:title_key"`
		DescKey    string `gorm:"column:desc_key"`
		IconName   string `gorm:"column:icon_name"`
		SortOrder  int    `gorm:"column:sort_order"`
	}
	db.Table("site_features").Order("sort_order ASC").Find(&rows)
	if len(rows) == 0 {
		return getDefaultFeatures()
	}
	features := make([]SiteFeature, len(rows))
	for i, r := range rows {
		features[i] = SiteFeature{
			Key:      r.FeatureKey,
			Title:    r.TitleKey,
			Desc:     r.DescKey,
			IconName: r.IconName,
			Color:    "from-amber-400 to-orange-500",
		}
	}
	return features
}

func fetchProviders(db *gorm.DB) []SiteProvider {
	var rows []struct {
		ProviderName string `gorm:"column:provider_name"`
		ModelCount   string `gorm:"column:model_count"`
		SortOrder    int    `gorm:"column:sort_order"`
	}
	db.Table("site_providers").Order("sort_order ASC").Find(&rows)
	if len(rows) == 0 {
		return getDefaultProviders()
	}
	providers := make([]SiteProvider, len(rows))
	for i, r := range rows {
		providers[i] = SiteProvider{Name: r.ProviderName, Models: r.ModelCount}
	}
	return providers
}

// ── Fallbacks (当数据库无数据时) ──

func getDefaultStats() []SiteStat {
	return []SiteStat{
		{Key: "models", Value: "400+", Label: "AI Models", Detail: "Chat, Code, Vision, Audio"},
		{Key: "providers", Value: "60+", Label: "Providers", Detail: "OpenAI, Anthropic, Google..."},
		{Key: "tokens", Value: "50M+", Label: "Daily Tokens", Detail: "Processed per day"},
		{Key: "uptime", Value: "99.9%", Label: "Uptime SLA", Detail: "Enterprise grade"},
	}
}

func getDefaultFeatures() []SiteFeature {
	return []SiteFeature{
		{Key: "chat", Title: "Chat & Assistant", Desc: "GPT-4, Claude, Gemini and more", IconName: "bolt", Color: "from-amber-400 to-orange-500"},
		{Key: "code", Title: "Code Generation", Desc: "DeepSeek Coder, Code Llama, Qwen Coder", IconName: "cpu", Color: "from-amber-400 to-orange-500"},
		{Key: "reason", Title: "Reasoning & Logic", Desc: "Advanced reasoning models for complex tasks", IconName: "sparkle", Color: "from-amber-400 to-orange-500"},
		{Key: "quantum", Title: "Quantum Computing", Desc: "IonQ, IBM, Rigetti quantum processors", IconName: "layers", Color: "from-amber-400 to-orange-500"},
	}
}

func getDefaultProviders() []SiteProvider {
	return []SiteProvider{
		{Name: "OpenAI", Models: "20+"}, {Name: "Anthropic", Models: "8+"},
		{Name: "Google", Models: "15+"}, {Name: "DeepSeek", Models: "12+"},
		{Name: "Meta", Models: "10+"}, {Name: "Mistral", Models: "6+"},
		{Name: "Cohere", Models: "5+"}, {Name: "Groq", Models: "4+"},
	}
}
