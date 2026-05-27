package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/service"
)

//   

// GetSearchConfig 
// GET /api/search/config
func GetSearchConfig(c *gin.Context) {
	cfg := service.GetWebSearchConfig()
	// Redact API key for frontend display
	safeCfg := struct {
		Enabled       bool   `json:"enabled"`
		Provider      string `json:"provider"`
		HasAPIKey     bool   `json:"has_api_key"`
		Endpoint      string `json:"endpoint"`
		MaxResults    int    `json:"max_results"`
		TimeoutSec    int    `json:"timeout_sec"`
		AutoSearch    bool   `json:"auto_search"`
		CostPerSearch int64  `json:"cost_per_search"`
	}{
		Enabled:       cfg.Enabled,
		Provider:      cfg.Provider,
		HasAPIKey:     cfg.APIKey != "",
		Endpoint:      cfg.Endpoint,
		MaxResults:    cfg.MaxResults,
		TimeoutSec:    cfg.TimeoutSec,
		AutoSearch:    cfg.AutoSearch,
		CostPerSearch: cfg.CostPerSearch,
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": safeCfg})
}

// UpdateSearchConfig 
// PUT /api/search/config
func UpdateSearchConfig(c *gin.Context) {
	var req struct {
		Enabled       *bool   `json:"enabled"`
		Provider      *string `json:"provider"`
		APIKey        *string `json:"api_key"`
		Endpoint      *string `json:"endpoint"`
		MaxResults    *int    `json:"max_results"`
		TimeoutSec    *int    `json:"timeout_sec"`
		AutoSearch    *bool   `json:"auto_search"`
		CostPerSearch *int64  `json:"cost_per_search"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid request"})
		return
	}

	cfg := service.GetWebSearchConfig()

	if req.Enabled != nil {
		cfg.Enabled = *req.Enabled
	}
	if req.Provider != nil {
		cfg.Provider = *req.Provider
	}
	if req.APIKey != nil && *req.APIKey != "" {
		cfg.APIKey = *req.APIKey
	}
	if req.Endpoint != nil {
		cfg.Endpoint = *req.Endpoint
	}
	if req.MaxResults != nil {
		cfg.MaxResults = *req.MaxResults
	}
	if req.TimeoutSec != nil {
		cfg.TimeoutSec = *req.TimeoutSec
	}
	if req.AutoSearch != nil {
		cfg.AutoSearch = *req.AutoSearch
	}
	if req.CostPerSearch != nil {
		cfg.CostPerSearch = *req.CostPerSearch
	}

	service.UpdateWebSearchConfig(cfg)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "search config updated"})
}

// TestSearch 
// POST /api/search/test
func TestSearch(c *gin.Context) {
	var req struct {
		Query string `json:"query" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "query is required"})
		return
	}

	result, err := service.Search(c.Request.Context(), req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	if result.Error != "" {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": result.Error})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// GetSearchProviders 
// GET /api/search/providers
func GetSearchProviders(c *gin.Context) {
	providers := []struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		NeedKey     bool   `json:"need_key"`
		NeedEndpoint bool  `json:"need_endpoint"`
	}{
		{
			ID: "bing", Name: "Bing Search",
			Description: "微软必应搜索 API，需�?Azure 订阅获取 API Key",
			NeedKey: true, NeedEndpoint: false,
		},
		{
			ID: "searxng", Name: "SearXNG",
			Description: "自托管搜索引擎，无需 API Key，需要部�?SearXNG 实例",
			NeedKey: false, NeedEndpoint: true,
		},
		{
			ID: "serpapi", Name: "SerpAPI",
			Description: "Google 搜索结果接口，需�?SerpAPI 订阅",
			NeedKey: true, NeedEndpoint: false,
		},
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": providers})
}
