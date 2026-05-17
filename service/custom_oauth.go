package service

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/quantumclaw/quantumclaw/common/logger"
)

// CustomOAuthProvider defines a dynamically loaded OAuth provider.
type CustomOAuthProvider struct {
	Name         string `json:"name"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	AuthUrl      string `json:"auth_url"`
	TokenUrl     string `json:"token_url"`
	UserInfoUrl  string `json:"user_info_url"`
	Scopes       string `json:"scopes"`
	Enabled      bool   `json:"enabled"`
}

var (
	customProviders   []CustomOAuthProvider
	customProvidersMu sync.RWMutex
)

// LoadCustomOAuthProviders loads custom OAuth providers from JSON config.
// Supports: environment variable CUSTOM_OAUTH_PROVIDERS or a JSON file.
func LoadCustomOAuthProviders() {
	providersJSON := os.Getenv("CUSTOM_OAUTH_PROVIDERS")
	if providersJSON == "" {
		return
	}

	var providers []CustomOAuthProvider
	if err := json.Unmarshal([]byte(providersJSON), &providers); err != nil {
		logger.SysError("failed to parse CUSTOM_OAUTH_PROVIDERS: " + err.Error())
		return
	}

	customProvidersMu.Lock()
	customProviders = providers
	customProvidersMu.Unlock()

	logger.SysLog(fmt.Sprintf("loaded %d custom OAuth providers", len(providers)))
}

// GetCustomOAuthProviders returns the list of loaded custom OAuth providers.
func GetCustomOAuthProviders() []CustomOAuthProvider {
	customProvidersMu.RLock()
	defer customProvidersMu.RUnlock()
	cp := make([]CustomOAuthProvider, len(customProviders))
	copy(cp, customProviders)
	return cp
}

// GetCustomOAuthProviderByName finds a custom provider by name.
func GetCustomOAuthProviderByName(name string) *CustomOAuthProvider {
	customProvidersMu.RLock()
	defer customProvidersMu.RUnlock()
	for _, p := range customProviders {
		if p.Name == name && p.Enabled {
			return &p
		}
	}
	return nil
}
