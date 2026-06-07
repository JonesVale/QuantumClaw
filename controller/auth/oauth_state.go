package auth

import (
	"sync"

	"github.com/quantumclaw/quantumclaw/common/random"
)

// oauthStateStore supports concurrent OAuth flows with multiple pending states.
// Previously, state was stored in a single session key, causing conflicts when
// users opened multiple OAuth tabs simultaneously.
var oauthStateStore = struct {
	mu     sync.RWMutex
	states map[string]bool // state → valid
}{
	states: make(map[string]bool),
}

// GenerateOAuthState creates a new random state string and stores it.
// Returns the state to be sent to the OAuth provider.
func GenerateOAuthState() string {
	state := random.GetRandomString(12)
	oauthStateStore.mu.Lock()
	oauthStateStore.states[state] = true
	oauthStateStore.mu.Unlock()
	return state
}

// VerifyOAuthState checks if the state is valid and removes it (one-time use).
// Returns true only if state was previously generated and not yet consumed.
func VerifyOAuthState(state string) bool {
	if state == "" {
		return false
	}
	oauthStateStore.mu.Lock()
	defer oauthStateStore.mu.Unlock()
	if oauthStateStore.states[state] {
		delete(oauthStateStore.states, state) // consume the state
		return true
	}
	return false
}
