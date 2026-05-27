package service

import (
	"encoding/json"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
)

// Sub2APIDiscovery handles health checking and automatic schema fallback.
type Sub2APIDiscovery struct {
	stopCh chan struct{}
}

var Sub2APIDiscoverySvc = &Sub2APIDiscovery{
	stopCh: make(chan struct{}),
}

// StartHealthChecker begins periodic health checks for all active schemas.
// Runs every HealthCheckInterval; marks broken schemas and triggers fallback.
func (d *Sub2APIDiscovery) StartHealthChecker(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Run immediately on start
		d.checkAll()

		for {
			select {
			case <-ticker.C:
				d.checkAll()
			case <-d.stopCh:
				logger.SysLog("Sub2API health checker stopped")
				return
			}
		}
	}()
	logger.SysLog("Sub2API health checker started (interval: " + interval.String() + ")")
}

// StopHealthChecker stops the background health checker.
func (d *Sub2APIDiscovery) StopHealthChecker() {
	close(d.stopCh)
}

// checkAll iterates all active schemas and performs a lightweight validation.
func (d *Sub2APIDiscovery) checkAll() {
	schemas, err := model.ListSub2APISchemas()
	if err != nil {
		logger.SysError("Sub2API health check: list schemas error: " + err.Error())
		return
	}

	now := time.Now().UnixMilli()
	for _, s := range schemas {
		if s.Status != 1 {
			continue
		}

		// Lightweight check: verify template validity
		// Full connectivity check requires actual credentials - done per-user-request
		hadIssue := false

		// Check: endpoint URL not empty
		if s.EndpointURL == "" {
			s.LastError = "endpoint_url is empty"
			hadIssue = true
		}

		// Check: auth fields valid
		if s.AuthType == "" {
			s.LastError = "auth_type is empty"
			hadIssue = true
		}

		// Check: template is valid JSON
		if !isValidJSON(s.RequestTemplate) {
			s.LastError = "request_template is not valid JSON"
			hadIssue = true
		}

		if hadIssue {
			// Mark as broken after 3 consecutive failures
			// (for now just update last_error)
			s.LastHealthAt = now
			s.LastError = s.LastError + " (detected at " + time.Now().Format("15:04:05") + ")"
			model.UpdateSub2APISchema(&s)
			continue
		}

		// Schema looks good - update health timestamp
		s.LastHealthAt = now
		s.LastError = ""
		model.UpdateSub2APISchema(&s)
	}
}

// FindAvailableSchema tries active schemas in priority order, falling back on failure.
func (d *Sub2APIDiscovery) FindAvailableSchema(provider string) (*model.Sub2APISchema, error) {
	schemas, err := model.ListActiveSchemasByProvider(provider)
	if err != nil || len(schemas) == 0 {
		return nil, err
	}
	// First check if any schema has been recently verified
	var bestSchema *model.Sub2APISchema
	recentThreshold := time.Now().Add(-6 * time.Hour).UnixMilli()

	for _, s := range schemas {
		if s.LastHealthAt >= recentThreshold && s.LastError == "" {
			// Verified healthy within 6 hours
			return &s, nil
		}
		if bestSchema == nil && s.LastHealthAt > 0 {
			bestSchema = &s
		}
	}
	// Fallback: return the best available (highest priority, last known good)
	if bestSchema != nil {
		return bestSchema, nil
	}
	if len(schemas) > 0 {
		return &schemas[0], nil // last resort
	}
	return nil, nil
}

// isValidJSON checks if a string is valid JSON.
func isValidJSON(s string) bool {
	if s == "" {
		return false
	}
	var v interface{}
	return json.Unmarshal([]byte(s), &v) == nil
}
