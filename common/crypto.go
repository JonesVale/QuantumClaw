package common

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/encrypt"
	"golang.org/x/crypto/bcrypt"
)

// Password2Hash 密码 → bcrypt 哈希 → AES-256-GCM 加密 → base64
// 存储的是加密后的密码，不是明文 bcrypt 哈希
func Password2Hash(password string) (string, error) {
	fmt.Printf("[DBG-PW] Password2Hash called, secret_len=%d, secret_first8=%.8s\n", len(config.CryptoSecret), config.CryptoSecret)
	result, err := encrypt.EncryptPassword(password, config.CryptoSecret)
	if err != nil {
		fmt.Printf("[DBG-PW] Password2Hash FAILED: %v\n", err)
		return "", err
	}
	fmt.Printf("[DBG-PW] Password2Hash result len=%d prefix=%.20s\n", len(result), result)
	return result, nil
}

// ValidatePasswordAndHash 验证密码
// encryptedHash 是 AES-GCM 加密的 bcrypt 哈希
func ValidatePasswordAndHash(password string, encryptedHash string) bool {
	// Try to decrypt (new format: AES-GCM encrypted)
	bcryptHash, err := encrypt.DecryptPassword(encryptedHash, config.CryptoSecret)
	if err == nil {
		// AES-GCM decryption succeeded → compare with bcrypt
		err = bcrypt.CompareHashAndPassword(bcryptHash, []byte(password))
		return err == nil
	}

	// Fallback: maybe it's an old plain bcrypt hash (pre-migration)
	if encrypt.IsBcryptHash(encryptedHash) {
		err = bcrypt.CompareHashAndPassword([]byte(encryptedHash), []byte(password))
		return err == nil
	}

	return false
}

// MigratePasswordHash 将旧版明文 bcrypt 哈希迁移为加密格式
// 如果已经是加密格式则原样返回
func MigratePasswordHash(stored string) string {
	// Already encrypted? (not a bcrypt-looking string)
	if !encrypt.IsBcryptHash(stored) {
		return stored
	}
	// Encrypt the bcrypt hash
	encrypted, err := encrypt.EncryptPasswordFromHash([]byte(stored), config.CryptoSecret)
	if err != nil {
		return stored // fallback: keep as-is
	}
	return encrypted
}

// SHA256Hash 计算 SHA-256 哈希，返回十六进制字符串
func SHA256Hash(data string) string {
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}
