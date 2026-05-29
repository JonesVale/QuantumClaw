package model

import (
	"fmt"
	"testing"
	"time"

	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/setting/operation_setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// disableRedis forces Redis off so goroutines in UserCheckin don't panic.
func disableRedis(t *testing.T) {
	t.Helper()
	was := common.RedisEnabled
	common.RedisEnabled = false
	t.Cleanup(func() { common.RedisEnabled = was })
}

// setupTestDBFull returns a DB with all models migrated (Checkin + Token included).
// Reuses the helper from billing_test.go and adds models it doesn't have.
func setupTestDBFull(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)

	// Additional migrations that billing_test doesn't cover
	err := db.AutoMigrate(
		&Checkin{},
		&Token{},
	)
	require.NoError(t, err)
	return db
}

// ── HasCheckedInToday ──────────────────────────────────

func TestHasCheckedInToday_False(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 10000, 0, nil)

	// No checkin record for today
	checked, err := HasCheckedInToday(userId)
	assert.NoError(t, err)
	assert.False(t, checked, "user has not checked in today")
}

func TestHasCheckedInToday_True(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 10000, 0, nil)
	today := time.Now().Format("2006-01-02")

	// Insert a checkin record for today
	checkin := &Checkin{
		UserId:       userId,
		CheckinDate:  today,
		QuotaAwarded: 1000,
		CreatedAt:    time.Now().Unix(),
	}
	err := db.Create(checkin).Error
	require.NoError(t, err)

	checked, err := HasCheckedInToday(userId)
	assert.NoError(t, err)
	assert.True(t, checked, "user has checked in today")
}

func TestHasCheckedInToday_OtherUser(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId1 := createTestUser(t, db, 10000, 0, nil)
	userId2 := createTestUser(t, db, 10000, 0, nil)
	today := time.Now().Format("2006-01-02")

	// Only user1 checked in
	checkin := &Checkin{
		UserId:       userId1,
		CheckinDate:  today,
		QuotaAwarded: 1000,
		CreatedAt:    time.Now().Unix(),
	}
	err := db.Create(checkin).Error
	require.NoError(t, err)

	checked, err := HasCheckedInToday(userId2)
	assert.NoError(t, err)
	assert.False(t, checked, "other user has not checked in today")
}

// ── UserCheckin ────────────────────────────────────────

func TestUserCheckin_Disabled(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 10000, 0, nil)

	// Checkin setting defaults to disabled
	_, err := UserCheckin(userId)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "签到功能未启用")
}

func TestUserCheckin_Success(t *testing.T) {
	disableRedis(t)
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	// Enable checkin for this test
	operation_setting.SetCheckinSetting(operation_setting.CheckinSetting{
		Enabled:  true,
		MinQuota: 500,
		MaxQuota: 500,
	})

	userId := createTestUser(t, db, 10000, 0, nil)

	record, err := UserCheckin(userId)
	require.NoError(t, err)
	assert.Equal(t, userId, record.UserId)
	assert.Equal(t, time.Now().Format("2006-01-02"), record.CheckinDate)
	assert.Equal(t, 500, record.QuotaAwarded)
	assert.Greater(t, record.CreatedAt, int64(0))

	// Verify quota increased
	var user User
	err = db.First(&user, userId).Error
	require.NoError(t, err)
	assert.Equal(t, int64(10000+500), user.Quota, "quota should have increased by award amount")
}

func TestUserCheckin_Duplicate(t *testing.T) {
	disableRedis(t)
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	operation_setting.SetCheckinSetting(operation_setting.CheckinSetting{
		Enabled:  true,
		MinQuota: 500,
		MaxQuota: 500,
	})

	userId := createTestUser(t, db, 10000, 0, nil)

	// First checkin succeeds
	_, err := UserCheckin(userId)
	require.NoError(t, err)

	// Second checkin should fail (already checked in today)
	_, err = UserCheckin(userId)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "今日已签到")
}

func TestUserCheckin_RandomQuota(t *testing.T) {
	disableRedis(t)
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	operation_setting.SetCheckinSetting(operation_setting.CheckinSetting{
		Enabled:  true,
		MinQuota: 100,
		MaxQuota: 10000,
	})

	userId := createTestUser(t, db, 50000, 0, nil)

	record, err := UserCheckin(userId)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, record.QuotaAwarded, 100)
	assert.LessOrEqual(t, record.QuotaAwarded, 10000)

	var user User
	err = db.First(&user, userId).Error
	require.NoError(t, err)
	assert.Equal(t, int64(50000+record.QuotaAwarded), user.Quota)
}

// ── GetUserCheckinStats ─────────────────────────────────

func TestGetUserCheckinStats_Empty(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 10000, 0, nil)

	now := time.Now()
	month := now.Format("2006-01")

	stats, err := GetUserCheckinStats(userId, month)
	assert.NoError(t, err)
	assert.NotNil(t, stats)

	assert.Equal(t, int64(0), stats["total_quota"])
	assert.Equal(t, int64(0), stats["total_checkins"])
	assert.Equal(t, 0, stats["checkin_count"])
	assert.Equal(t, false, stats["checked_in_today"])
	assert.Empty(t, stats["records"])
}

func TestGetUserCheckinStats_WithRecords(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 10000, 0, nil)
	now := time.Now()
	month := now.Format("2006-01")

	// Insert 3 checkin records in this month
	for i := 1; i <= 3; i++ {
		dayStr := fmt.Sprintf("%s-%02d", month, i)
		checkin := &Checkin{
			UserId:       userId,
			CheckinDate:  dayStr,
			QuotaAwarded: i * 1000,
			CreatedAt:    now.AddDate(0, 0, -i).Unix(),
		}
		err := db.Create(checkin).Error
		require.NoError(t, err)
	}

	stats, err := GetUserCheckinStats(userId, month)
	assert.NoError(t, err)
	assert.NotNil(t, stats)

	assert.Equal(t, int64(1000+2000+3000), stats["total_quota"])
	assert.Equal(t, int64(3), stats["total_checkins"])
	assert.Equal(t, 3, stats["checkin_count"])

	records, ok := stats["records"].([]CheckinRecord)
	assert.True(t, ok, "records should be []CheckinRecord")
	assert.Equal(t, 3, len(records))
}

func TestGetUserCheckinStats_CheckedInToday(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 10000, 0, nil)
	now := time.Now()
	month := now.Format("2006-01")
	today := now.Format("2006-01-02")

	// Insert today's checkin
	checkin := &Checkin{
		UserId:       userId,
		CheckinDate:  today,
		QuotaAwarded: 2000,
		CreatedAt:    now.Unix(),
	}
	err := db.Create(checkin).Error
	require.NoError(t, err)

	stats, err := GetUserCheckinStats(userId, month)
	assert.NoError(t, err)
	assert.True(t, stats["checked_in_today"].(bool))
}

// ── IncreaseUserQuotaForCheckin ─────────────────────────

func TestIncreaseUserQuotaForCheckin(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 5000, 0, nil)

	err := IncreaseUserQuotaForCheckin(userId, 3000)
	assert.NoError(t, err)

	var user User
	err = db.First(&user, userId).Error
	require.NoError(t, err)
	assert.Equal(t, int64(5000+3000), user.Quota, "quota should increase by 3000")
}

func TestIncreaseUserQuotaForCheckin_Zero(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 5000, 0, nil)

	err := IncreaseUserQuotaForCheckin(userId, 0)
	assert.NoError(t, err)

	var user User
	err = db.First(&user, userId).Error
	require.NoError(t, err)
	assert.Equal(t, int64(5000), user.Quota)
}
