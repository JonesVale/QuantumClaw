package common

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
	"gorm.io/gorm"
)

type ConnectionStatus string

const (
	StatusConnected    ConnectionStatus = "connected"
	StatusDisconnected ConnectionStatus = "disconnected"
	StatusReconnecting ConnectionStatus = "reconnecting"
)

type DBHealthChecker struct {
	mu           sync.RWMutex
	status       ConnectionStatus
	lastCheck    time.Time
	healthFn     func() error
	onDisconnect func()
	onReconnect  func()
}

func NewDBHealthChecker(healthFn func() error) *DBHealthChecker {
	return &DBHealthChecker{
		status:    StatusConnected,
		healthFn:  healthFn,
		lastCheck: time.Now(),
	}
}

func (h *DBHealthChecker) SetOnDisconnect(fn func()) {
	h.mu.Lock()
	h.onDisconnect = fn
	h.mu.Unlock()
}

func (h *DBHealthChecker) SetOnReconnect(fn func()) {
	h.mu.Lock()
	h.onReconnect = fn
	h.mu.Unlock()
}

func (h *DBHealthChecker) Check() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.lastCheck = time.Now()

	err := h.healthFn()
	if err != nil {
		if h.status != StatusDisconnected {
			h.status = StatusDisconnected
			logger.SysError("Database connection lost: " + err.Error())
			if h.onDisconnect != nil {
				go h.onDisconnect()
			}
		}
		return err
	}

	if h.status == StatusDisconnected {
		h.status = StatusConnected
		logger.SysLog("Database connection restored")
		if h.onReconnect != nil {
			go h.onReconnect()
		}
	}

	return nil
}

func (h *DBHealthChecker) GetStatus() ConnectionStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.status
}

func (h *DBHealthChecker) GetLastCheck() time.Time {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.lastCheck
}

func StartDBHealthChecker(db *gorm.DB, interval time.Duration) *DBHealthChecker {
	checker := NewDBHealthChecker(func() error {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Ping()
	})

	checker.SetOnReconnect(func() {
		logger.SysLog("Database health check: reconnection successful")
	})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			if err := checker.Check(); err != nil {
				logger.SysError("Database health check failed: " + err.Error())
			}
		}
	}()

	return checker
}

type RedisHealthChecker struct {
	mu           sync.RWMutex
	status       ConnectionStatus
	lastCheck    time.Time
	healthFn     func() error
	onDisconnect func()
	onReconnect  func()
}

func NewRedisHealthChecker(healthFn func() error) *RedisHealthChecker {
	return &RedisHealthChecker{
		status:    StatusConnected,
		healthFn:  healthFn,
		lastCheck: time.Now(),
	}
}

func (h *RedisHealthChecker) SetOnDisconnect(fn func()) {
	h.mu.Lock()
	h.onDisconnect = fn
	h.mu.Unlock()
}

func (h *RedisHealthChecker) SetOnReconnect(fn func()) {
	h.mu.Lock()
	h.onReconnect = fn
	h.mu.Unlock()
}

func (h *RedisHealthChecker) Check() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.lastCheck = time.Now()

	err := h.healthFn()
	if err != nil {
		if h.status != StatusDisconnected {
			h.status = StatusDisconnected
			logger.SysError("Redis connection lost: " + err.Error())
			if h.onDisconnect != nil {
				go h.onDisconnect()
			}
		}
		return err
	}

	if h.status == StatusDisconnected {
		h.status = StatusConnected
		logger.SysLog("Redis connection restored")
		if h.onReconnect != nil {
			go h.onReconnect()
		}
	}

	return nil
}

func (h *RedisHealthChecker) GetStatus() ConnectionStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.status
}

func StartRedisHealthChecker(redisPingFn func() error, interval time.Duration) *RedisHealthChecker {
	checker := NewRedisHealthChecker(redisPingFn)

	checker.SetOnReconnect(func() {
		logger.SysLog("Redis health check: reconnection successful")
	})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			if err := checker.Check(); err != nil {
				logger.SysError("Redis health check failed: " + err.Error())
			}
		}
	}()

	return checker
}

type RecoveryManager struct {
	mu                sync.RWMutex
	dbHealthChecker   *DBHealthChecker
	redisHealthChecker *RedisHealthChecker
	recovering        bool
}

var globalRecoveryManager = &RecoveryManager{}

func GetRecoveryManager() *RecoveryManager {
	return globalRecoveryManager
}

func (m *RecoveryManager) InitDBRecovery(db *gorm.DB) {
	m.dbHealthChecker = StartDBHealthChecker(db, 30*time.Second)
	logger.SysLog("Database health checker started")
}

func (m *RecoveryManager) InitRedisRecovery(redisPingFn func() error) {
	m.redisHealthChecker = StartRedisHealthChecker(redisPingFn, 15*time.Second)
	logger.SysLog("Redis health checker started")
}

func (m *RecoveryManager) IsDBConnected() bool {
	if m.dbHealthChecker == nil {
		return true
	}
	return m.dbHealthChecker.GetStatus() == StatusConnected
}

func (m *RecoveryManager) IsRedisConnected() bool {
	if m.redisHealthChecker == nil {
		return true
	}
	return m.redisHealthChecker.GetStatus() == StatusConnected
}

func (m *RecoveryManager) ForceReconnect() error {
	m.mu.Lock()
	if m.recovering {
		m.mu.Unlock()
		return fmt.Errorf("recovery already in progress")
	}
	m.recovering = true
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		m.recovering = false
		m.mu.Unlock()
	}()

	logger.SysLog("Starting forced reconnection...")

	if m.dbHealthChecker != nil {
		if err := m.dbHealthChecker.Check(); err != nil {
			logger.SysError("Database reconnection failed: " + err.Error())
		}
	}

	if m.redisHealthChecker != nil {
		if err := m.redisHealthChecker.Check(); err != nil {
			logger.SysError("Redis reconnection failed: " + err.Error())
		}
	}

	return nil
}

type GracefulShutdown struct {
	mu          sync.RWMutex
	shutdowning bool
	cleanupFns  []func(context.Context)
}

var globalShutdown = &GracefulShutdown{}

func (s *GracefulShutdown) RegisterCleanup(fn func(context.Context)) {
	s.mu.Lock()
	s.cleanupFns = append(s.cleanupFns, fn)
	s.mu.Unlock()
}

func (s *GracefulShutdown) IsShuttingDown() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.shutdowning
}

func (s *GracefulShutdown) BeginShutdown() {
	s.mu.Lock()
	s.shutdowning = true
	s.mu.Unlock()

	logger.SysLog("Graceful shutdown initiated")
}

func (s *GracefulShutdown) Cleanup(ctx context.Context) {
	s.mu.RLock()
	fns := s.cleanupFns
	s.mu.RUnlock()

	for _, fn := range fns {
		go fn(ctx)
	}
}

func GetGracefulShutdown() *GracefulShutdown {
	return globalShutdown
}