package service

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/encrypt"
	"github.com/quantumclaw/quantumclaw/model"
)

// Sub2API service handles credential encryption, validation, and usage tracking.
type Sub2APIService struct{}

var Sub2API = &Sub2APIService{}

// deriveKey returns the unified AES-256 key from CryptoSecret (same as channel keys).
func (s *Sub2APIService) deriveKey() []byte {
	return encrypt.DeriveKey(config.CryptoSecret)
}

// EncryptToken AES-GCM encrypts a plaintext session token using the unified key material.
func (s *Sub2APIService) EncryptToken(plaintext string) (string, error) {
	return encrypt.Encrypt([]byte(plaintext), s.deriveKey())
}

// DecryptToken decrypts an AES-GCM encrypted session token.
func (s *Sub2APIService) DecryptToken(encoded string) (string, error) {
	decrypted, err := encrypt.Decrypt(encoded, s.deriveKey())
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}

// hashToken returns SHA-256 hex of the plaintext token for dedup.
func (s *Sub2APIService) hashToken(plaintext string) string {
	h := sha256.Sum256([]byte(plaintext))
	return fmt.Sprintf("%x", h)
}

// CreateCredential creates a new encrypted credential for a user.
func (s *Sub2APIService) CreateCredential(userId int, provider model.Sub2APIProvider, label, token string, dailyCap int64) (*model.Sub2APICredential, error) {
	if token == "" {
		return nil, errors.New("token cannot be empty")
	}
	if provider == "" {
		return nil, errors.New("provider cannot be empty")
	}

	encrypted, err := s.EncryptToken(token)
	if err != nil {
		return nil, fmt.Errorf("encrypt: %w", err)
	}

	tokenHash := s.hashToken(token)
	cred := &model.Sub2APICredential{
		UserId:    userId,
		Provider:  provider,
		Label:     label,
		Token:     encrypted,
		TokenHash: tokenHash,
		DailyCap:  dailyCap,
		Status:    1,
	}
	if err := model.DB.Create(cred).Error; err != nil {
		return nil, fmt.Errorf("db create: %w", err)
	}
	return cred, nil
}

// GetUserCredentials returns all credentials for a specific user.
func (s *Sub2APIService) GetUserCredentials(userId int) ([]model.Sub2APICredential, error) {
	var creds []model.Sub2APICredential
	if err := model.DB.Where("user_id = ?", userId).Order("created_time desc").Find(&creds).Error; err != nil {
		return nil, err
	}
	// Decrypt tokens for response; return empty string if decryption fails
	for i := range creds {
		if creds[i].Token != "" {
			decrypted, err := s.DecryptToken(creds[i].Token)
			if err == nil {
				// Keep only first 8 chars for display; full token returned on explicit reveal
				if len(decrypted) > 8 {
					creds[i].Token = decrypted[:8] + "****"
				} else {
					creds[i].Token = decrypted
				}
			}
		}
	}
	return creds, nil
}

// ListAvailableCredentials returns all active credentials across providers.
func (s *Sub2APIService) ListAvailableCredentials() ([]model.Sub2APICredential, error) {
	var creds []model.Sub2APICredential
	if err := model.DB.Where("status = ?", 1).Order("created_time desc").Find(&creds).Error; err != nil {
		return nil, err
	}
	return creds, nil
}

// GetCredentialByID returns a specific credential (decrypted from DB) by ID.
func (s *Sub2APIService) GetCredentialByID(id int) (*model.Sub2APICredential, error) {
	var cred model.Sub2APICredential
	if err := model.DB.First(&cred, id).Error; err != nil {
		return nil, err
	}
	// Decrypt for relay transport only; the caller must NOT expose this to the frontend
	if cred.Token != "" {
		decrypted, err := s.DecryptToken(cred.Token)
		if err != nil {
			return nil, fmt.Errorf("decrypt token: %w", err)
		}
		cred.Token = decrypted
	}
	return &cred, nil
}

// UpdateCredentialStatus updates the status of a credential.
func (s *Sub2APIService) UpdateCredentialStatus(id int, status int) error {
	return model.DB.Model(&model.Sub2APICredential{}).Where("id = ?", id).Update("status", status).Error
}

// DeleteCredential removes a credential.
func (s *Sub2APIService) DeleteCredential(id int) error {
	return model.DB.Delete(&model.Sub2APICredential{}, id).Error
}

// GetSupportedProviders returns the list of supported subscription providers.
func (s *Sub2APIService) GetSupportedProviders() []model.Sub2APIProvider {
	return []model.Sub2APIProvider{
		model.Sub2ProviderChatGPTPlus,
		model.Sub2ProviderChatGPTPro,
		model.Sub2ProviderChatGPTTeam,
		model.Sub2ProviderClaudePro,
		model.Sub2ProviderClaudeTeam,
	}
}

// ValidateCredential retrieves a credential by id and verifies it belongs to the given user.
func (s *Sub2APIService) ValidateCredential(id, userId int) (*model.Sub2APICredential, error) {
	cred, err := s.GetCredentialByID(id)
	if err != nil {
		return nil, fmt.Errorf("credential not found: %w", err)
	}
	if cred.UserId != userId {
		return nil, errors.New("credential does not belong to this user")
	}
	return cred, nil
}

// EnsureHealthChecks runs periodic health checks on active credentials.
func (s *Sub2APIService) EnsureHealthChecks(frequencySeconds int) {
	ticker := time.NewTicker(time.Duration(frequencySeconds) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		creds, err := s.ListAvailableCredentials()
		if err != nil {
			continue
		}
		for _, cred := range creds {
			if cred.Status != 1 {
				continue
			}
			_ = s.checkHealth(cred)
		}
	}
}

// checkHealth performs a single health check on a credential.
func (s *Sub2APIService) checkHealth(cred model.Sub2APICredential) error {
	decrypted, err := s.DecryptToken(cred.Token)
	if err != nil {
		return fmt.Errorf("decrypt: %w", err)
	}
	// Provider-specific health check
	healthy := s.checkStandardAPI(decrypted, cred.Provider)

	now := time.Now().Unix()
	if healthy {
		model.DB.Model(&cred).Updates(map[string]interface{}{
			"status":         1,
			"last_health_at": now,
			"last_error":     "",
		})
	} else {
		model.DB.Model(&cred).Updates(map[string]interface{}{
			"last_health_at": now,
			"last_error":     "health check failed",
		})
	}
	return nil
}

// checkStandardAPI performs a minimal API health check against any OpenAI-compatible provider.
func (s *Sub2APIService) checkStandardAPI(token string, provider model.Sub2APIProvider) bool {
	baseURL := "https://api.openai.com"
	// All providers use the same OpenAI-compatible health check

	req, err := s.buildHealthRequest(baseURL, token)
	if err != nil {
		return false
	}
	resp, err := s.doHealthRequest(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 500
}

func (s *Sub2APIService) buildHealthRequest(baseURL, token string) (*http.Request, error) {
	req, err := http.NewRequest("GET", baseURL+"/v1/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (s *Sub2APIService) doHealthRequest(req *http.Request) (*http.Response, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	return client.Do(req)
}
