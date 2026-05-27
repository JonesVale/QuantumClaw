package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/quantumclaw/quantumclaw/model"
	"github.com/gin-gonic/gin"
)

// ── Schema CRUD ──

func ListSub2APISchemas(c *gin.Context) {
	schemas, err := model.ListSub2APISchemas()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": schemas})
}

func GetSub2APISchema(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid id"})
		return
	}
	s, err := model.GetSub2APISchema(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "schema not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": s})
}

func CreateSub2APISchema(c *gin.Context) {
	var s model.Sub2APISchema
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid request: " + err.Error()})
		return
	}
	if err := model.CreateSub2APISchema(&s); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "schema created", "data": s})
}

func UpdateSub2APISchema(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid id"})
		return
	}
	existing, err := model.GetSub2APISchema(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "schema not found"})
		return
	}
	if err := c.ShouldBindJSON(existing); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid request: " + err.Error()})
		return
	}
	existing.Id = id
	if err := model.UpdateSub2APISchema(existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "schema updated"})
}

func DeleteSub2APISchema(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid id"})
		return
	}
	if err := model.DeleteSub2APISchema(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "schema deleted"})
}

// TestSub2APISchema validates a schema by performing a dry-run parse of its template.
func TestSub2APISchema(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid id"})
		return
	}
	s, err := model.GetSub2APISchema(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "schema not found"})
		return
	}
	// Validate JSON templates parse correctly
	var headers map[string]interface{}
	if err := json.Unmarshal([]byte(s.HeadersTemplate), &headers); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false, "message": "headers_template is not valid JSON",
			"data": gin.H{"schema_id": id, "provider": s.Provider, "version": s.Version},
		})
		return
	}
	var mapping map[string]interface{}
	if s.ModelMapping != "" && s.ModelMapping != "{}" {
		if err := json.Unmarshal([]byte(s.ModelMapping), &mapping); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false, "message": "model_mapping is not valid JSON",
				"data": gin.H{"schema_id": id, "provider": s.Provider, "version": s.Version},
			})
			return
		}
	}
	// Test model mapping for a known model
	mapped, ok := s.MapModel("gpt-4o")
	_ = ok
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "schema validation passed",
		"data": gin.H{
			"schema_id":          id,
			"provider":           s.Provider,
			"version":            s.Version,
			"status":             s.Status,
			"headers_valid":      true,
			"model_mapping_valid": true,
			"model_mapping_test": map[string]string{
				"gpt-4o → " + s.Provider: mapped,
			},
			"auth_type":    s.AuthType,
			"auth_key":     s.AuthKeyName,
			"stream_mode":  s.StreamMode,
			"response_path": s.ResponsePath,
		},
	})
}

// ── Schema Health Dashboard ──

func GetSub2APISchemaHealth(c *gin.Context) {
	schemas, err := model.ListSub2APISchemas()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	// Aggregate by provider
	type ProviderHealth struct {
		Provider    string `json:"provider"`
		Total       int    `json:"total"`
		Active      int    `json:"active"`
		Draft       int    `json:"draft"`
		Deprecated  int    `json:"deprecated"`
		Broken      int    `json:"broken"`
		HighestVer  int    `json:"highest_version"`
		LastHealth  int64  `json:"last_health_at"`
		LastError   string `json:"last_error"`
	}
	providerMap := make(map[string]*ProviderHealth)
	for _, s := range schemas {
		if _, ok := providerMap[s.Provider]; !ok {
			providerMap[s.Provider] = &ProviderHealth{
				Provider: s.Provider,
			}
		}
		ph := providerMap[s.Provider]
		ph.Total++
		if s.Version > ph.HighestVer {
			ph.HighestVer = s.Version
		}
		switch s.Status {
		case 1:
			ph.Active++
		case 2:
			ph.Draft++
		case 3:
			ph.Deprecated++
		case 4:
			ph.Broken++
		}
		if s.LastHealthAt > ph.LastHealth {
			ph.LastHealth = s.LastHealthAt
		}
		if s.LastError != "" {
			ph.LastError = s.LastError
		}
	}
	result := make([]*ProviderHealth, 0, len(providerMap))
	for _, ph := range providerMap {
		result = append(result, ph)
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": result})
}

// ── Manual Trigger Health Check ──

func TriggerSub2APIHealthCheck(c *gin.Context) {
	// Just marks all active schemas as needing re-check
	// The actual health polling runs in the background
	cnt := model.DB.Model(&model.Sub2APISchema{}).
		Where("status = 1").
		Update("last_health_at", 0).RowsAffected
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "health check triggered",
		"data":    gin.H{"schemas_queued": cnt},
	})
}
