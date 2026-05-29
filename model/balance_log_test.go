package model

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── CreateBalanceLog ───────────────────────────────────

func TestCreateBalanceLog(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 10000, 5000, nil)

	err := CreateBalanceLog(userId, BalanceLogTypeRecharge, 5000, 10000, 0, "test recharge")
	assert.NoError(t, err)

	var log BalanceLog
	err = db.Where("user_id = ?", userId).First(&log).Error
	require.NoError(t, err)
	assert.Equal(t, BalanceLogTypeRecharge, log.Type)
	assert.Equal(t, int64(5000), log.Amount)
	assert.Equal(t, int64(10000), log.Balance)
	assert.Equal(t, 0, log.ChannelId)
	assert.Equal(t, "test recharge", log.Remark)
	assert.Greater(t, log.CreatedAt, int64(0))
}

func TestCreateBalanceLog_ConsumeType(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 10000, 5000, nil)

	err := CreateBalanceLog(userId, BalanceLogTypeConsume, -500, 4500, 42, "test consume")
	assert.NoError(t, err)

	var log BalanceLog
	err = db.Where("user_id = ?", userId).First(&log).Error
	require.NoError(t, err)
	assert.Equal(t, BalanceLogTypeConsume, log.Type)
	assert.Equal(t, int64(-500), log.Amount)
	assert.Equal(t, int64(4500), log.Balance)
	assert.Equal(t, 42, log.ChannelId)
}

func TestCreateBalanceLog_Refund(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 10000, 5000, nil)

	err := CreateBalanceLog(userId, BalanceLogTypeRefund, 200, 5200, 0, "refund")
	assert.NoError(t, err)

	var log BalanceLog
	err = db.Where("user_id = ? AND type = ?", userId, BalanceLogTypeRefund).First(&log).Error
	require.NoError(t, err)
	assert.Equal(t, BalanceLogTypeRefund, log.Type)
}

func TestCreateBalanceLog_AdminAdjust(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 10000, 5000, nil)

	err := CreateBalanceLog(userId, BalanceLogTypeAdmin, 10000, 15000, 0, "admin adjustment")
	assert.NoError(t, err)

	var log BalanceLog
	err = db.Where("user_id = ? AND type = ?", userId, BalanceLogTypeAdmin).First(&log).Error
	require.NoError(t, err)
	assert.Equal(t, BalanceLogTypeAdmin, log.Type)
}

func TestCreateBalanceLog_DebtDeduct(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 10000, 5000, nil)

	err := CreateBalanceLog(userId, BalanceLogTypeDebtDeduct, -3000, 2000, 0, "debt deduction")
	assert.NoError(t, err)

	var log BalanceLog
	err = db.Where("user_id = ? AND type = ?", userId, BalanceLogTypeDebtDeduct).First(&log).Error
	require.NoError(t, err)
	assert.Equal(t, BalanceLogTypeDebtDeduct, log.Type)
	assert.Equal(t, int64(-3000), log.Amount)
	assert.Equal(t, int64(2000), log.Balance)
}

// ── CreateBalanceLog with multiple records ──────────────

func TestCreateBalanceLog_MultipleRecords(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 10000, 5000, nil)

	// Create multiple logs
	for i := 0; i < 5; i++ {
		err := CreateBalanceLog(userId, BalanceLogTypeRecharge, int64(i*1000), int64(5000+i*1000), 0,
			fmt.Sprintf("record %d", i))
		require.NoError(t, err)
	}

	var logs []BalanceLog
	err := db.Where("user_id = ?", userId).Order("id asc").Find(&logs).Error
	require.NoError(t, err)
	assert.Equal(t, 5, len(logs))
}

// ── GetUserBalanceLogs ─────────────────────────────────

func TestGetUserBalanceLogs_Empty(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 10000, 5000, nil)

	logs, err := GetUserBalanceLogs(userId, 10)
	assert.NoError(t, err)
	assert.Empty(t, logs)
}

func TestGetUserBalanceLogs_WithRecords(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 10000, 5000, nil)

	// Create multiple logs
	for i := 1; i <= 3; i++ {
		err := CreateBalanceLog(userId, BalanceLogTypeRecharge, int64(i*1000), int64(5000+i*1000), 0,
			fmt.Sprintf("record %d", i))
		require.NoError(t, err)
	}

	logs, err := GetUserBalanceLogs(userId, 10)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(logs))
	// Should be ordered by id desc (most recent first)
	assert.Equal(t, "record 3", logs[0].Remark)
	assert.Equal(t, "record 2", logs[1].Remark)
	assert.Equal(t, "record 1", logs[2].Remark)
}

func TestGetUserBalanceLogs_Limit(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 10000, 5000, nil)

	// Create 5 logs
	for i := 1; i <= 5; i++ {
		err := CreateBalanceLog(userId, BalanceLogTypeRecharge, int64(i*1000), int64(5000+i*1000), 0,
			fmt.Sprintf("record %d", i))
		require.NoError(t, err)
	}

	// Limit to 2
	logs, err := GetUserBalanceLogs(userId, 2)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(logs))
	assert.Equal(t, "record 5", logs[0].Remark)
	assert.Equal(t, "record 4", logs[1].Remark)
}

func TestGetUserBalanceLogs_OtherUser(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId1 := createTestUser(t, db, 10000, 5000, nil)
	userId2 := createTestUser(t, db, 10000, 5000, nil)

	// Create logs for user1 only
	err := CreateBalanceLog(userId1, BalanceLogTypeRecharge, 1000, 6000, 0, "user1 log")
	require.NoError(t, err)

	// User2 should have no logs
	logs, err := GetUserBalanceLogs(userId2, 10)
	assert.NoError(t, err)
	assert.Empty(t, logs)
}
