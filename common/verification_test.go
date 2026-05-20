package common

import (
	"fmt"
	"sync"
	"testing"
)

// Reset verification state for tests
func resetVerification() {
	verificationMutex.Lock()
	defer verificationMutex.Unlock()
	verificationMap = make(map[string]verificationValue)
}

func TestGenerateVerificationCode(t *testing.T) {
	tests := []struct {
		name   string
		length int
		want   int
	}{
		{"zero length returns full UUID (no dashes)", 0, 32},
		{"length 6", 6, 6},
		{"length 4", 4, 4},
		{"length 32", 32, 32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateVerificationCode(tt.length)
			if len(got) != tt.want {
				t.Errorf("GenerateVerificationCode(%d) length = %d, want %d", tt.length, len(got), tt.want)
			}
		})
	}
}

func TestGenerateVerificationCode_NoDashes(t *testing.T) {
	// UUIDs contain dashes, but GenerateVerificationCode should strip them
	code := GenerateVerificationCode(0)
	for _, c := range code {
		if c == '-' {
			t.Error("Generated code contains dashes, expected none")
		}
	}
}

func TestRegisterAndVerifyCode(t *testing.T) {
	resetVerification()

	key := "user@example.com"
	code := "123456"
	purpose := EmailVerificationPurpose

	RegisterVerificationCodeWithKey(key, code, purpose)

	if !VerifyCodeWithKey(key, code, purpose) {
		t.Error("VerifyCodeWithKey returned false for valid code")
	}
}

func TestVerifyCode_WrongCode(t *testing.T) {
	resetVerification()

	RegisterVerificationCodeWithKey("user@example.com", "123456", EmailVerificationPurpose)

	if VerifyCodeWithKey("user@example.com", "wrong", EmailVerificationPurpose) {
		t.Error("VerifyCodeWithKey returned true for wrong code")
	}
}

func TestVerifyCode_NonexistentKey(t *testing.T) {
	resetVerification()

	if VerifyCodeWithKey("nonexistent", "123456", EmailVerificationPurpose) {
		t.Error("VerifyCodeWithKey returned true for nonexistent key")
	}
}

func TestVerifyCode_ExpiredCode(t *testing.T) {
	originalValidMinutes := VerificationValidMinutes
	VerificationValidMinutes = 0 // 0 minutes = immediately expired
	defer func() { VerificationValidMinutes = originalValidMinutes }()

	resetVerification()

	RegisterVerificationCodeWithKey("user@test.com", "code123", EmailVerificationPurpose)

	if VerifyCodeWithKey("user@test.com", "code123", EmailVerificationPurpose) {
		t.Error("VerifyCodeWithKey returned true for expired code (0-minute validity)")
	}
}

func TestDeleteKey(t *testing.T) {
	resetVerification()

	RegisterVerificationCodeWithKey("user@delete.com", "code123", EmailVerificationPurpose)

	DeleteKey("user@delete.com", EmailVerificationPurpose)

	if VerifyCodeWithKey("user@delete.com", "code123", EmailVerificationPurpose) {
		t.Error("VerifyCodeWithKey returned true after DeleteKey")
	}
}

func TestDeleteKey_Nonexistent(t *testing.T) {
	resetVerification()

	// Deleting a nonexistent key should not panic
	DeleteKey("nonexistent", EmailVerificationPurpose)
}

func TestDifferentPurposes_SameKey(t *testing.T) {
	resetVerification()

	key := "user@example.com"

	RegisterVerificationCodeWithKey(key, "email-code", EmailVerificationPurpose)
	RegisterVerificationCodeWithKey(key, "reset-code", PasswordResetPurpose)

	if !VerifyCodeWithKey(key, "email-code", EmailVerificationPurpose) {
		t.Error("Email verification code should be valid")
	}
	if !VerifyCodeWithKey(key, "reset-code", PasswordResetPurpose) {
		t.Error("Password reset code should be valid")
	}

	if VerifyCodeWithKey(key, "email-code", PasswordResetPurpose) {
		t.Error("Email code should not work for password reset purpose")
	}
}

func TestMaxSizeLimit_TriggerWithoutPanic(t *testing.T) {
	resetVerification()

	originalMax := verificationMapMaxSize
	originalValidMinutes := VerificationValidMinutes
	verificationMapMaxSize = 100
	VerificationValidMinutes = 100 // prevent expiration during test
	defer func() {
		verificationMapMaxSize = originalMax
		VerificationValidMinutes = originalValidMinutes
	}()

	// Insert 101 entries (1 over limit) — triggers removeExpiredPairs
	// Since none are expired, cleanup won't remove anything.
	// This tests that the trigger runs without panic/crash.
	for i := 0; i < 101; i++ {
		RegisterVerificationCodeWithKey(
			fmt.Sprintf("user-%d", i),
			"code123",
			EmailVerificationPurpose,
		)
	}

	verificationMutex.Lock()
	size := len(verificationMap)
	verificationMutex.Unlock()

	if size < 100 {
		t.Errorf("verificationMap size = %d, expected >= 100 (fresh entries shouldn't be cleaned)", size)
	}
}

func TestConcurrentAccess(t *testing.T) {
	resetVerification()

	var wg sync.WaitGroup
	n := 50

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("concurrent-user-%d", i%26)
			RegisterVerificationCodeWithKey(key, "code123", EmailVerificationPurpose)
		}(i)
	}
	wg.Wait()

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("concurrent-user-%d", i%26)
			VerifyCodeWithKey(key, "code123", EmailVerificationPurpose)
		}(i)
	}
	wg.Wait()

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("concurrent-user-%d", i%26)
			DeleteKey(key, EmailVerificationPurpose)
		}(i)
	}
	wg.Wait()
}

func TestRemoveExpiredPairs(t *testing.T) {
	resetVerification()

	RegisterVerificationCodeWithKey("old-user", "old-code", EmailVerificationPurpose)

	originalValidMinutes := VerificationValidMinutes
	VerificationValidMinutes = -1 // already expired
	defer func() { VerificationValidMinutes = originalValidMinutes }()

	verificationMutex.Lock()
	removeExpiredPairs()
	verificationMutex.Unlock()

	if VerifyCodeWithKey("old-user", "old-code", EmailVerificationPurpose) {
		t.Error("Expired code should have been removed by cleanup")
	}
}

func TestDifferentPurposesDifferentValidity(t *testing.T) {
	resetVerification()

	RegisterVerificationCodeWithKey("email@test.com", "ecode", EmailVerificationPurpose)
	RegisterVerificationCodeWithKey("pwreset@test.com", "pcode", PasswordResetPurpose)

	if !VerifyCodeWithKey("email@test.com", "ecode", EmailVerificationPurpose) {
		t.Error("Fresh email code should be valid")
	}
	if !VerifyCodeWithKey("pwreset@test.com", "pcode", PasswordResetPurpose) {
		t.Error("Fresh password reset code should be valid")
	}
}
