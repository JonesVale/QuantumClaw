package model

import (
	"database/sql"
	"fmt"
	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/env"
	"github.com/quantumclaw/quantumclaw/common/helper"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/common/random"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"strings"
	"time"
)

var DB *gorm.DB
var LOG_DB *gorm.DB

func CreateRootAccountIfNeed() error {
	// Ensure at least one admin/root user exists (role >= 10)
	var rootUser User
	adminExists := DB.Where("role >= ?", RoleAdminUser).First(&rootUser).Error == nil
	
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
			AccessToken: accessToken,
			Quota:       500000000000000,
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
		PrepareStmt: true, // precompile SQL
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
		logger.FatalLog("failed to migrate database: " + err.Error())
		return
	}
	logger.SysLog("database migrated")

	// Initialize language types
	InitLanguageTypes()
	logger.SysLog("language types initialized")

	// Initialize Chinese language resources
	InitChineseLanguageResources()
	logger.SysLog("Chinese language resources initialized")
}

func migrateDB() error {
	var err error
	if err = DB.AutoMigrate(&Token{}); err != nil {
		return err
	}
	if err = DB.AutoMigrate(&User{}); err != nil {
		return err
	}
	if err = DB.AutoMigrate(&Option{}); err != nil {
		return err
	}
	if err = DB.AutoMigrate(&Redemption{}); err != nil {
		return err
	}
	if err = DB.AutoMigrate(&Ability{}); err != nil {
		return err
	}
	if err = DB.AutoMigrate(&Log{}); err != nil {
		return err
	}
	if err = DB.AutoMigrate(&Channel{}); err != nil {
		return err
	}
	// 订阅制计费系统
	if err = DB.AutoMigrate(&SubscriptionPlan{}); err != nil {
		return err
	}
	if err = DB.AutoMigrate(&SubscriptionOrder{}); err != nil {
		return err
	}
	if err = DB.AutoMigrate(&UserSubscription{}); err != nil {
		return err
	}
	if err = DB.AutoMigrate(&SubscriptionPreConsumeRecord{}); err != nil {
		return err
	}
	// 签到系统
	if err = DB.AutoMigrate(&Checkin{}); err != nil {
		return err
	}
	// 自定义 OAuth 提供商
	if err = DB.AutoMigrate(&CustomOAuthProvider{}); err != nil {
		return err
	}
	// 异步任务系统（Midjourney/视频生成/Suno音乐）
	if err = DB.AutoMigrate(&AsyncTask{}); err != nil {
		return err
	}
	if err = DB.AutoMigrate(&MidjourneyTask{}); err != nil {
		return err
	}
	if err = DB.AutoMigrate(&VideoTask{}); err != nil {
		return err
	}
	if err = DB.AutoMigrate(&SunoTask{}); err != nil {
		return err
	}
	// WebAuthn/Passkey 无密码登录
	if err = DB.AutoMigrate(&WebAuthnCredential{}); err != nil {
		return err
	}
	// 多语言系统
	if err = DB.AutoMigrate(&LanguageType{}); err != nil {
		return err
	}
	if err = DB.AutoMigrate(&LanguageResource{}); err != nil {
		return err
	}
	// RSS 文章
	if err = DB.AutoMigrate(&RssArticle{}); err != nil {
		return err
	}
	// 交易审计日志
	if err = DB.AutoMigrate(&TransactionLog{}); err != nil {
		return err
	}
	// 用户通知
	if err = DB.AutoMigrate(&Notification{}); err != nil {
		return err
	}
	return nil
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
