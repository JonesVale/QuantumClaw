package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/quantumclaw/quantumclaw/common/ctxkey"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/sub2api"
	"github.com/quantumclaw/quantumclaw/relay/channeltype"
)

// Sub2APIRouter intercepts requests that can be served by Sub2API web model providers.
// It runs BEFORE Distribute(). If the user has a valid Sub2API credential + schema for
// the requested model, it sets up a virtual Sub2API channel context and skips normal
// channel selection.
//
// Middleware order:
//
//	Auth → Search → Geo → PromptOpt → ParamValidate → IntelligentRouter
//	→ Sub2APIRouter (NEW) → ModelRateLimit → Distribute → relay
func Sub2APIRouter() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only intercept chat/completions requests
		if c.Request.Method != "POST" {
			c.Next()
			return
		}
		path := c.Request.URL.Path
		if path != "/v1/chat/completions" && path != "/v1/completions" {
			c.Next()
			return
		}

		// Check if Sub2API can handle this request
		if !sub2api.MatchAndSetProvider(c) {
			c.Next() // No Sub2API match → continue to normal Distribute
			return
		}

		// Set up a virtual Sub2API channel context
		overrideChannelContext(c)

		// Skip distributor (we handle it ourselves)
		c.Next()
	}
}

// overrideChannelContext sets up the Gin context to look like a Sub2API channel was selected.
// This makes the relay controller think it's using a normal channel, while actually the
// Sub2API adaptor handles everything.
func overrideChannelContext(c *gin.Context) {
	// These mirror what Distribute.setupAndContinue does
	c.Set(ctxkey.Channel, channeltype.Sub2API)
	c.Set(ctxkey.ChannelId, -1) // virtual channel, no real DB record
	c.Set(ctxkey.ChannelName, "Sub2API")
	c.Set(ctxkey.ChannelOwner, 0)

	// Get the original model from the parsed request body
	var requestModel string
	if body, exists := c.Get("parsed_request_body"); exists {
		if m, ok := body.(map[string]interface{}); ok {
			if model, ok := m["model"].(string); ok {
				requestModel = model
			}
		}
	}
	if requestModel == "" {
		requestModel = c.GetString(ctxkey.RequestModel)
	}

	c.Set(ctxkey.OriginalModel, requestModel)
	c.Set(ctxkey.BaseURL, "")
	c.Set(ctxkey.ModelMapping, map[string]string{})
	c.Set(ctxkey.Config, struct{}{})

	// Mark the source so the relay controller knows not to look for a real API key
	c.Header("X-Sub2API-Provider", c.GetString("sub2api_provider"))
}
