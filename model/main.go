package model

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"
	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/encrypt"
	"github.com/quantumclaw/quantumclaw/common/env"
	"github.com/quantumclaw/quantumclaw/common/helper"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/common/random"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

var DB *gorm.DB
var LOG_DB *gorm.DB

func CreateRootAccountIfNeed() error {
	// Determine root password from config
	// ⚠️ 不再使用硬编码默认密码 "123456"。
	// 如果未设置 INITIAL_ROOT_PASSWORD，将生成随机强密码并在日志中输出。
	rootPassword := config.InitialRootPassword
	if rootPassword == "" {
		rootPassword = random.GetRandomString(24)
		logger.SysLog("⚠️ WARNING: 未设置 INITIAL_ROOT_PASSWORD，已自动生成随机密码，请立即修改！")
		logger.SysLogf("Generated root password: %s", rootPassword)
	}

	// Ensure at least one admin/root user exists (role >= 10)
	var rootUser User
	adminExists := DB.Where("role >= ?", RoleAdminUser).First(&rootUser).Error == nil

	// 确保管理员用户有 cash_balance（兼容旧数据库）
	DB.Model(&User{}).Where("role >= ? AND cash_balance = 0", RoleAdminUser).Update("cash_balance", 500000000000000)

	if !adminExists {
		logger.SysLog("no root user exists, creating a root user for you: username is root")
		hashedPassword, err := common.Password2Hash(rootPassword)
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
			Email:       "admin@quantumclaw.local",
			Phone:       "+8600000000",
			QQ:          "00000",
			AccessToken: common.SHA256Hash(accessToken),
			Quota:       500000000000000,
			CashBalance: 500000000000000,
			AffCode:     random.GetRandomString(8),
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
	} else {
		// Self-healing: verify existing root password hash matches .env, update if stale
		var firstAdmin User
		if DB.Where("role >= ?", RoleAdminUser).Order("id asc").First(&firstAdmin).Error == nil {
			// 仅当 INITIAL_ROOT_PASSWORD 是显式自定义值时同步密码（防止重启覆盖 UI 修改）
			if rootPassword != "123456" && rootPassword != "admin123456" {
				if !common.ValidatePasswordAndHash(rootPassword, firstAdmin.Password) {
					newHash, hashErr := common.Password2Hash(rootPassword)
					if hashErr == nil {
						DB.Model(&firstAdmin).Update("password", newHash)
						logger.SysLogf("root password synced to custom INITIAL_ROOT_PASSWORD for %s", firstAdmin.Username)
					}
				}
			}
			// Self-healing: generate aff_code if missing
			if firstAdmin.AffCode == "" {
				newAffCode := random.GetRandomString(8)
				DB.Model(&firstAdmin).Update("aff_code", newAffCode)
				logger.SysLogf("root aff_code self-healed: generated %s for %s", newAffCode, firstAdmin.Username)
			}
			// Self-healing: set email/phone/qq if empty (prevents UNIQUE constraint conflict)
			if firstAdmin.Email == "" {
				DB.Model(&firstAdmin).Update("email", "admin@quantumclaw.local")
				logger.SysLogf("root email self-healed: set to admin@quantumclaw.local for %s", firstAdmin.Username)
			}
			if firstAdmin.Phone == "" {
				DB.Model(&firstAdmin).Update("phone", "+8600000000")
				logger.SysLogf("root phone self-healed for %s", firstAdmin.Username)
			}
			if firstAdmin.QQ == "" {
				DB.Model(&firstAdmin).Update("qq", "00000")
				logger.SysLogf("root qq self-healed for %s", firstAdmin.Username)
			}
		}
	}

	// Migrate existing plaintext bcrypt password hashes to AES-GCM encrypted format + Qpw suffix
	MigratePlaintextPasswords()

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
	// 运行时重新读取 SQLITE_PATH 环境变量（确保 godotenv 已加载 .env）
	if envPath := os.Getenv("SQLITE_PATH"); envPath != "" {
		common.SQLitePath = envPath
	}
	logger.SysLog("SQLite path: " + common.SQLitePath)
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
	// ⚠️ 以下函数只增不删。不要在此处加任何 DELETE/TRUNCATE/DROP 逻辑。
	// 不要删除 SQLite DB 文件来触发重新播种——AutoPopulateModelMetadataFromRatio
	// 每次启动自动增量填充，不需要清库。
	SeedModelMetadata()
	AutoPopulateModelMetadataFromRatio()
	SeedDefaultChannels()

	// 渠道提供商表 seed（从已存在的硬编码数据填充）
	SeedChannelProviders()

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
	attempt("PlatformIncome", func() error { return DB.AutoMigrate(&PlatformIncome{}) })
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
	attempt("RssSource", func() error { return DB.AutoMigrate(&DbRssSource{}) })
	attempt("TransactionLog", func() error { return DB.AutoMigrate(&TransactionLog{}) })
	attempt("ModelMetadata", func() error { return DB.AutoMigrate(&ModelMetadata{}) })
	attempt("Notification", func() error { return DB.AutoMigrate(&Notification{}) })
	attempt("SettlementConfig", func() error { return DB.AutoMigrate(&SettlementConfig{}) })
	attempt("TokenTransaction", func() error { return DB.AutoMigrate(&TokenTransaction{}) })
	attempt("Reseller", func() error { return DB.AutoMigrate(&Reseller{}) })
	attempt("AffiliateRelation", func() error { return DB.AutoMigrate(&AffiliateRelation{}) })
	attempt("PlatformConfig", func() error { return DB.AutoMigrate(&PlatformConfig{}) })
	attempt("ReconciliationLog", func() error { return DB.AutoMigrate(&ReconciliationLog{}) })
	attempt("HourlySettlement", func() error { return DB.AutoMigrate(&HourlySettlement{}) })

	// ── 交易手续费默认值 ──
	attempt("TransactionFeeDefaults", func() error {
		defaults := map[string]string{
			"transaction_fee_domestic":     "1.0",
			"transaction_fee_foreign":       "3.0",
			"transaction_fee_foreign_min":   "5.00",
			"new_user_trial_balance_cents": "0",
			"platform_fee_min_revenue_cents": "100",
			"platform_fee_rate_percent":     "5.0",
			"icp_beian":                     "粤ICP备2021033000号-2",
		}
		now := helper.GetTimestamp()
		for k, v := range defaults {
			var existing PlatformConfig
			if DB.Where("`key` = ?", k).First(&existing).Error != nil {
				DB.Create(&PlatformConfig{Key: k, Value: v, UpdatedTime: now})
			}
		}
		return nil
	})

	// ── 级联架构表 ──
	attempt("CascadeNode", func() error { return DB.AutoMigrate(&CascadeNode{}) })
	attempt("CascadeBillingBatch", func() error { return DB.AutoMigrate(&CascadeBillingBatch{}) })

	// ── Sub2API 订阅凭证 ──
	attempt("Sub2APICredential", func() error { return DB.AutoMigrate(&Sub2APICredential{}) })
	attempt("Sub2APIUsage", func() error { return DB.AutoMigrate(&Sub2APIUsage{}) })
	attempt("Sub2APISchema", func() error { return DB.AutoMigrate(&Sub2APISchema{}) })

	// ── Seed built-in Sub2API schemas ──

	attempt("Feedback", func() error { return DB.AutoMigrate(&Feedback{}) })
	attempt("FAQ", func() error { return DB.AutoMigrate(&FAQ{}) })
	attempt("AppMarket", func() error { return DB.AutoMigrate(&AppMarket{}) })
	attempt("InferenceNode", func() error { return DB.AutoMigrate(&InferenceNode{}) })

	// ── 组织/团队表 ──
	InitOrganizationTables()
	// ── 渠道商店铺表 ──
	InitStoreTables()
	// ── 企业管理表 ──
	InitEnterpriseTables()
	// ── 消息发送引擎表 ──
	InitMessageTables()
	// ── 多语言翻译表 ──
	InitTranslationTables()
	// init market tables
	InitMarketTables(DB)
	// ── Sub2API 内置模版种子 ──
	attempt("SeedSub2APISchemas", func() error {
		var count int64
		if count > 0 {
			return nil // already seeded
		}
		for _, s := range SeedSub2APISchemas() {
			if err := DB.Create(&s).Error; err != nil {
				logger.SysLog("seed sub2api schema: " + s.Provider + ": " + err.Error())
			}
		}
		return nil
	})

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

	// 级联架构：Token 表新增 updated_time 列
	if !DB.Migrator().HasColumn(&Token{}, "updated_time") {
		if addErr := DB.Migrator().AddColumn(&Token{}, "updated_time"); addErr != nil {
			logger.SysLog("add token.updated_time column note: " + addErr.Error())
		}
	}

	// 渠道余额预警阈值
	if !DB.Migrator().HasColumn(&Channel{}, "balance_alert_threshold") {
		if addErr := DB.Migrator().AddColumn(&Channel{}, "balance_alert_threshold"); addErr != nil {
			logger.SysLog("add balance_alert_threshold column note: " + addErr.Error())
		}
	}
	if !DB.Migrator().HasColumn(&Channel{}, "balance_disable_threshold") {
		if addErr := DB.Migrator().AddColumn(&Channel{}, "balance_disable_threshold"); addErr != nil {
			logger.SysLog("add balance_disable_threshold column note: " + addErr.Error())
		}
	}

	// Pricing: channel_markup column
	if !DB.Migrator().HasColumn(&Channel{}, "channel_markup") {
		if addErr := DB.Migrator().AddColumn(&Channel{}, "channel_markup"); addErr != nil {
			logger.SysLog("add channel_markup column note: " + addErr.Error())
		}
	}

	// Pricing: tiers_json on SubscriptionPlan
	if !DB.Migrator().HasColumn(&SubscriptionPlan{}, "tiers_json") {
		if addErr := DB.Migrator().AddColumn(&SubscriptionPlan{}, "tiers_json"); addErr != nil {
			logger.SysLog("add tiers_json column note: " + addErr.Error())
		}
	}

	// ── 渠道提供商表（统一管理品牌/URL/模型列表）──
	attempt("ChannelProvider", func() error { return DB.AutoMigrate(&ChannelProvider{}) })

	SeedMarketDefaults()

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

// MigratePlaintextPasswords 加密所有明文 bcrypt 密码为 AES-GCM 格式
// 旧格式: $2a$10$... (bcrypt 哈希)
// 新格式: base64(AES-GCM(bcrypt_hash + Qpw))
func MigratePlaintextPasswords() {
	var plaintextUsers []User
	// 使用 Raw SQL 查找所有密码以 $2 开头的用户（GORM 的 Where LIKE 可能转义 $）
	if err := DB.Raw("SELECT * FROM users WHERE password LIKE '$2%'").Scan(&plaintextUsers).Error; err != nil {
		logger.SysWarnf("password migration query failed: %v", err)
		return
	}
	if len(plaintextUsers) == 0 {
		logger.SysLog("password migration: no plaintext bcrypt hashes found (already encrypted or no users)")
		return
	}
	count := 0
	for _, u := range plaintextUsers {
		if config.CryptoSecret == "" {
			logger.SysError("MigratePlaintextPasswords: CRYPTO_SECRET is empty, cannot encrypt passwords")
			return
		}
		encrypted, err := encrypt.EncryptPasswordFromHash([]byte(u.Password), config.CryptoSecret)
		if err != nil {
			logger.SysErrorf("MigratePlaintextPasswords: failed to encrypt password for user %d: %v", u.Id, err)
			continue
		}
		DB.Model(&u).Update("password", encrypted)
		count++
	}
	if count > 0 {
		logger.SysLogf("password migration: encrypted %d plaintext bcrypt hashes (+Qpw suffix)", count)
	}
	// Also migrate cascade_node APIKeyHash
	MigrateCascadeNodeKeys()
}

// MigrateCascadeNodeKeys 加密级联节点的明文 bcrypt 密钥哈希
func MigrateCascadeNodeKeys() {
	var plainNodes []CascadeNode
	if err := DB.Raw("SELECT * FROM cascade_nodes WHERE api_key_hash LIKE '$2%'").Scan(&plainNodes).Error; err != nil {
		logger.SysWarnf("cascade node key migration query failed: %v", err)
		return
	}
	if len(plainNodes) == 0 {
		return
	}
	count := 0
	for _, n := range plainNodes {
		if config.CryptoSecret == "" {
			return
		}
		encrypted, err := encrypt.EncryptPasswordFromHash([]byte(n.APIKeyHash), config.CryptoSecret)
		if err != nil {
			logger.SysErrorf("MigrateCascadeNodeKeys: failed for node %d: %v", n.Id, err)
			continue
		}
		DB.Model(&n).Update("api_key_hash", encrypted)
		count++
	}
	if count > 0 {
		logger.SysLogf("cascade node key migration: encrypted %d api_key_hash values (+Qpw suffix)", count)
	}
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
