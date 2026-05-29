package model

import (
	"fmt"
	"testing"
	"time"

	"github.com/quantumclaw/quantumclaw/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// createRawToken inserts a Token with a raw (unencrypted) key for testing.
// Returns (tokenId, rawKey).
func createRawToken(t *testing.T, db *gorm.DB, userId int, remainQuota int64, unlimited bool) (int, string) {
	t.Helper()
	rawKey := fmt.Sprintf("sk-test-key-%d-%d", userId, time.Now().UnixNano())
	keyHash := common.SHA256Hash(rawKey)
	token := &Token{
		UserId:         userId,
		Key:            rawKey,
		KeyHash:        keyHash,
		Name:           fmt.Sprintf("test-token-%d", userId),
		Status:         TokenStatusEnabled,
		CreatedTime:    time.Now().Unix(),
		AccessedTime:   time.Now().Unix(),
		ExpiredTime:    -1,
		RemainQuota:    remainQuota,
		UnlimitedQuota: unlimited,
		UsedQuota:      0,
	}
	err := db.Create(token).Error
	require.NoError(t, err)
	require.Greater(t, token.Id, 0)
	return token.Id, rawKey
}

// ── ValidateUserToken ──────────────────────────────────

func TestValidateUserToken_Success(t *testing.T) {
	disableRedis(t)
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 50000, 0, nil)
	tokenId, rawKey := createRawToken(t, db, userId, 10000, false)

	token, err := ValidateUserToken(rawKey)
	require.NoError(t, err)
	assert.Equal(t, tokenId, token.Id)
	assert.Equal(t, userId, token.UserId)
	assert.Equal(t, TokenStatusEnabled, token.Status)
}

func TestValidateUserToken_EmptyKey(t *testing.T) {
	disableRedis(t)
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	_, err := ValidateUserToken("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "未提供令牌")
}

func TestValidateUserToken_InvalidKey(t *testing.T) {
	disableRedis(t)
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	_, err := ValidateUserToken("invalid-key-that-does-not-exist")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无效的令牌")
}

func TestValidateUserToken_Exhausted(t *testing.T) {
	disableRedis(t)
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 50000, 0, nil)
	_, rawKey := createRawToken(t, db, userId, 0, false) // remain_quota = 0

	_, err := ValidateUserToken(rawKey)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "额度已用尽")
}

func TestValidateUserToken_Disabled(t *testing.T) {
	disableRedis(t)
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 50000, 0, nil)
	rawKey := fmt.Sprintf("sk-disabled-%d", time.Now().UnixNano())
	keyHash := common.SHA256Hash(rawKey)
	token := &Token{
		UserId:         userId,
		Key:            rawKey,
		KeyHash:        keyHash,
		Name:           "disabled-token",
		Status:         TokenStatusDisabled,
		CreatedTime:    time.Now().Unix(),
		AccessedTime:   time.Now().Unix(),
		ExpiredTime:    -1,
		RemainQuota:    10000,
		UnlimitedQuota: false,
	}
	err := db.Create(token).Error
	require.NoError(t, err)

	_, err = ValidateUserToken(rawKey)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "该令牌状态不可用")
}

func TestValidateUserToken_Expired(t *testing.T) {
	disableRedis(t)
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 50000, 0, nil)
	_, rawKey := createRawToken(t, db, userId, 10000, false)

	// Manually expire the token
	db.Model(&Token{}).Where("user_id = ?", userId).Updates(map[string]interface{}{
		"status":      TokenStatusExpired,
		"expired_time": time.Now().Unix() - 3600,
	})

	_, err := ValidateUserToken(rawKey)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "该令牌已过期")
}

func TestValidateUserToken_Unlimited(t *testing.T) {
	disableRedis(t)
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 50000, 0, nil)
	_, rawKey := createRawToken(t, db, userId, 0, true) // unlimited with 0 remain

	token, err := ValidateUserToken(rawKey)
	require.NoError(t, err)
	assert.True(t, token.UnlimitedQuota)
}

// ── GetTokenByKey ──────────────────────────────────────

func TestGetTokenByKey_Success(t *testing.T) {
	disableRedis(t)
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 50000, 0, nil)
	tokenId, rawKey := createRawToken(t, db, userId, 10000, false)

	// Without cache
	token, err := GetTokenByKey(rawKey, false)
	require.NoError(t, err)
	assert.Equal(t, tokenId, token.Id)
	assert.Equal(t, rawKey, token.Key) // Key is stored as-is since no encryption
}

func TestGetTokenByKey_WithCache(t *testing.T) {
	disableRedis(t)
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 50000, 0, nil)
	tokenId, rawKey := createRawToken(t, db, userId, 10000, false)

	// With cache (Redis disabled → falls through to DB)
	token, err := GetTokenByKey(rawKey, true)
	require.NoError(t, err)
	assert.Equal(t, tokenId, token.Id)
}

func TestGetTokenByKey_Empty(t *testing.T) {
	disableRedis(t)
	_, err := GetTokenByKey("", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key is empty")
}

func TestGetTokenByKey_NotFound(t *testing.T) {
	disableRedis(t)
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	_, err := GetTokenByKey("non-existent-key", false)
	assert.Error(t, err)
}

// ── DecreaseTokenQuota ─────────────────────────────────

func TestDecreaseTokenQuota(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 50000, 0, nil)
	tokenId, _ := createRawToken(t, db, userId, 10000, false)

	err := DecreaseTokenQuota(tokenId, 3000)
	assert.NoError(t, err)

	var token Token
	err = db.First(&token, tokenId).Error
	require.NoError(t, err)
	assert.Equal(t, int64(10000-3000), token.RemainQuota)
	assert.Equal(t, int64(3000), token.UsedQuota)
}

func TestDecreaseTokenQuota_Negative(t *testing.T) {
	err := DecreaseTokenQuota(1, -100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "quota 不能为负数")
}

func TestDecreaseTokenQuota_MultipleDecreases(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 50000, 0, nil)
	tokenId, _ := createRawToken(t, db, userId, 10000, false)

	err := DecreaseTokenQuota(tokenId, 2000)
	require.NoError(t, err)
	err = DecreaseTokenQuota(tokenId, 3000)
	require.NoError(t, err)

	var token Token
	err = db.First(&token, tokenId).Error
	require.NoError(t, err)
	assert.Equal(t, int64(10000-2000-3000), token.RemainQuota)
	assert.Equal(t, int64(5000), token.UsedQuota)
}

func TestDecreaseTokenQuota_Nonexistent(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	err := DecreaseTokenQuota(99999, 100)
	assert.NoError(t, err) // GORM updates with 0 rows affected, no error
}

// ── PreConsumeTokenQuota ──────────────────────────────

func TestPreConsumeTokenQuota_Success(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 50000, 0, nil)
	tokenId, _ := createRawToken(t, db, userId, 10000, false)

	err := PreConsumeTokenQuota(tokenId, 3000)
	assert.NoError(t, err)

	// Token quota decreased
	var token Token
	err = db.First(&token, tokenId).Error
	require.NoError(t, err)
	assert.Equal(t, int64(10000-3000), token.RemainQuota)

	// User quota decreased
	var user User
	err = db.First(&user, userId).Error
	require.NoError(t, err)
	assert.Equal(t, int64(50000-3000), user.Quota)
}

func TestPreConsumeTokenQuota_InsufficientTokenQuota(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 50000, 0, nil)
	tokenId, _ := createRawToken(t, db, userId, 1000, false) // only 1000 remain

	err := PreConsumeTokenQuota(tokenId, 5000)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "令牌额度不足")
}

func TestPreConsumeTokenQuota_InsufficientUserQuota(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 2000, 0, nil) // only 2000 user quota
	tokenId, _ := createRawToken(t, db, userId, 50000, false)

	err := PreConsumeTokenQuota(tokenId, 5000)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "用户额度不足")
}

func TestPreConsumeTokenQuota_NegativeQuota(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 50000, 0, nil)
	tokenId, _ := createRawToken(t, db, userId, 10000, false)

	err := PreConsumeTokenQuota(tokenId, -100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "quota 不能为负数")
}

func TestPreConsumeTokenQuota_UnlimitedToken(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 50000, 0, nil)
	tokenId, _ := createRawToken(t, db, userId, 0, true) // unlimited

	err := PreConsumeTokenQuota(tokenId, 3000)
	assert.NoError(t, err)

	// User quota decreased
	var user User
	err = db.First(&user, userId).Error
	require.NoError(t, err)
	assert.Equal(t, int64(50000-3000), user.Quota)

	// Token quota unchanged (unlimited)
	var token Token
	err = db.First(&token, tokenId).Error
	require.NoError(t, err)
	assert.Equal(t, int64(0), token.RemainQuota)
}

// ── PostConsumeTokenQuota ─────────────────────────────

func TestPostConsumeTokenQuota_Deduct(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 50000, 0, nil)
	tokenId, _ := createRawToken(t, db, userId, 10000, false)

	err := PostConsumeTokenQuota(tokenId, 2000)
	assert.NoError(t, err)

	// User quota decreased
	var user User
	err = db.First(&user, userId).Error
	require.NoError(t, err)
	assert.Equal(t, int64(50000-2000), user.Quota)

	// Token quota decreased
	var token Token
	err = db.First(&token, tokenId).Error
	require.NoError(t, err)
	assert.Equal(t, int64(10000-2000), token.RemainQuota)
	assert.Equal(t, int64(2000), token.UsedQuota)
}

func TestPostConsumeTokenQuota_Refund(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 50000, 0, nil)
	tokenId, _ := createRawToken(t, db, userId, 10000, false)

	// Negative quota means refund (increase back)
	err := PostConsumeTokenQuota(tokenId, -2000)
	assert.NoError(t, err)

	// User quota increased
	var user User
	err = db.First(&user, userId).Error
	require.NoError(t, err)
	assert.Equal(t, int64(50000+2000), user.Quota)

	// Token quota increased, used_quota goes negative (refund reverses used quota)
	var token Token
	err = db.First(&token, tokenId).Error
	require.NoError(t, err)
	assert.Equal(t, int64(10000+2000), token.RemainQuota)
	assert.Equal(t, int64(-2000), token.UsedQuota) // refund subtracts from used_quota
}

func TestPostConsumeTokenQuota_Unlimited(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 50000, 0, nil)
	tokenId, _ := createRawToken(t, db, userId, 0, true)

	err := PostConsumeTokenQuota(tokenId, 2000)
	assert.NoError(t, err)

	// User quota decreased
	var user User
	err = db.First(&user, userId).Error
	require.NoError(t, err)
	assert.Equal(t, int64(50000-2000), user.Quota)

	// Token quota unchanged (unlimited)
	var token Token
	err = db.First(&token, tokenId).Error
	require.NoError(t, err)
	assert.Equal(t, int64(0), token.RemainQuota)
}

func TestPostConsumeTokenQuota_UnlimitedRefund(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 50000, 0, nil)
	tokenId, _ := createRawToken(t, db, userId, 0, true)

	err := PostConsumeTokenQuota(tokenId, -2000)
	assert.NoError(t, err)

	// User quota increased
	var user User
	err = db.First(&user, userId).Error
	require.NoError(t, err)
	assert.Equal(t, int64(50000+2000), user.Quota)
}

// ── Pre + Post in sequence (simulating real flow) ─────

func TestPreAndPostConsumeTokenQuota(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 50000, 0, nil)
	tokenId, _ := createRawToken(t, db, userId, 10000, false)

	// Pre-consume: reserve 4000
	err := PreConsumeTokenQuota(tokenId, 4000)
	require.NoError(t, err)

	// Actual usage: 3500, so refund 500
	err = PostConsumeTokenQuota(tokenId, -500)
	require.NoError(t, err)

	// Final user quota: 50000 - 4000 + 500 = 46500
	var user User
	err = db.First(&user, userId).Error
	require.NoError(t, err)
	assert.Equal(t, int64(50000-4000+500), user.Quota)

	// Final token quota: 10000 - 4000 + 500 = 6500
	var token Token
	err = db.First(&token, tokenId).Error
	require.NoError(t, err)
	assert.Equal(t, int64(10000-4000+500), token.RemainQuota)
}
