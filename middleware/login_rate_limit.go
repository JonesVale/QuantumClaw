package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/logger"
)

// bodyCaptureWriter captures the response body for inspection in middleware
type bodyCaptureWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyCaptureWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func getResponseBody(c *gin.Context) []byte {
	if bw, ok := c.Writer.(*bodyCaptureWriter); ok {
		return bw.body.Bytes()
	}
	return nil
}

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

// ClearAllLoginLocks 清除所有登录锁定状态（由紧急重置触发时调用）
func ClearAllLoginLocks() {
	loginAttempts.Range(func(key, value interface{}) bool {
		attempt := value.(*loginAttempt)
		attempt.mu.Lock()
		attempt.lockedUntil = time.Time{}
		attempt.consecutiveFails = 0
		attempt.count = 0
		attempt.mu.Unlock()
		return true
	})
}

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
// After 3 consecutive failures, account-user is locked for 24 hours (daily reset)
// Uses c.GetRawData() (cached) to read username without consuming the body
func LoginRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		// 仅登录路径触发连续失败锁定，注册不计入
		isLoginPath := strings.HasSuffix(c.Request.URL.Path, "/login")

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
			remaining := attempt.lockedUntil.Sub(now).Round(time.Second)
			attempt.mu.Unlock()
			unlock()
			logger.Warn(c.Request.Context(), "rate limited (locked) login: "+lockKey)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"message":      fmt.Sprintf("密码错误次数过多，账号已锁定(剩余%s)。请稍后再试", remaining.String()),
				"success":      false,
				"locked":       true,
				"locked_until": attempt.lockedUntil.Unix(),
				"hint":         "POST /api/password/emergency-reset",
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

		// Wrap response writer to capture body
		blw := &bodyCaptureWriter{ResponseWriter: c.Writer, body: bytes.NewBuffer(nil)}
		c.Writer = blw

		c.Next()

		// Post-response: track failures（仅登录路径触发锁定）
		unlock2 := accountMutexes.Lock(lockKey)
		attempt.mu.Lock()

		if isLoginPath {
			// Check response body for success field
			isFailure := true // default: treat as failure
			respBody := getResponseBody(c)
			if len(respBody) > 0 {
				var resp struct {
					Success bool `json:"success"`
				}
				if json.Unmarshal(respBody, &resp) == nil {
					isFailure = !resp.Success
				}
			}

			if isFailure {
				attempt.consecutiveFails++
				logger.Warn(c.Request.Context(), fmt.Sprintf("login failed (attempt %d): %s", attempt.consecutiveFails, lockKey))
				// 阶梯锁：对不同的连续失败次数施加不同的锁定时间
				// 3次→5min  5次→15min  8次→1h  10次+→24h
				var lockDuration time.Duration
				switch {
				case attempt.consecutiveFails >= 10:
					lockDuration = 24 * time.Hour
				case attempt.consecutiveFails >= 8:
					lockDuration = 1 * time.Hour
				case attempt.consecutiveFails >= 5:
					lockDuration = 15 * time.Minute
				case attempt.consecutiveFails >= 3:
					lockDuration = 5 * time.Minute
				}
				if lockDuration > 0 {
					attempt.lockedUntil = now.Add(lockDuration)
					logger.Warn(c.Request.Context(), fmt.Sprintf("login locked (%d consecutive fails, %v): %s", attempt.consecutiveFails, lockDuration, lockKey))
				}
			} else {
				// Login success, reset counters
				attempt.count = 0
				attempt.consecutiveFails = 0
				attempt.lockedUntil = time.Time{}
			}
		}

		attempt.mu.Unlock()
		unlock2()
	}
}
