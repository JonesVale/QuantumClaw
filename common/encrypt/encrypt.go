package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
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
func DeriveKey(secret string) []byte {
	key := make([]byte, 32)
	copy(key, []byte(secret))
	return key
}

// Env key protection constants
// Format: QC!<base64-AES-GCM(plaintext+Qsc)>
// - QC! 标记是加密的 env key
// - 解密后去掉尾部的 Qsc 才是真正的 API Key
// 两层保护：即使攻击者拿到 .env + CRYPTO_SECRET，不知道 Qsc 约定也拿不到真 Key
const EnvKeyPrefix = "QC!"
const EnvKeySuffix = "Qsc"

// Channel encryption prefix/suffix for tamper detection
const ChannelKeyPrefix = "QC!"
const ChannelKeySuffix = "Qpw"

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

// EncryptChannelKey 加密 Channel API Key
// 在 Key 末尾追加 Qpw 后缀后 AES-GCM 加密，返回 QC!+base64
func EncryptChannelKey(plaintext string, cryptoSecret string) (string, error) {
	if plaintext == "" {
		return "", errors.New("cannot encrypt empty key")
	}
	withSuffix := []byte(plaintext + ChannelKeySuffix)
	encrypted, err := Encrypt(withSuffix, DeriveKey(cryptoSecret))
	if err != nil {
		return "", err
	}
	return ChannelKeyPrefix + encrypted, nil
}

// DecryptChannelKey 解密 Channel API Key
// 支持新旧两种格式：QC!+base64（新）和裸 base64（旧，向后兼容）
func DecryptChannelKey(encrypted string, cryptoSecret string) (string, error) {
	inner := encrypted
	// 新格式有 QC! 前缀
	if strings.HasPrefix(encrypted, ChannelKeyPrefix) {
		inner = encrypted[len(ChannelKeyPrefix):]
	}
	decrypted, err := Decrypt(inner, DeriveKey(cryptoSecret))
	if err != nil {
		return "", err
	}
	plaintext := string(decrypted)
	// 检查 Qpw 后缀（新格式）
	if strings.HasSuffix(plaintext, ChannelKeySuffix) {
		return plaintext[:len(plaintext)-len(ChannelKeySuffix)], nil
	}
	// 旧格式：没有 Qpw 后缀，直接返回（向后兼容）
	return plaintext, nil
}

// IsChannelKeyEncrypted 判断 Key 是否已使用新格式加密
func IsChannelKeyEncrypted(value string) bool {
	return strings.HasPrefix(value, ChannelKeyPrefix)
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
		return "", err
	}
	// Step 2: append Qpw suffix
	withSuffix := append(hashBytes, []byte(PasswordSuffix)...)
	// Step 3: AES-256-GCM encrypt
	key := DeriveKey(cryptoSecret)
	encrypted, err := Encrypt(withSuffix, key)
	if err != nil {
		return "", err
	}
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
