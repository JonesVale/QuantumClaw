package model

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── GenerateSecureTradeNo ──────────────────────────────────

func TestGenerateSecureTradeNo_Format(t *testing.T) {
	userId := 42
	tradeNo, err := GenerateSecureTradeNo(userId)
	require.NoError(t, err)
	// 格式: QC<timestamp><16 hex><4 hex user hash>
	assert.Regexp(t, `^QC\d{10}[0-9a-f]{16}[0-9a-f]{4}$`, tradeNo)
	assert.Contains(t, tradeNo, "002a") // 42 % 65536 = 002a
}

func TestGenerateSecureTradeNo_ZeroUser(t *testing.T) {
	tradeNo, err := GenerateSecureTradeNo(0)
	require.NoError(t, err)
	assert.Regexp(t, `^QC\d{10}[0-9a-f]{16}0000$`, tradeNo)
}

func TestGenerateSecureTradeNo_Unique(t *testing.T) {
	n := 10
	seen := make(map[string]bool, n)
	for i := 0; i < n; i++ {
		tradeNo, err := GenerateSecureTradeNo(i)
		require.NoError(t, err)
		assert.False(t, seen[tradeNo], "订单号重复: %s", tradeNo)
		seen[tradeNo] = true
	}
}

// ── TopUp.IsExpired ───────────────────────────────────────

func TestTopUpIsExpired(t *testing.T) {
	now := time.Now().Unix()
	tests := []struct {
		name     string
		expireAt int64
		want     bool
	}{
		{"刚创建未过期", now + 3600, false},
		{"已过期", now - 1, true},
		{"刚好过期", now - 3600, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topUp := TopUp{ExpireTime: tt.expireAt}
			assert.Equal(t, tt.want, topUp.IsExpired())
		})
	}
}

// ── TopUp.CanProcess ───────────────────────────────────────

func TestTopUpCanProcess_Pending(t *testing.T) {
	topUp := TopUp{Id: 1, Status: TopUpStatusPending, ExpireTime: time.Now().Unix() + 3600}
	assert.NoError(t, topUp.CanProcess())
}

func TestTopUpCanProcess_AlreadySuccess(t *testing.T) {
	topUp := TopUp{Id: 1, Status: TopUpStatusSuccess, ExpireTime: time.Now().Unix() + 3600}
	assert.ErrorIs(t, topUp.CanProcess(), ErrTopUpStatusInvalid)
}

func TestTopUpCanProcess_ExpiredOrder(t *testing.T) {
	db := setupTestDB(t)
	defer withDB(t, db)()

	// 插入一个过期订单
	createTime := time.Now().Unix() - 7200
	topUp := TopUp{
		UserId:          6,
		Amount:          1000,
		TradeNo:         "QC-CANPROCESS-EXP-001",
		PaymentMethod:   "alipay",
		PaymentProvider: "alipay",
		Status:          TopUpStatusPending,
		CreateTime:      createTime,
		ExpireTime:      createTime + 1, // 1秒后过期
	}
	err := topUp.Insert()
	require.NoError(t, err)

	// 等待一小段时间确保过期
	err = topUp.CanProcess()
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTopUpExpired)
}

// ── TopUp Signature ────────────────────────────────────────

func TestTopUpCalculateVerifySignature(t *testing.T) {
	topUp := TopUp{
		UserId:     1,
		Amount:     5000,
		TradeNo:    "QC1234567890abcdef123456780001",
		CreateTime: time.Now().Unix(),
	}
	secret := "test-secret-key"
	sig := topUp.CalculateSignature(secret)
	assert.NotEmpty(t, sig)

	// 设置 Signature 后验证
	topUp.Signature = sig
	assert.True(t, topUp.VerifySignature(secret))
	assert.False(t, topUp.VerifySignature("wrong-secret"))
}

// ── TopUp Insert / GetTopUpById ────────────────────────────

func TestTopUpInsertAndGetById(t *testing.T) {
	db := setupTestDB(t)
	defer withDB(t, db)()

	topUp := TopUp{
		UserId:          1,
		Amount:          10000,
		Money:           100.0,
		TradeNo:         "QC-INSERT-TEST-001",
		PaymentMethod:   "alipay",
		PaymentProvider: "alipay",
		Status:          TopUpStatusPending,
		CreateTime:      time.Now().Unix(),
		ExpireTime:      time.Now().Unix() + 1800,
	}
	err := topUp.Insert()
	require.NoError(t, err)
	assert.Greater(t, topUp.Id, int64(0))

	// 通过 ID 查询
	found, err := GetTopUpById(topUp.Id)
	require.NoError(t, err)
	assert.Equal(t, topUp.TradeNo, found.TradeNo)
	assert.Equal(t, topUp.Status, found.Status)

	// 不存在的 ID
	_, err = GetTopUpById(99999)
	require.Error(t, err)
}

// ── GetTopUpByTradeNo ──────────────────────────────────────

func TestGetTopUpByTradeNo(t *testing.T) {
	db := setupTestDB(t)
	defer withDB(t, db)()

	topUp := TopUp{
		UserId:          2,
		Amount:          5000,
		TradeNo:         "QC-TRADENO-TEST-001",
		PaymentMethod:   "stripe",
		PaymentProvider: "stripe",
		Status:          TopUpStatusPending,
		CreateTime:      time.Now().Unix(),
	}
	err := topUp.Insert()
	require.NoError(t, err)

	found, err := GetTopUpByTradeNo("QC-TRADENO-TEST-001")
	require.NoError(t, err)
	assert.Equal(t, topUp.Id, found.Id)

	_, err = GetTopUpByTradeNo("NONEXISTENT")
	require.Error(t, err)
}

// ── UpdateTopUpStatus ─────────────────────────────────────

func TestUpdateTopUpStatus_PendingToSuccess(t *testing.T) {
	db := setupTestDB(t)
	defer withDB(t, db)()

	topUp := TopUp{
		UserId:          3,
		Amount:          5000,
		TradeNo:         "QC-STATUS-UPDATE-001",
		PaymentMethod:   "alipay",
		PaymentProvider: "alipay",
		Status:          TopUpStatusPending,
		CreateTime:      time.Now().Unix(),
	}
	err := topUp.Insert()
	require.NoError(t, err)

	err = UpdateTopUpStatus("QC-STATUS-UPDATE-001", "alipay", TopUpStatusSuccess)
	require.NoError(t, err)

	found, _ := GetTopUpByTradeNo("QC-STATUS-UPDATE-001")
	assert.Equal(t, TopUpStatusSuccess, found.Status)
}

func TestUpdateTopUpStatus_WrongProvider(t *testing.T) {
	db := setupTestDB(t)
	defer withDB(t, db)()

	topUp := TopUp{
		UserId:          4,
		Amount:          5000,
		TradeNo:         "QC-STATUS-UPDATE-002",
		PaymentMethod:   "alipay",
		PaymentProvider: "alipay",
		Status:          TopUpStatusPending,
		CreateTime:      time.Now().Unix(),
	}
	err := topUp.Insert()
	require.NoError(t, err)

	err = UpdateTopUpStatus("QC-STATUS-UPDATE-002", "stripe", TopUpStatusSuccess)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrPaymentMethodMismatch)

	// 状态不应变化
	found, _ := GetTopUpByTradeNo("QC-STATUS-UPDATE-002")
	assert.Equal(t, TopUpStatusPending, found.Status)
}

// ── GetUserTopUps ──────────────────────────────────────────

func TestGetUserTopUps(t *testing.T) {
	db := setupTestDB(t)
	defer withDB(t, db)()

	userId := 10
	for i := 0; i < 5; i++ {
		topUp := TopUp{
			UserId:          userId,
			Amount:          int64(1000 * (i + 1)),
			TradeNo:         fmt.Sprintf("QC-USER-TOPUPS-%03d", i),
			PaymentMethod:   "alipay",
			PaymentProvider: "alipay",
			Status:          TopUpStatusSuccess,
			CreateTime:      time.Now().Unix(),
		}
		err := topUp.Insert()
		require.NoError(t, err)
	}

	// 分页查询
	topUps, total, err := GetUserTopUps(userId, 1, 3)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, topUps, 3)

	// 第二页
	topUps2, total2, err := GetUserTopUps(userId, 2, 3)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total2)
	assert.Len(t, topUps2, 2)
}

// ── CleanExpiredTopUps ─────────────────────────────────────

func TestCleanExpiredTopUps(t *testing.T) {
	db := setupTestDB(t)
	defer withDB(t, db)()

	// 创建过期订单
	expired := TopUp{
		UserId:          5,
		Amount:          1000,
		TradeNo:         "QC-EXPIRED-001",
		PaymentMethod:   "alipay",
		PaymentProvider: "alipay",
		Status:          TopUpStatusPending,
		CreateTime:      time.Now().Unix() - 7200,
		ExpireTime:      time.Now().Unix() - 3600,
	}
	err := expired.Insert()
	require.NoError(t, err)

	// 创建未过期订单
	valid := TopUp{
		UserId:          5,
		Amount:          2000,
		TradeNo:         "QC-VALID-001",
		PaymentMethod:   "alipay",
		PaymentProvider: "alipay",
		Status:          TopUpStatusPending,
		CreateTime:      time.Now().Unix(),
		ExpireTime:      time.Now().Unix() + 3600,
	}
	err = valid.Insert()
	require.NoError(t, err)

	err = CleanExpiredTopUps()
	require.NoError(t, err)

	// 验证过期订单状态
	exp, _ := GetTopUpByTradeNo("QC-EXPIRED-001")
	assert.Equal(t, TopUpStatusExpired, exp.Status)

	// 未过期的不变
	v, _ := GetTopUpByTradeNo("QC-VALID-001")
	assert.Equal(t, TopUpStatusPending, v.Status)
}
