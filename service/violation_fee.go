package service

import (
	"fmt"
	"sync"

	"github.com/quantumclaw/quantumclaw/common/logger"
)

// ViolationFeeConfig holds configuration for a violation fee entry.
type ViolationFeeConfig struct {
	Model    string
	Fee      int64
	MinCount int
}

var (
	violationFees []ViolationFeeConfig
	vfMu          sync.RWMutex
)

// RegisterViolationFee registers a violation fee configuration.
func RegisterViolationFee(cfg ViolationFeeConfig) {
	vfMu.Lock()
	defer vfMu.Unlock()
	violationFees = append(violationFees, cfg)
	logger.SysLog(fmt.Sprintf("violation fee registered: model=%s fee=%d", cfg.Model, cfg.Fee))
}

// GetViolationFee returns the violation fee for the given model.
func GetViolationFee(model string) int64 {
	vfMu.RLock()
	defer vfMu.RUnlock()
	for _, cfg := range violationFees {
		if cfg.Model == model || cfg.Model == "*" {
			return cfg.Fee
		}
	}
	return 0
}
