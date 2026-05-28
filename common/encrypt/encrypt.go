package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// Encrypt AES-GCM 加密
func Encrypt(plaintext []byte, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt AES-GCM 解密
func Decrypt(ciphertextB64 string, key []byte) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// DeriveKey 从任意字符串派生 AES-256 key (32 bytes)
// 使用 SHA-256 确保 full entropy 利用
func DeriveKey(secret string) []byte {
	h := sha256.Sum256([]byte(secret))
	return h[:]
}

// Env key protection constants
// Format: QC!<base64-AES-GCM(plaintext+Qsc)>
// - QC! 标记是加密的 env key
// - 解密后去掉尾部的 Qsc 才是真正的 API Key
// 两层保护：即使攻击者拿到 .env + CRYPTO_SECRET，不知道 Qsc 约定也拿不到真 Key
const EnvKeyPrefix = "QC!"
const EnvKeySuffix = "Qsc"

// EncryptEnvKey 加密环境变量中的 API Key
// 在 Key 末尾追加 Qsc 后 AES-GCM 加密，返回 QC!+base64
func EncryptEnvKey(plaintext string, secret string) (string, error) {
	if plaintext == "" {
		return "", errors.New("cannot encrypt empty key")
	}
	// Append secret suffix before encryption
	withSuffix := []byte(plaintext + EnvKeySuffix)
	encrypted, err := Encrypt(withSuffix, DeriveKey(secret))
	if err != nil {
		return "", err
	}
	return EnvKeyPrefix + encrypted, nil
}

// DecryptEnvKey 解密环境变量中的 API Key
// 去除 QC! 前缀 → AES-GCM 解密 → 验证并去除 Qsc 后缀 → 返回真 Key
func DecryptEnvKey(encrypted string, secret string) (string, error) {
	if !strings.HasPrefix(encrypted, EnvKeyPrefix) {
		return "", errors.New("not an encrypted env key (missing QC! prefix)")
	}
	inner := encrypted[len(EnvKeyPrefix):]
	decrypted, err := Decrypt(inner, DeriveKey(secret))
	if err != nil {
		return "", err
	}
	plaintext := string(decrypted)
	if !strings.HasSuffix(plaintext, EnvKeySuffix) {
		return "", errors.New("decrypted key missing Qsc suffix - key may be corrupted")
	}
	return plaintext[:len(plaintext)-len(EnvKeySuffix)], nil
}

// IsEnvKeyEncrypted 判断环境变量 key 是否已加密
func IsEnvKeyEncrypted(value string) bool {
	return strings.HasPrefix(value, EnvKeyPrefix)
}

// Password suffix for DB password encryption
// Even with CRYPTO_SECRET exposed, attacker doesn't know the Qpw suffix convention
const PasswordSuffix = "Qpw"

// EncryptPassword bcrypt 哈希后 AES-GCM 加密
// 先 bcrypt（单向不可逆）→ 追加 Qpw 后缀 → AES-GCM 加密 → base64
// 两层防护：泄露 DB 看不到明文 bcrypt 哈希，泄露 CryptoSecret 不知道 Qpw 约定
func EncryptPassword(password string, cryptoSecret string) (string, error) {
	// Step 1: bcrypt hash (one-way, defends against CryptoSecret-only compromise)
	hashBytes, err := bcryptFromPassword(password)
	if err != nil {
		fmt.Printf("[DBG-EP] bcryptFromPassword FAILED: %v\n", err)
		return "", err
	}
	fmt.Printf("[DBG-EP] bcrypt hash: %s (len=%d)\n", string(hashBytes), len(hashBytes))
	// Step 2: append Qpw suffix
	withSuffix := append(hashBytes, []byte(PasswordSuffix)...)
	fmt.Printf("[DBG-EP] withSuffix len=%d\n", len(withSuffix))
	// Step 3: AES-256-GCM encrypt
	key := DeriveKey(cryptoSecret)
	fmt.Printf("[DBG-EP] cryptoSecret len=%d, key len=%d, key hex=%.32s\n", len(cryptoSecret), len(key), fmt.Sprintf("%x", key))
	encrypted, err := Encrypt(withSuffix, key)
	if err != nil {
		fmt.Printf("[DBG-EP] Encrypt FAILED: %v\n", err)
		return "", err
	}
	fmt.Printf("[DBG-EP] encrypted len=%d prefix=%.20s\n", len(encrypted), encrypted)
	return encrypted, nil
}

// EncryptPasswordFromHash 加密已有 bcrypt 哈希（用于迁移旧数据）
// 直接取 bcrypt hash 追加 Qpw 后缀 → AES-GCM 加密
func EncryptPasswordFromHash(bcryptHash []byte, cryptoSecret string) (string, error) {
	withSuffix := append(bcryptHash, []byte(PasswordSuffix)...)
	encrypted, err := Encrypt(withSuffix, DeriveKey(cryptoSecret))
	if err != nil {
		return "", err
	}
	return encrypted, nil
}

// DecryptPassword 解密密码：base64 解码 → AES-GCM 解密 → 验证并去除 Qpw → 返回 bcrypt 哈希
func DecryptPassword(encryptedB64 string, cryptoSecret string) ([]byte, error) {
	decrypted, err := Decrypt(encryptedB64, DeriveKey(cryptoSecret))
	if err != nil {
		return nil, err
	}
	// Verify and strip Qpw suffix
	if len(decrypted) < len(PasswordSuffix) || string(decrypted[len(decrypted)-len(PasswordSuffix):]) != PasswordSuffix {
		return nil, errors.New("decrypted password missing Qpw suffix - key may be corrupted")
	}
	return decrypted[:len(decrypted)-len(PasswordSuffix)], nil
}

// bcryptFromPassword generates bcrypt hash (extracted for reuse, avoids import cycle)
func bcryptFromPassword(password string) ([]byte, error) {
	passwordBytes := []byte(password)
	hashedPassword, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)
	return hashedPassword, err
}

// isBcryptHash checks if a string looks like a bcrypt hash (for migration detection)
func IsBcryptHash(s string) bool {
	return strings.HasPrefix(s, "$2a$") || strings.HasPrefix(s, "$2b$") || strings.HasPrefix(s, "$2y$")
}
