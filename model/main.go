package model

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"
	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/env"
	"github.com/quantumclaw/quantumclaw/common/helper"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/common/random"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

var DB *gorm.DB
var LOG_DB *gorm.DB

func CreateRootAccountIfNeed() error {
	// Ensure at least one admin/root user exists (role >= 10)
	var rootUser User
	adminExists := DB.Where("role >= ?", RoleAdminUser).First(&rootUser).Error == nil

	// 确保管理员用户有 cash_balance（兼容旧数据库）
	DB.Model(&User{}).Where("role >= ? AND cash_balance = 0", RoleAdminUser).Update("cash_balance", 500000000000000)

	if !adminExists {
		logger.SysLog("no root user exists, creating a root user for you: username is root, password is 123456")
		hashedPassword, err := common.Password2Hash("123456")
		if err != nil {
			return err
		}
		accessToken := random.GetUUID()
		if config.InitialRootAccessToken != "" {
			accessToken = config.InitialRootAccessToken
		}
		rootUser := User{
			Username:    "root",
			Password:    hashedPassword,
			Role:        RoleRootUser,
			Status:      UserStatusEnabled,
			DisplayName: "Root User",
			AccessToken: common.SHA256Hash(accessToken),
			Quota:       500000000000000,
			CashBalance: 500000000000000,
		}
		DB.Create(&rootUser)
		if config.InitialRootToken != "" {
			logger.SysLog("creating initial root token as requested")
			token := Token{
				Id:             1,
				UserId:         rootUser.Id,
				Key:            config.InitialRootToken,
				Status:         TokenStatusEnabled,
				Name:           "Initial Root Token",
				CreatedTime:    helper.GetTimestamp(),
				AccessedTime:   helper.GetTimestamp(),
				ExpiredTime:    -1,
				RemainQuota:    500000000000000,
				UnlimitedQuota: true,
			}
			DB.Create(&token)
		}
	}
	return nil
}

func chooseDB(envName string) (*gorm.DB, error) {
	dsn := os.Getenv(envName)

	switch {
	case strings.HasPrefix(dsn, "sqlserver://"):
		// Use MSSQL
		return openMSSQL(dsn)
	case strings.HasPrefix(dsn, "postgres://"):
		// Use PostgreSQL
		return openPostgreSQL(dsn)
	case dsn != "":
		// Use MySQL
		return openMySQL(dsn)
	default:
		// Use SQLite
		return openSQLite()
	}
}

func openMSSQL(dsn string) (*gorm.DB, error) {
	logger.SysLog("using MSSQL as database")
	common.UsingMySQL = true // reuse MySQL flag for MSSQL (mostly compatible)
	return gorm.Open(sqlserver.Open(dsn), &gorm.Config{
		PrepareStmt: true,
	})
}

func openPostgreSQL(dsn string) (*gorm.DB, error) {
	logger.SysLog("using PostgreSQL as database")
	common.UsingPostgreSQL = true
	return gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{
		PrepareStmt: true, // precompile SQL
	})
}

func openMySQL(dsn string) (*gorm.DB, error) {
	logger.SysLog("using MySQL as database")
	common.UsingMySQL = true
	return gorm.Open(mysql.Open(dsn), &gorm.Config{
		PrepareStmt: true, // precompile SQL
	})
}

func openSQLite() (*gorm.DB, error) {
	logger.SysLog("SQL_DSN not set, using SQLite as database")
	common.UsingSQLite = true
	dsn := fmt.Sprintf("%s?_busy_timeout=%d", common.SQLitePath, common.SQLiteBusyTimeout)
	return gorm.Open(sqlite.Open(dsn), &gorm.Config{
		PrepareStmt: true,
	})
}

func InitDB() {
	var err error
	DB, err = chooseDB("SQL_DSN")
	if err != nil {
		logger.FatalLog("failed to initialize database: " + err.Error())
		return
	}

	setDBConns(DB)

	if !config.IsMasterNode {
		return
	}

	// Legacy MySQL index migration: old schema had an index on TEXT column
	// MySQL 8.0+ requires explicit index name for DROP INDEX
	if common.UsingMySQL {
		if DB.Migrator().HasIndex("channels", "idx_channels_key") {
			_ = DB.Migrator().DropIndex("channels", "idx_channels_key")
			logger.SysLog("dropped legacy index idx_channels_key from channels table")
		}
	}

	logger.SysLog("database migration started")
	if err = migrateDB(); err != nil {
		logger.SysWarn("database migration warning (non-fatal): " + err.Error())
	} else {
		logger.SysLog("database migrated")
	}


	// 预设默认渠道(检测为空时自动插入)
	SeedModelMetadata()
	SeedDefaultChannels()

	// (language initialization removed; frontend i18n only)
}

func migrateDB() error {
	var lastErr error
	attempt := func(name string, fn func() error) {
		if lastErr != nil {
			return
		}
		if e := fn(); e != nil {
			// Skip non-fatal duplicate key/index errors
			if strings.Contains(e.Error(), "Error 1062") || strings.Contains(e.Error(), "Duplicate entry") {
				logger.SysWarn("migrate " + name + " (non-fatal, continuing): " + e.Error())
				return
			}
			logger.SysWarn("migrate " + name + ": " + e.Error())
			lastErr = e
		}
	}
	attempt("Token", func() error { return DB.AutoMigrate(&Token{}) })
	attempt("User", func() error { return DB.AutoMigrate(&User{}) })
	attempt("Option", func() error { return DB.AutoMigrate(&Option{}) })
	attempt("Redemption", func() error { return DB.AutoMigrate(&Redemption{}) })
	attempt("Ability", func() error { return DB.AutoMigrate(&Ability{}) })
	attempt("Log", func() error { return DB.AutoMigrate(&Log{}) })
	attempt("Channel", func() error { return DB.AutoMigrate(&Channel{}) })
	attempt("BalanceLog", func() error { return DB.AutoMigrate(&BalanceLog{}) })
	attempt("ProviderEarning", func() error { return DB.AutoMigrate(&ProviderEarning{}) })
	attempt("WithdrawalRequest", func() error { return DB.AutoMigrate(&WithdrawalRequest{}) })
	attempt("PlatformFeeRecord", func() error { return DB.AutoMigrate(&PlatformFeeRecord{}) })
	// 手动迁移:添加 category 列(SQLite 的 AutoMigrate 有时不添加新列)
	if common.UsingSQLite {
		if !DB.Migrator().HasColumn(&Channel{}, "category") {
			if addErr := DB.Migrator().AddColumn(&Channel{}, "category"); addErr != nil {
				logger.SysLog("add category column note: " + addErr.Error())
			}
		}
	}

	attempt("SubscriptionPlan", func() error { return DB.AutoMigrate(&SubscriptionPlan{}) })
	attempt("SubscriptionOrder", func() error { return DB.AutoMigrate(&SubscriptionOrder{}) })
	attempt("UserSubscription", func() error { return DB.AutoMigrate(&UserSubscription{}) })
	attempt("SubscriptionPreConsume", func() error { return DB.AutoMigrate(&SubscriptionPreConsumeRecord{}) })
	attempt("Checkin", func() error { return DB.AutoMigrate(&Checkin{}) })
	attempt("CustomOAuth", func() error { return DB.AutoMigrate(&CustomOAuthProvider{}) })
	attempt("AsyncTask", func() error { return DB.AutoMigrate(&AsyncTask{}) })
	attempt("MidjourneyTask", func() error { return DB.AutoMigrate(&MidjourneyTask{}) })
	attempt("VideoTask", func() error { return DB.AutoMigrate(&VideoTask{}) })
	attempt("SunoTask", func() error { return DB.AutoMigrate(&SunoTask{}) })
	attempt("WebAuthnCredential", func() error { return DB.AutoMigrate(&WebAuthnCredential{}) })
	attempt("MenuItem", func() error { return DB.AutoMigrate(&MenuItem{}) })
	attempt("RssArticle", func() error { return DB.AutoMigrate(&RssArticle{}) })
	attempt("TransactionLog", func() error { return DB.AutoMigrate(&TransactionLog{}) })
	attempt("ModelMetadata", func() error { return DB.AutoMigrate(&ModelMetadata{}) })
	attempt("Notification", func() error { return DB.AutoMigrate(&Notification{}) })
	attempt("SettlementConfig", func() error { return DB.AutoMigrate(&SettlementConfig{}) })
	attempt("TokenTransaction", func() error { return DB.AutoMigrate(&TokenTransaction{}) })
	attempt("Reseller", func() error { return DB.AutoMigrate(&Reseller{}) })
	attempt("AffiliateRelation", func() error { return DB.AutoMigrate(&AffiliateRelation{}) })
	attempt("PlatformConfig", func() error { return DB.AutoMigrate(&PlatformConfig{}) })

	// ── 结算系统新表 ──

	// 手动迁移：Ability 表新增 user_id 列
	if !DB.Migrator().HasColumn(&Ability{}, "user_id") {
		if addErr := DB.Migrator().AddColumn(&Ability{}, "user_id"); addErr != nil {
			logger.SysLog("add ability.user_id column note: " + addErr.Error())
		}
	}

	// 手动迁移：Channel 表新增 cost_price 列
	if !DB.Migrator().HasColumn(&Channel{}, "cost_price") {
		if addErr := DB.Migrator().AddColumn(&Channel{}, "cost_price"); addErr != nil {
			logger.SysLog("add channel.cost_price column note: " + addErr.Error())
		}
	}

	return lastErr
}

func InitLogDB() {
	if os.Getenv("LOG_SQL_DSN") == "" {
		LOG_DB = DB
		return
	}

	logger.SysLog("using secondary database for table logs")
	var err error
	LOG_DB, err = chooseDB("LOG_SQL_DSN")
	if err != nil {
		logger.FatalLog("failed to initialize secondary database: " + err.Error())
		return
	}

	setDBConns(LOG_DB)

	if !config.IsMasterNode {
		return
	}

	logger.SysLog("secondary database migration started")
	err = migrateLOGDB()
	if err != nil {
		logger.FatalLog("failed to migrate secondary database: " + err.Error())
		return
	}
	logger.SysLog("secondary database migrated")
}

func migrateLOGDB() error {
	var err error
	if err = LOG_DB.AutoMigrate(&Log{}); err != nil {
		return err
	}
	return nil
}

func setDBConns(db *gorm.DB) *sql.DB {
	if config.DebugSQLEnabled {
		db = db.Debug()
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.FatalLog("failed to connect database: " + err.Error())
		return nil
	}

	sqlDB.SetMaxIdleConns(env.Int("SQL_MAX_IDLE_CONNS", 100))
	sqlDB.SetMaxOpenConns(env.Int("SQL_MAX_OPEN_CONNS", 1000))
	sqlDB.SetConnMaxLifetime(time.Second * time.Duration(env.Int("SQL_MAX_LIFETIME", 60)))
	return sqlDB
}

func closeDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	err = sqlDB.Close()
	return err
}

func CloseDB() error {
	if LOG_DB != DB {
		err := closeDB(LOG_DB)
		if err != nil {
			return err
		}
	}
	return closeDB(DB)
}
