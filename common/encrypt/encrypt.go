package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"strings"
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
