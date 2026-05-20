package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/logger"
)

var (
	loginAttempts sync.Map
)

type accountMutexPool struct {
	pool sync.Map
}

func (p *accountMutexPool) Lock(key string) func() {
	actual, _ := p.pool.LoadOrStore(key, &sync.Mutex{})
	mu := actual.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}

var accountMutexes accountMutexPool

func loginLockKey(ip, username string) string {
	return ip + ":" + strings.ToLower(username)
}

type loginAttempt struct {
	mu               sync.Mutex
	count            int
	firstFail        time.Time
	consecutiveFails int
	lockedUntil      time.Time
}

// LoginRateLimit limits login/register attempts per account (user+ip)
// Rate: 10 attempts per 15-minute window
// After 5 consecutive failures, account-user is locked for 15 minutes
// Uses c.GetRawData() (cached) to read username without consuming the body
func LoginRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		// Read username from body without consuming it
		username := "unknown"
		rawData, err := c.GetRawData()
		if err == nil && len(rawData) > 0 {
			var body struct {
				Username string `json:"username"`
			}
			if json.Unmarshal(rawData, &body) == nil && body.Username != "" {
				username = strings.ToLower(body.Username)
			}
			// Re-store the body for downstream handlers
			c.Request.Body = io.NopCloser(bytes.NewBuffer(rawData))
		}

		lockKey := loginLockKey(ip, username)

		unlock := accountMutexes.Lock(lockKey)

		val, _ := loginAttempts.LoadOrStore(lockKey, &loginAttempt{
			firstFail: now,
		})

		attempt := val.(*loginAttempt)
		attempt.mu.Lock()

		// Check if locked
		if !attempt.lockedUntil.IsZero() && now.Before(attempt.lockedUntil) {
			remaining := time.Until(attempt.lockedUntil).Round(time.Second).String()
			attempt.mu.Unlock()
			unlock()
			logger.Warn(c.Request.Context(), "rate limited (locked) login: "+lockKey)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"message":      "登录锁定，请 " + remaining + " 后重试，或使用紧急重置",
				"success":      false,
				"locked":       true,
				"locked_until": attempt.lockedUntil.Unix(),
				"hint":         "Use POST /api/password/emergency-reset with EMERGENCY_RESET_TOKEN to reset",
			})
			return
		}

		// Auto-unlock after lockout period
		if !attempt.lockedUntil.IsZero() && !now.Before(attempt.lockedUntil) {
			attempt.lockedUntil = time.Time{}
			attempt.consecutiveFails = 0
			attempt.count = 0
		}

		// Reset window after 15 minutes
		if now.Sub(attempt.firstFail) > 15*time.Minute {
			attempt.count = 0
			attempt.firstFail = now
		}

		attempt.count++

		if attempt.count > 10 {
			attempt.mu.Unlock()
			unlock()
			logger.Warn(c.Request.Context(), "rate limited (window exceeded) login: "+lockKey)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"message": "too many requests, please try again later",
				"success": false,
			})
			return
		}

		attempt.mu.Unlock()
		unlock()

		c.Next()

		// Post-response: track failures
		unlock2 := accountMutexes.Lock(lockKey)
		attempt.mu.Lock()

		if c.Writer.Status() != http.StatusOK {
			attempt.consecutiveFails++
			if attempt.consecutiveFails >= 5 {
				attempt.lockedUntil = now.Add(15 * time.Minute)
				logger.Warn(c.Request.Context(), "login locked (5 consecutive failures): "+lockKey)
			}
		} else {
			// Login success, reset counters
			attempt.count = 0
			attempt.consecutiveFails = 0
			attempt.lockedUntil = time.Time{}
		}

		attempt.mu.Unlock()
		unlock2()
	}
}
