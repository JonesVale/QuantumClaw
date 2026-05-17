package middleware

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/ctxkey"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/setting/operation_setting"
)

var (
	inMemoryModelRateLimiter common.InMemoryRateLimiter
	modelRateLimitTimeFormat = "2006-01-02T15:04:05.000Z"
)

// modelMatches checks if modelName matches the pattern.
// Supports exact match and glob-style suffix match (e.g., "gpt-4*").
func modelMatches(modelName, pattern string) bool {
	if pattern == "" {
		// Empty pattern means catch-all
		return true
	}
	if pattern == modelName {
		return true
	}
	// Suffix glob: "gpt-4*" matches "gpt-4", "gpt-4-turbo", etc.
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(modelName, prefix)
	}
	// Try regex match
	matched, err := regexp.MatchString(pattern, modelName)
	if err == nil && matched {
		return true
	}
	return false
}

// modelRateLimitRedis performs rate limit check using Redis.
func modelRateLimitRedis(key string, maxRequests int, duration int64) bool {
	ctx := context.Background()
	rdb := common.RDB
	redisKey := "model-ratelimit:" + key

	listLength, err := rdb.LLen(ctx, redisKey).Result()
	if err != nil {
		logger.SysError("model rate limit redis error: " + err.Error())
		// On Redis error, allow the request (fail open)
		return true
	}

	now := time.Now()
	nowStr := now.Format(modelRateLimitTimeFormat)

	if int(listLength) < maxRequests {
		rdb.LPush(ctx, redisKey, nowStr)
		rdb.Expire(ctx, redisKey, config.RateLimitKeyExpirationDuration)
		return true
	}

	oldTimeStr, err := rdb.LIndex(ctx, redisKey, -1).Result()
	if err != nil {
		return true
	}
	oldTime, err := time.Parse(modelRateLimitTimeFormat, oldTimeStr)
	if err != nil {
		return true
	}

	if int64(now.Sub(oldTime).Seconds()) < duration {
		rdb.Expire(ctx, redisKey, config.RateLimitKeyExpirationDuration)
		return false
	}

	rdb.LPush(ctx, redisKey, nowStr)
	rdb.LTrim(ctx, redisKey, 0, int64(maxRequests-1))
	rdb.Expire(ctx, redisKey, config.RateLimitKeyExpirationDuration)
	return true
}

// memoryModelRateLimiter performs rate limit check using in-memory limiter.
func memoryModelRateLimiter(key string, maxRequests int, duration int64) bool {
	return inMemoryModelRateLimiter.Request(key, maxRequests, duration)
}

// ModelRateLimit checks per-model rate limits before distributing to channels.
// It must run after TokenAuth (which sets RequestModel) and before Distribute().
func ModelRateLimit() func(c *gin.Context) {
	if config.DebugEnabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// Initialize in-memory rate limiter if Redis is not enabled
	if !common.RedisEnabled {
		inMemoryModelRateLimiter.Init(config.RateLimitKeyExpirationDuration)
	}

	return func(c *gin.Context) {
		setting := operation_setting.GetModelRateLimitSetting()
		if !setting.Enabled || len(setting.Rules) == 0 {
			c.Next()
			return
		}

		requestModel := c.GetString(ctxkey.RequestModel)
		if requestModel == "" {
			c.Next()
			return
		}

		// Fetch user group (same as Distribute() does)
		userId := c.GetInt(ctxkey.Id)
		group := c.GetString(ctxkey.Group)
		if group == "" && userId != 0 {
			group, _ = model.CacheGetUserGroup(userId)
		}

		// Find the first matching rule
		for _, rule := range setting.Rules {
			if !modelMatches(requestModel, rule.Model) {
				continue
			}

			// Build a rate limit key: "model-ratelimit:<model>:<group>"
			key := rule.Model + ":" + group

			// Check rate limit
			var allowed bool
			if common.RedisEnabled {
				allowed = modelRateLimitRedis(key, rule.MaxRequests, rule.Duration)
			} else {
				allowed = memoryModelRateLimiter(key, rule.MaxRequests, rule.Duration)
			}

			if !allowed {
				logger.Warnf(c.Request.Context(), "model rate limit exceeded: model=%s, group=%s, limit=%d/%ds",
					requestModel, group, rule.MaxRequests, rule.Duration)
				c.Status(http.StatusTooManyRequests)
				c.Abort()
				return
			}

			// First matching rule wins
			break
		}

		c.Next()
	}
}

// SetupModelRateLimiterFromOption parses a ModelRateLimitSetting from JSON
// and applies it. This is intended to be called when the admin updates the setting.
func SetupModelRateLimiterFromOption(jsonData string) error {
	setting, err := operation_setting.ParseModelRateLimitSetting(jsonData)
	if err != nil {
		return fmt.Errorf("failed to parse model rate limit setting: %w", err)
	}
	operation_setting.SetModelRateLimitSetting(setting)
	return nil
}
