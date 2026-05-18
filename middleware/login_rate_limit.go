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

// ipMutexPool 为每个 IP 提供独立的互斥锁，避免热 IP 影响冷 IP
type ipMutexPool struct {
	pool sync.Map
}

func (p *ipMutexPool) Lock(ip string) func() {
	// 使用 *sync.Mutex 避免值拷贝
	actual, _ := p.pool.LoadOrStore(ip, &sync.Mutex{})
	mu := actual.(*sync.Mutex)
	mu.Lock()
	// 返回解锁函数
	return mu.Unlock
}

var ipMutexes ipMutexPool

type loginAttempt struct {
	mu               sync.Mutex
	count            int
	firstFail        time.Time
	consecutiveFails int
	lockedUntil      time.Time
}

// LoginRateLimit limits login/register attempts per IP
// Rate: 10 attempts per 15-minute window
// After 5 consecutive failures, account is locked for 15 minutes
func LoginRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		unlock := ipMutexes.Lock(ip)

		val, _ := loginAttempts.LoadOrStore(ip, &loginAttempt{
			firstFail: now,
		})

		attempt := val.(*loginAttempt)

		// 上锁保护结构体字段的并发读写
		attempt.mu.Lock()

		// 检查是否被锁定
		if !attempt.lockedUntil.IsZero() && now.Before(attempt.lockedUntil) {
			attempt.mu.Unlock()
			unlock()
			logger.Warn(c.Request.Context(), "rate limited (locked) login from IP: "+ip)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"message": "请求过于频繁，请稍后再试",
				"success": false,
			})
			return
		}

		// 锁定到期后自动解锁
		if !attempt.lockedUntil.IsZero() && !now.Before(attempt.lockedUntil) {
			attempt.lockedUntil = time.Time{}
			attempt.consecutiveFails = 0
			attempt.count = 0
		}

		// 窗口过期后重置
		if now.Sub(attempt.firstFail) > 15*time.Minute {
			attempt.count = 0
			attempt.firstFail = now
		}

		attempt.count++

		if attempt.count > 10 {
			attempt.mu.Unlock()
			unlock()
			logger.Warn(c.Request.Context(), "rate limited login from IP: "+ip)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"message": "请求过于频繁，请稍后再试",
				"success": false,
			})
			return
		}

		attempt.mu.Unlock()
		unlock()

		c.Next()

		// 响应后记录失败（重新上锁）
		unlock2 := ipMutexes.Lock(ip)
		attempt.mu.Lock()

		if c.Writer.Status() != http.StatusOK {
			attempt.consecutiveFails++
			if attempt.consecutiveFails >= 5 {
				attempt.lockedUntil = now.Add(15 * time.Minute)
				logger.Warn(c.Request.Context(), "login locked for IP due to consecutive failures: "+ip)
			}
		} else {
			// 登录成功，重置计数器
			attempt.count = 0
			attempt.consecutiveFails = 0
			attempt.lockedUntil = time.Time{}
		}

		attempt.mu.Unlock()
		unlock2()
	}
}
