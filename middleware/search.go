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

//  Search Injection Middleware 
// Intercepts /v1/chat/completions requests, optionally performs a web search,
// and injects the results as context into the system message.

// SearchMode defines how search is triggered.
type SearchMode string

const (
	SearchModeDisabled   SearchMode = "disabled"
	SearchModeAuto       SearchMode = "auto"       // auto-detect based on message content
	SearchModeManual     SearchMode = "manual"     // only when user explicitly requests
	SearchModeAlways     SearchMode = "always"     // always search (expensive!)
)

// UserSearchPreference stores per-user or per-request search preference.
// Passed via header: X-Enable-Search: true/false/auto
type UserSearchPreference struct {
	Enabled bool
	Mode    SearchMode
	Query   string // explicit search query (overrides auto-detect)
}

// extractSearchPref reads search preference from request headers.
func extractSearchPref(c *gin.Context) UserSearchPreference {
	pref := UserSearchPreference{
		Enabled: false,
		Mode:    SearchModeAuto,
	}

	// Check X-Enable-Search header
	switch c.GetHeader("X-Enable-Search") {
	case "true", "1", "yes":
		pref.Enabled = true
	case "false", "0", "no":
		pref.Enabled = false
		pref.Mode = SearchModeDisabled
		return pref
	}

	// Check X-Search-Query for explicit search query
	if q := c.GetHeader("X-Search-Query"); q != "" {
		pref.Enabled = true
		pref.Mode = SearchModeManual
		pref.Query = q
	}

	// Check X-Search-Mode
	switch c.GetHeader("X-Search-Mode") {
	case "auto":
		pref.Mode = SearchModeAuto
	case "always":
		pref.Mode = SearchModeAlways
		pref.Enabled = true
	case "manual":
		pref.Mode = SearchModeManual
	}

	return pref
}

// messageEntry represents a single message in the chat completions request.
type messageEntry struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatRequest is the minimal chat completions request structure.
type chatRequest struct {
	Model    string         `json:"model"`
	Messages []messageEntry `json:"messages"`
}

// SearchMiddleware intercepts chat/completions requests and injects
// web search results as context. It should be placed AFTER TokenAuth
// but BEFORE the relay chain.
func SearchMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if !strings.HasSuffix(path, "/chat/completions") {
			c.Next()
			return
		}

		cfg := service.GetWebSearchConfig()
		if !cfg.Enabled {
			c.Next()
			return
		}

		// Read user search preference
		pref := extractSearchPref(c)
		if !pref.Enabled && !cfg.AutoSearch {
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

		// Determine search query
		query := pref.Query
		if query == "" && pref.Mode == SearchModeAuto {
			// Use the last user message for auto-detect
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
			if !service.ShouldAutoSearch(lastUserMsg) {
				c.Next()
				return
			}
			query = lastUserMsg
		}

		if query == "" && pref.Mode == SearchModeManual {
			// Manual mode requires explicit X-Search-Query header
			c.Next()
			return
		}

		if query == "" {
			c.Next()
			return
		}

		// Perform search with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(cfg.TimeoutSec)*time.Second)
		defer cancel()

		searchResult, err := service.Search(ctx, query)
		if err != nil {
			logger.Warn(ctx, fmt.Sprintf("SearchMiddleware: search failed for %q: %v", query, err))
			c.Next()
			return
		}

		if len(searchResult.Results) == 0 {
			c.Next()
			return
		}

		// Format search results and inject into system message
		formatted := service.FormatResultsForPrompt(searchResult)
		logger.Info(ctx, fmt.Sprintf("SearchMiddleware: injected %d search results for query %q (model: %s)",
			len(searchResult.Results), query, req.Model))

		// Inject as first system message or append to existing system message
		injected := false
		for i, msg := range req.Messages {
			if msg.Role == "system" {
				req.Messages[i].Content = msg.Content + "\n\n" + formatted
				injected = true
				break
			}
		}
		if !injected {
			// No existing system message, prepend one
			req.Messages = append([]messageEntry{
				{Role: "system", Content: formatted},
			}, req.Messages...)
		}

		// Rebuild request body
		newBody, err := json.Marshal(req)
		if err != nil {
			c.Next()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(newBody))
		c.Request.ContentLength = int64(len(newBody))

		// Deduct quota for the search
		if cfg.CostPerSearch > 0 {
			// Store search cost info in context for billing
			c.Set("search_performed", true)
			c.Set("search_cost", cfg.CostPerSearch)
		}

		c.Next()
	}
}
