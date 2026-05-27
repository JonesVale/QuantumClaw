package web_shared

import (
	"sync"
	"time"

	"github.com/quantumclaw/quantumclaw/model"
)

// ── Circuit Breaker ──

// CircuitState tracks the health of a specific schema version.
type CircuitState struct {
	SchemaID        int
	Provider        string
	Version         int
	Failures        int
	LastFailureAt   time.Time
	IsOpen          bool
	OpenedAt        time.Time
	HalfOpenAfter   time.Duration // default 5min
	ConsecutiveFails int
}

// CircuitBreaker manages per-schema health tracking and automatic fallback.
type CircuitBreaker struct {
	mu       sync.RWMutex
	circuits map[int]*CircuitState // keyed by schema ID
}

var DefaultCircuitBreaker = &CircuitBreaker{
	circuits: make(map[int]*CircuitState),
}

// RecordFailure records a failure for the given schema.
// Returns true if the circuit should be opened (auto-fallback needed).
func (cb *CircuitBreaker) RecordFailure(schemaID int) bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	state, ok := cb.circuits[schemaID]
	if !ok {
		state = &CircuitState{
			SchemaID:      schemaID,
			HalfOpenAfter: 5 * time.Minute,
		}
		cb.circuits[schemaID] = state
	}

	state.Failures++
	state.ConsecutiveFails++
	state.LastFailureAt = time.Now()

	// Open circuit after 3 consecutive failures
	if state.ConsecutiveFails >= 3 && !state.IsOpen {
		state.IsOpen = true
		state.OpenedAt = time.Now()

		// Mark schema as broken in DB
		s, err := model.GetSub2APISchema(schemaID)
		if err == nil && s.Status == 1 {
			s.Status = 4 // broken
			s.LastError = "circuit opened after " + itoa(state.ConsecutiveFails) + " consecutive failures"
			s.LastHealthAt = time.Now().UnixMilli()
			_ = model.UpdateSub2APISchema(s)
		}
		return true // signal fallback
	}

	return false
}

// RecordSuccess clears failure counts for a recovering schema.
func (cb *CircuitBreaker) RecordSuccess(schemaID int) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	state, ok := cb.circuits[schemaID]
	if !ok {
		return
	}
	state.ConsecutiveFails = 0
	state.Failures = 0

	if state.IsOpen && time.Since(state.OpenedAt) > state.HalfOpenAfter {
		// Auto-recover: mark as active again
		state.IsOpen = false

		s, err := model.GetSub2APISchema(schemaID)
		if err == nil && s.Status == 4 {
			s.Status = 1 // restore active
			s.LastError = "auto-recovered"
			_ = model.UpdateSub2APISchema(s)
		}
	}
}

// IsCircuitOpen checks if the circuit for a schema is open.
func (cb *CircuitBreaker) IsCircuitOpen(schemaID int) bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	state, ok := cb.circuits[schemaID]
	if !ok {
		return false
	}
	if !state.IsOpen {
		return false
	}
	// Auto half-open after timeout
	if time.Since(state.OpenedAt) > state.HalfOpenAfter {
		return false // allow a probe request
	}
	return true
}

// ── Schema Fallback ──

// FindFallbackSchema finds the next available schema version for a provider when
// the current one has failed.
func FindFallbackSchema(provider string, failedVersion int) (*model.Sub2APISchema, error) {
	schemas, err := model.ListActiveSchemasByProvider(provider)
	if err != nil {
		return nil, err
	}

	var best *model.Sub2APISchema
	for _, s := range schemas {
		if s.Version == failedVersion {
			continue // skip the failed version
		}
		if DefaultCircuitBreaker.IsCircuitOpen(s.Id) {
			continue // skip open circuits
		}
		if best == nil || s.Priority > best.Priority || s.Version > best.Version {
			best = &s
		}
	}
	return best, nil
}

// helper: int to string
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
