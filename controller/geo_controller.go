package controller

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/service"
)

// GetGeoConfig returns the current geo service configuration (admin only).
func GetGeoConfig(c *gin.Context) {
	cfg := service.GetGeoConfig()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"enabled":       cfg.Enabled,
			"provider":      cfg.Provider,
			"max_results":   cfg.MaxResults,
			"timeout_sec":   cfg.TimeoutSec,
			"region_filter": cfg.RegionFilter,
			"cost_per_query": cfg.CostPerQuery,
		},
	})
}

// UpdateGeoConfig updates the geo service configuration (admin only).
func UpdateGeoConfig(c *gin.Context) {
	var req struct {
		Enabled       *bool   `json:"enabled"`
		Provider      string  `json:"provider"`
		APIKey        string  `json:"api_key"`
		GeoCodeKey    string  `json:"geocode_key"`
		MaxResults    *int    `json:"max_results"`
		TimeoutSec    *int    `json:"timeout_sec"`
		RegionFilter  string  `json:"region_filter"`
		CostPerQuery  *int64  `json:"cost_per_query"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid request"})
		return
	}

	cfg := service.GetGeoConfig()
	if req.Enabled != nil {
		cfg.Enabled = *req.Enabled
	}
	if req.Provider != "" {
		cfg.Provider = req.Provider
	}
	if req.APIKey != "" {
		cfg.APIKey = req.APIKey
	}
	if req.GeoCodeKey != "" {
		cfg.GeoCodeKey = req.GeoCodeKey
	}
	if req.MaxResults != nil {
		cfg.MaxResults = *req.MaxResults
	}
	if req.TimeoutSec != nil {
		cfg.TimeoutSec = *req.TimeoutSec
	}
	if req.RegionFilter != "" {
		cfg.RegionFilter = req.RegionFilter
	}
	if req.CostPerQuery != nil {
		cfg.CostPerQuery = *req.CostPerQuery
	}

	service.UpdateGeoConfig(cfg)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "geo config updated"})
}

// GetGeoProviders returns available geo service providers.
func GetGeoProviders(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": []gin.H{
			{"id": "amap", "name": "Amap (High德地图)", "region": "china"},
			{"id": "google", "name": "Google Maps", "region": "global"},
		},
	})
}

// TestGeo tests a geo service query for validation.
func TestGeo(c *gin.Context) {
	var req struct {
		Query     string `json:"query"`
		QueryType string `json:"query_type"` // weather, poi, geocode
		Provider  string `json:"provider"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid request"})
		return
	}

	if req.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "query is required"})
		return
	}

	geoType := service.GeoQueryType(req.QueryType)
	if geoType == "" {
		geoType = service.GeoTypePOI
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	result, err := service.GeoQuery(ctx, geoType, req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	formatted := service.FormatGeoResultsForPrompt(result)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "",
		"data":      result,
		"formatted": formatted,
	})
}

// GeoRedirectHandler handles user-facing geo service requests (non-admin).
// GET /api/geo/query?type=weather&q=Beijing
// GET /api/geo/query?type=poi&q=restaurants+near+me
// GET /api/geo/query?type=geocode&q=Shanghai+Tower
func GeoRedirectHandler(c *gin.Context) {
	queryType := c.Query("type")
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "missing query parameter 'q'"})
		return
	}

	cfg := service.GetGeoConfig()
	if !cfg.Enabled {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "geo service not enabled"})
		return
	}

	var geoType service.GeoQueryType
	switch queryType {
	case "weather":
		geoType = service.GeoTypeWeather
	case "poi", "places":
		geoType = service.GeoTypePOI
	case "geocode":
		geoType = service.GeoTypeGeocode
	default:
		// Auto-detect
		detectedType, _ := service.DetectGeoIntent(query)
		if detectedType == "" {
			geoType = service.GeoTypePOI
		} else {
			geoType = detectedType
		}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(cfg.TimeoutSec)*time.Second)
	defer cancel()

	result, err := service.GeoQuery(ctx, geoType, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	formatted := service.FormatGeoResultsForPrompt(result)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "",
		"data":      result,
		"formatted": formatted,
	})
}
