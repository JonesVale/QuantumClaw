package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/service"
)

//  Geo Injection Middleware
// Intercepts /v1/chat/completions requests, detects geo-related queries
// (weather, POI search, geocode), and injects the results as context.

// GeoRequestPreference holds per-request geo service preferences.
// Passed via headers: X-Enable-Geo: true/false, X-Geo-Query: explicit query.
type GeoRequestPreference struct {
	Enabled bool
	Auto    bool
	Query   string
}

func extractGeoPref(c *gin.Context) GeoRequestPreference {
	pref := GeoRequestPreference{}

	switch c.GetHeader("X-Enable-Geo") {
	case "true", "1", "yes":
		pref.Enabled = true
	case "false", "0", "no":
		return GeoRequestPreference{Enabled: false}
	}

	if q := c.GetHeader("X-Geo-Query"); q != "" {
		pref.Enabled = true
		pref.Query = q
	}

	return pref
}

// GeoMiddleware intercepts chat/completions requests and injects
// geo service results (weather, POI, etc.) as context.
// Should be placed after TokenAuth and SearchMiddleware, before PromptOptimizerMiddleware.
func GeoMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if !strings.HasSuffix(path, "/chat/completions") {
			c.Next()
			return
		}

		cfg := service.GetGeoConfig()
		if !cfg.Enabled {
			c.Next()
			return
		}

		pref := extractGeoPref(c)
		if !pref.Enabled {
			c.Next()
			return
		}

		// Read request body
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.Next()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		var req chatRequest
		if err := json.Unmarshal(bodyBytes, &req); err != nil || len(req.Messages) == 0 {
			c.Next()
			return
		}

		// Determine query and geo type
		query := pref.Query
		if query == "" {
			// Use last user message for auto-detection
			lastUserMsg := ""
			for i := len(req.Messages) - 1; i >= 0; i-- {
				if req.Messages[i].Role == "user" {
					lastUserMsg = req.Messages[i].Content
					break
				}
			}
			if lastUserMsg == "" {
				c.Next()
				return
			}
			query = lastUserMsg
		}

		// Detect geo intent
		geoType, location := service.DetectGeoIntent(query)
		if geoType == "" {
			c.Next()
			return
		}

		// Execute geo query
		ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(cfg.TimeoutSec)*time.Second)
		defer cancel()

		result, err := service.GeoQuery(ctx, geoType, location)
		if err != nil {
			logger.Warn(ctx, fmt.Sprintf("GeoMiddleware: query failed for %q: %v", location, err))
			c.Next()
			return
		}
		if result == nil || result.Error != "" {
			c.Next()
			return
		}

		// Format results and inject
		formatted := service.FormatGeoResultsForPrompt(result)
		if formatted == "" {
			c.Next()
			return
		}

		logger.Info(ctx, fmt.Sprintf("GeoMiddleware: injected %s result for %q (model: %s)",
			geoType, location, req.Model))

		// Inject as system message
		injected := false
		for i, msg := range req.Messages {
			if msg.Role == "system" {
				req.Messages[i].Content = msg.Content + "\n\n" + formatted
				injected = true
				break
			}
		}
		if !injected {
			req.Messages = append([]messageEntry{
				{Role: "system", Content: formatted},
			}, req.Messages...)
		}

		newBody, err := json.Marshal(req)
		if err != nil {
			c.Next()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(newBody))
		c.Request.ContentLength = int64(len(newBody))

		if cfg.CostPerQuery > 0 {
			c.Set("geo_performed", true)
			c.Set("geo_cost", cfg.CostPerQuery)
		}

		c.Next()
	}
}
