package web_shared

import (
	"strings"

	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/service"
)

// ── Auth Helpers ──

// ResolveCredential finds the best Sub2APICredential for a user + provider combination.
// provider is the schema provider prefix (e.g. "chatgpt" matches "chatgpt_plus", "chatgpt_pro", etc.)
// Returns the decrypted token value (AES-GCM decrypted), and whether one was found.
func ResolveCredential(userId int, provider string) (string, bool, error) {
	creds, err := model.ListSub2CredentialsByProviderPrefix(userId, provider)
	if err != nil {
		return "", false, err
	}
	if len(creds) == 0 {
		return "", false, nil
	}

	// Pick the most recently updated credential
	best := &creds[0]
	for i := range creds {
		if creds[i].UpdatedTime > best.UpdatedTime {
			best = &creds[i]
		}
	}

	// Decrypt the token
	token, err := service.Sub2API.DecryptToken(best.Token)
	if err != nil {
		return "", false, err
	}
	return token, true, nil
}

// ResolveCredentialByModel resolves a credential by mapping the API model name to a provider
// through schema model mappings.
func ResolveCredentialByModel(userId int, apiModel string) (string, string, bool, error) {
	// Find all active schemas
	schemas, err := model.ListSub2APISchemas()
	if err != nil {
		return "", "", false, err
	}

	// Find which provider can serve this model
	var matchedProvider string
	for _, s := range schemas {
		if s.Status != 1 {
			continue
		}
		mapping, err := s.ParseModelMapping()
		if err != nil {
			continue
		}
		if _, ok := mapping[apiModel]; ok {
			matchedProvider = s.Provider
			break
		}
		if _, ok := mapping["*"]; ok {
			matchedProvider = s.Provider
			break
		}
		if strings.HasPrefix(apiModel, s.Provider+"-") || strings.HasPrefix(apiModel, s.Provider+"/") {
			matchedProvider = s.Provider
			break
		}
	}

	if matchedProvider == "" {
		return "", "", false, nil
	}

	token, found, err := ResolveCredential(userId, matchedProvider)
	return token, matchedProvider, found, err
}
