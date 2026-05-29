package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── IsAdmin ────────────────────────────────────────────

func TestIsAdmin_RootUser(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 0, 0, map[string]interface{}{
		"role": RoleRootUser,
	})

	assert.True(t, IsAdmin(userId), "root user should be admin")
}

func TestIsAdmin_AdminUser(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 0, 0, map[string]interface{}{
		"role": RoleAdminUser,
	})

	assert.True(t, IsAdmin(userId), "admin user should be admin")
}

func TestIsAdmin_CommonUser(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 0, 0, nil) // default role = RoleCommonUser

	assert.False(t, IsAdmin(userId), "common user should not be admin")
}

func TestIsAdmin_Supplier(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 0, 0, map[string]interface{}{
		"role": RoleSupplier,
	})

	assert.False(t, IsAdmin(userId), "supplier should not be admin")
}

func TestIsAdmin_GuestUser(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	userId := createTestUser(t, db, 0, 0, map[string]interface{}{
		"role": RoleGuestUser,
	})

	assert.False(t, IsAdmin(userId), "guest user should not be admin")
}

func TestIsAdmin_ZeroId(t *testing.T) {
	assert.False(t, IsAdmin(0), "zero id should not be admin")
}

func TestIsAdmin_NonexistentUser(t *testing.T) {
	// No user in DB with id 99999
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	assert.False(t, IsAdmin(99999), "nonexistent user should not be admin")
}

// ── IsEmailAlreadyTaken ───────────────────────────────

func TestIsEmailAlreadyTaken_True(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	email := "test-existing@example.com"
	createTestUser(t, db, 0, 0, map[string]interface{}{
		"email": email,
	})

	assert.True(t, IsEmailAlreadyTaken(email))
}

func TestIsEmailAlreadyTaken_False(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	// No user with this email
	assert.False(t, IsEmailAlreadyTaken("nonexistent@example.com"))
}

func TestIsEmailAlreadyTaken_EmptyString(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	// Empty string should not match any email
	assert.False(t, IsEmailAlreadyTaken(""))
}

func TestIsEmailAlreadyTaken_DifferentCase(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	email := "CaseSensitive@Example.com"
	createTestUser(t, db, 0, 0, map[string]interface{}{
		"email": email,
	})

	// GORM/SQLite should match exact case
	assert.True(t, IsEmailAlreadyTaken(email))
	assert.False(t, IsEmailAlreadyTaken("casesensitive@example.com"))
}

// ── IsUsernameAlreadyTaken ────────────────────────────

func TestIsUsernameAlreadyTaken_True(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	username := "testuser_taken"
	// createTestUser generates unique usernames, but we can override with map
	userId := createTestUser(t, db, 0, 0, nil) // create our user

	// We need a known username; create one manually
	u := User{
		Username:    username,
		Password:    "hashed_pw",
		DisplayName: "Test User",
		Email:       "taken_username@test.local",
		AccessToken: "at_taken_username_hash",
		AffCode:     "ac_taken_uname",
		Phone:       "ph_taken_uname",
		QQ:          "qq_taken_uname",
		Role:        RoleCommonUser,
		Status:      UserStatusEnabled,
		Quota:       10000,
		CashBalance: 0,
		UserType:    UserTypeConsumer,
	}
	err := db.Create(&u).Error
	require.NoError(t, err)
	_ = userId // suppress unused warning

	assert.True(t, IsUsernameAlreadyTaken(username))
}

func TestIsUsernameAlreadyTaken_False(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	assert.False(t, IsUsernameAlreadyTaken("nonexistent_user_xyz"))
}

func TestIsUsernameAlreadyTaken_EmptyString(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	assert.False(t, IsUsernameAlreadyTaken(""))
}

// ── IsEmailAlreadyTaken + IsUsernameAlreadyTaken together ──

func TestEmailAndUsernameUniquenessIndependently(t *testing.T) {
	db := setupTestDBFull(t)
	defer withDB(t, db)()

	// Create first user
	email := "unique_combined@test.local"
	username := "unique_combined_user"
	userId := createTestUser(t, db, 0, 0, map[string]interface{}{
		"email": email,
	})
	_ = userId

	// This specific username isn't taken yet
	assert.False(t, IsUsernameAlreadyTaken(username),
		"username should not be taken before we create a user with it")

	// Create second user with this username
	u := User{
		Username:    username,
		Password:    "hashed_pw",
		DisplayName: "Test User 2",
		Email:       "unique_combined2@test.local",
		AccessToken: "at_combined2_hash",
		AffCode:     "ac_combined2",
		Phone:       "ph_combined2",
		QQ:          "qq_combined2",
		Role:        RoleCommonUser,
		Status:      UserStatusEnabled,
		Quota:       10000,
		CashBalance: 0,
		UserType:    UserTypeConsumer,
	}
	err := db.Create(&u).Error
	require.NoError(t, err)

	assert.True(t, IsUsernameAlreadyTaken(username))
	assert.True(t, IsEmailAlreadyTaken(email))

	// Different email/username should not be taken
	assert.False(t, IsEmailAlreadyTaken("other@test.local"))
	assert.False(t, IsUsernameAlreadyTaken("other_user"))
}
