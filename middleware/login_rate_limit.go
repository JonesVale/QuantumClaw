package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/logger"
)

var (
	loginAttempts sync.Map
)

type loginAttempt struct {
	count           int
	firstFail       time.Time
	consecutiveFails int
	lockedUntil     time.Time
}

// LoginRateLimit limits login/register attempts per IP
// Rate: 10 attempts per 15-minute window
// After 5 consecutive failures, account is locked for 15 minutes (P3 enhancement)
func LoginRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		val, _ := loginAttempts.LoadOrStore(ip, &loginAttempt{
			count:           0,
			firstFail:       now,
			consecutiveFails: 0,
		})

		attempt := val.(*loginAttempt)

		// Check if currently locked out
		if !attempt.lockedUntil.IsZero() && now.Before(attempt.lockedUntil) {
			logger.Warn(c.Request.Context(), "rate limited (locked) login from IP: "+ip)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"message": "请求过于频繁，请稍后再试",
				"success": false,
			})
			return
		}

		// Reset lock if lock period expired
		if !attempt.lockedUntil.IsZero() && !now.Before(attempt.lockedUntil) {
			attempt.lockedUntil = time.Time{}
			attempt.consecutiveFails = 0
			attempt.count = 0
		}

		// Window-based: reset if 15-minute window expired
		if now.Sub(attempt.firstFail) > 15*time.Minute {
			attempt.count = 0
			attempt.firstFail = now
		}

		attempt.count++

		if attempt.count > 10 {
			logger.Warn(c.Request.Context(), "rate limited login from IP: "+ip)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"message": "请求过于频繁，请稍后再试",
				"success": false,
			})
			return
		}

		c.Next()

		// Track failures after the response
		if c.Writer.Status() != http.StatusOK {
			attempt.consecutiveFails++
			if attempt.consecutiveFails >= 5 {
				attempt.lockedUntil = now.Add(15 * time.Minute)
				logger.Warn(c.Request.Context(), "login locked for IP due to consecutive failures: "+ip)
			}
		} else {
			// Successful login - reset counters
			attempt.count = 0
			attempt.consecutiveFails = 0
			attempt.lockedUntil = time.Time{}
		}
	}
}
