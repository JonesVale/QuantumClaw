package main

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/grafana/pyroscope-go"
	_ "github.com/joho/godotenv/autoload"

	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/client"
	"github.com/quantumclaw/quantumclaw/common/encrypt"
	"github.com/quantumclaw/quantumclaw/relay/channeltype"
	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/i18n"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/controller"
	"github.com/quantumclaw/quantumclaw/middleware"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/openai"
	"github.com/quantumclaw/quantumclaw/router"
	"github.com/quantumclaw/quantumclaw/service"
)

// Known API key environment variables for QC! migration
var knownApiEnvVars = map[string]bool{
	"DEEPSEEK_API_KEY":   true,
	"GROQ_API_KEY":       true,
	"OPENAI_API_KEY":     true,
	"ANTHROPIC_API_KEY":  true,
	"GEMINI_API_KEY":     true,
	"SILICONFLOW_API_KEY": true,
	"MISTRAL_API_KEY":    true,
	"MOONSHOT_API_KEY":   true,
	"BAIDU_API_KEY":      true,
	"ZHIPU_API_KEY":      true,
	"TENCENT_API_KEY":    true,
	"ALIBABA_API_KEY":    true,
}

// migrateEnvKeysToEncrypted scans .env for plaintext API keys and encrypts them to QC! format.
// Adds defense-in-depth: even if .env is stolen, keys need decryption + Qsc suffix removal.
func migrateEnvKeysToEncrypted() int {
	if config.CryptoSecret == "" {
		return 0
	}
	envPath := ".env"
	envData, err := os.ReadFile(envPath)
	if err != nil {
		logger.SysWarn("migrateEnvKeys: cannot read .env: " + err.Error())
		return 0
	}
	lines := strings.Split(string(envData), "\n")
	changed := 0

	for i, line := range lines {
		lineTrim := strings.TrimSpace(line)
		if lineTrim == "" || strings.HasPrefix(lineTrim, "#") {
			continue
		}
		parts := strings.SplitN(lineTrim, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		value := parts[1]

		if !knownApiEnvVars[key] {
			continue
		}
		if value == "" || strings.HasPrefix(value, encrypt.EnvKeyPrefix) {
			continue // already encrypted or empty
		}
		if strings.Contains(value, "PUT_YOUR") {
			continue // placeholder, no need to encrypt
		}

		encrypted, err := encrypt.EncryptEnvKey(value, config.CryptoSecret)
		if err != nil {
			logger.SysWarn("migrateEnvKeys: encrypt " + key + " failed: " + err.Error())
			continue
		}
		lines[i] = key + "=" + encrypted
		changed++
	}

	if changed > 0 {
		if err := os.WriteFile(envPath, []byte(strings.Join(lines, "\n")), 0644); err != nil {
			logger.SysError("migrateEnvKeys: write .env failed: " + err.Error())
			return 0
		}
		logger.SysLogf("migrateEnvKeys: encrypted %d env var(s) to QC! format", changed)
	}
	return changed
}

//go:embed web/default/dist
var buildFS embed.FS

func startPyroscope() {
	if config.PyroscopeURL == "" {
		return
	}
	_, err := pyroscope.Start(pyroscope.Config{
		ApplicationName: config.PyroscopeAppName,
		ServerAddress:   config.PyroscopeURL,
		BasicAuthUser:   config.PyroscopeBasicAuthUser,
		BasicAuthPassword: config.PyroscopeBasicAuthPassword,
		Logger:          pyroscope.StandardLogger,
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,
			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		},
	})
	if err != nil {
		logger.SysWarn("failed to start pyroscope: " + err.Error())
	}
}

// safeGoWithRestart 鍚姩涓€涓?panic-safe 鐨?goroutine锛屽彂鐢?panic 鏃惰嚜鍔ㄩ噸鍚?
func safeGoWithRestart(name string, fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.SysError(fmt.Sprintf("[PANIC] %s: %v 鈥?restarting in 10s", name, r))
				time.Sleep(10 * time.Second)
				safeGoWithRestart(name, fn)
			}
		}()
		logger.SysLogf("[%s] started", name)
		fn()
	}()
}

func main() {
	common.Init()
	logger.SetupLogger()
	logger.SysLogf("QuantumClaw %s started", common.Version)

	// 楠岃瘉棰戦亾绫诲瀷瀹氫箟涓€鑷存€?
	if err := channeltype.ValidateChannelBaseURLs(); err != nil {
		logger.SysError(err.Error())
		os.Exit(1)
	}

	if os.Getenv("GIN_MODE") != gin.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	}
	if config.DebugEnabled {
		logger.SysLog("running in debug mode")
	}

	// Start Pyroscope profiling
	startPyroscope()

	// Initialize Quantum RNG (seed once, async)
	if config.QRNGEnabled {
		common.QRNGEnabled = true
		common.QRNGSourceURL = config.QRNGSourceURL
		common.InitQRNGSeed()
		logger.SysLog("quantum random number generator enabled (source: " + config.QRNGSourceURL + ")")
	}

	// Initialize SQL Database
	model.InitDB()

	// 鈹€鈹€ 鏁版嵁搴撹繛鎺ユ睜閰嶇疆 鈹€鈹€
	if sqlDB, err := model.DB.DB(); err == nil {
		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetConnMaxLifetime(5 * time.Minute)
		logger.SysLog("db pool: max_open=25 max_idle=10 max_lifetime=5m")
	}

	model.InitLogDB()

	// Initialize commission tables
	model.InitCommissionTables()

	// Initialize distributor tables
	model.InitDistributorTables()

	// Initialize promo ads tables
	model.InitPromoAdsTables()

	// Seed default menus (upsert by MenuKey, idempotent)
	model.SeedDefaultMenus()

	// Initialize reference pricing and brand ranking tables
	model.InitReferencePriceTable()
	model.InitBrandRankingTable()

	// Seed reference prices from ModelRatio (idempotent: skips existing)
	service.SeedReferencePrices()

	// Brand rankings are seeded via monthly cron only (avoids external API on startup)

	// Also try reading CRYPTO_SECRET from .env directly (godotenv/autoload fallback)
	if config.CryptoSecret == "" {
		envData, _ := os.ReadFile(".env")
		if envData != nil {
			for _, line := range strings.Split(string(envData), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "CRYPTO_SECRET=") {
					config.CryptoSecret = strings.TrimPrefix(line, "CRYPTO_SECRET=")
					break
				}
			}
		}
	}

	// Ensure CRYPTO_SECRET is set: if not, generate from QRNG and persist to .env
	if config.CryptoSecret == "" {
		var keyBytes []byte
		var err error
		keyBytes, err = common.GetQuantumRandomBytes(32)
		if err != nil {
			logger.SysWarn("QRNG unavailable, using crypto/rand for CRYPTO_SECRET")
			b := make([]byte, 32)
			_, _ = rand.Reader.Read(b)
			keyBytes = b
		}
		config.CryptoSecret = hex.EncodeToString(keyBytes)
		// Persist to .env so restarts use the same key (required for decryption)
		envPath := ".env"
		data, readErr := os.ReadFile(envPath)
		if readErr == nil {
			content := string(data)
			if !strings.Contains(content, "CRYPTO_SECRET") {
				f, openErr := os.OpenFile(envPath, os.O_APPEND|os.O_WRONLY, 0644)
				if openErr == nil {
					f.WriteString("\n# AES-256-GCM master key for channel API key encryption (32 bytes hex)\nCRYPTO_SECRET=" + config.CryptoSecret + "\n")
					f.Close()
					logger.SysLog("CRYPTO_SECRET persisted to .env")
				} else {
					logger.SysWarn("Could not write CRYPTO_SECRET to .env: " + openErr.Error())
				}
			}
		}
	}

	// Migrate plaintext API keys in .env to encrypted QC! format
	if config.CryptoSecret != "" {
		migrateCount := migrateEnvKeysToEncrypted()
		if migrateCount > 0 {
			logger.SysLogf("migrated %d plaintext env keys to QC! encrypted format", migrateCount)
		}
	}

	// Auto-configure channel API keys from environment variables FIRST (before encryption)
	// Must run BEFORE EncryptExistingChannelKeys so it can match plaintext PUT_YOUR patterns
	results := controller.AutoConfigureAllFromEnv()
	for _, r := range results {
		logger.SysLog("auto-configure: " + r)
	}

	// Encrypt any existing plaintext API keys in the DB (including newly injected real keys)
	if err := model.EncryptExistingChannelKeys(); err != nil {
		logger.SysError("encrypt existing channel keys: " + err.Error())
	}

	// Encrypt sensitive options (payment keys, OAuth secrets) in the database
	model.MigrateSensitiveOptions()
	// 鍚姩鏈堢粨瀹氭椂鍣細娆℃湀 1 鏃ュ噷鏅?2:00 鎵ц
	safeGoWithRestart("monthly-settlement-cron", func() {
		// 鍒濆寤惰繜 30 绉掞紝纭繚鏈嶅姟瀹屽叏灏辩华
		time.Sleep(30 * time.Second)
		for {
			now := time.Now()
			// 璁＄畻涓嬩竴娆℃墽琛屾椂闂达細褰撴湀/涓嬫湀 1 鏃ュ噷鏅?2:00
			next := time.Date(now.Year(), now.Month(), 1, 2, 0, 0, 0, now.Location())
			if !now.Before(next) {
				// 褰撳墠鏃堕棿宸茶繃鏈湀 1 鏃?2:00锛屾帹鍒颁笅涓湀
				next = time.Date(now.Year(), now.Month()+1, 1, 2, 0, 0, 0, now.Location())
			}
			if next.Before(now) {
				// 鏈堜唤婧㈠嚭鐨勫厹搴曞鐞嗭紙璺ㄥ勾锛?
				next = time.Date(now.Year()+1, 1, 1, 2, 0, 0, 0, now.Location())
			}

			sleepDuration := next.Sub(now)
			logger.SysLogf("[CRON] next monthly batch at %s (in %v)", next.Format("2006-01-02 15:04"), sleepDuration.Round(time.Second))
			timer := time.NewTimer(sleepDuration)
			<-timer.C
			timer.Stop()

			logger.SysLog("[CRON] auto settling monthly platform fees...")
			model.AutoSettleMonthlyFees()
			logger.SysLog("[CRON] monthly platform fee settlement completed")

			logger.SysLog("[CRON] fetching official reference pricing...")
			service.FetchOfficialPricing()
			logger.SysLog("[CRON] official reference pricing update completed")

			logger.SysLog("[CRON] fetching brand rankings...")
			service.FetchBrandRankings()
			logger.SysLog("[CRON] brand rankings update completed")

			logger.SysLog("[CRON] syncing popular AI apps...")
			service.SyncPopularApps(context.Background())
			logger.SysLog("[CRON] popular AI apps sync completed")
			// cleanup expired idempotency keys
			// 娓呯悊杩囨湡骞傜瓑閿?
			if err := model.CleanupExpiredIdempotencyKeys(); err != nil {
				logger.SysLog("[CRON] idempotency key cleanup: " + err.Error())
			}
		}
	})


	// 鈹€鈹€ Startup security audit 鈹€鈹€
	if os.Getenv("SESSION_SECRET") == "" || os.Getenv("SESSION_SECRET") == "test-session-secret-for-local-dev-only" {
		logger.SysWarn("[SECURITY] SESSION_SECRET is default/empty. Set a strong random value in production.")
	}
	if pwd := os.Getenv("INITIAL_ROOT_PASSWORD"); pwd == "admin123456" || pwd == "" {
		logger.SysWarn("[SECURITY] INITIAL_ROOT_PASSWORD is the default. Change it after first login.")
	}

	// Env-based admin password reset (emergency)
	if os.Getenv("RESET_ADMIN_PASSWORD") != "" {
		newPwd := os.Getenv("RESET_ADMIN_PASSWORD")
		hashed, err := common.Password2Hash(newPwd)
		if err == nil {
			var adminUser model.User
			if model.DB.Where("role >= ?", 10).First(&adminUser).Error == nil {
				model.DB.Model(&adminUser).Update("password", hashed)
				logger.SysLog("[SECURITY] admin password reset via env var")
			}
		}
	}

	var err error
	err = model.CreateRootAccountIfNeed()
	if err != nil {
		logger.FatalLog("database init error: " + err.Error())
	}
	defer func() {
		err := model.CloseDB()
		if err != nil {
			logger.FatalLog("failed to close database: " + err.Error())
		}
	}()

	// Initialize Redis
	err = common.InitRedisClient()
	if err != nil {
		logger.FatalLog("failed to initialize Redis: " + err.Error())
	}

	// 鈹€鈹€ If slave node, start cascade client 鈹€鈹€
	if !config.IsMasterNode {
		if config.CascadeMasterURL == "" {
			logger.FatalLog("slave node requires CASCADE_MASTER_URL")
		}
		if config.CascadeNodeName == "" {
			logger.FatalLog("slave node requires CASCADE_NODE_NAME")
		}
		if config.CascadeRegion == "" {
			logger.FatalLog("slave node requires CASCADE_REGION")
		}
		go service.StartCascadeClient(service.CascadeConfig{
			MasterURL: config.CascadeMasterURL,
			NodeName:  config.CascadeNodeName,
			Region:    config.CascadeRegion,
		})
		logger.SysLog("quantumclaw running in SLAVE mode (cascade client started)")
		goto AFTER_SLAVE_SETUP
	}

	// Initialize options
	model.InitOptionMap()
	logger.SysLog(fmt.Sprintf("using theme %s", config.Theme))

	if common.RedisEnabled {
		// for compatibility with old versions
		config.MemoryCacheEnabled = true
	}
	if config.MemoryCacheEnabled {
		logger.SysLog("memory cache enabled")
		logger.SysLog(fmt.Sprintf("sync frequency: %d seconds", config.SyncFrequency))
		model.InitChannelCache()
	}
	if config.MemoryCacheEnabled {
		go model.SyncOptions(context.Background(), config.SyncFrequency)
		go model.SyncChannelCache(context.Background(), config.SyncFrequency)
	}
	if os.Getenv("CHANNEL_TEST_FREQUENCY") != "" {
		frequency, err := strconv.Atoi(os.Getenv("CHANNEL_TEST_FREQUENCY"))
		if err != nil {
			logger.FatalLog("failed to parse CHANNEL_TEST_FREQUENCY: " + err.Error())
		}
		go controller.AutomaticallyTestChannels(context.Background(), frequency)
	}
	if os.Getenv("BATCH_UPDATE_ENABLED") == "true" {
		config.BatchUpdateEnabled = true
		logger.SysLog("batch update enabled with interval " + strconv.Itoa(config.BatchUpdateInterval) + "s")
		model.InitBatchUpdater(context.Background())
	}
	if config.EnableMetric {
		logger.SysLog("metric enabled, will disable channel if too much request failed")
	}

	// Initialize quantum RNG if enabled
	if config.QRNGEnabled {
		common.QRNGEnabled = true
		logger.SysLog("quantum random number generator enabled (ANU QRNG)")
	}
	go openai.InitTokenEncoders()
	client.Init()
	go service.TaskPollingLoop()
	go service.StartSubscriptionQuotaResetTask()
	go service.StartHourlySettlement()
	go service.StartRssService(context.Background())
	go service.StartDailyModelSync()
	// Provider 妯″瀷鍒楄〃鑷姩鍚屾锛堟瘡24h浠庡悇渚涘簲鍟嗘媺鍙栨渶鏂版ā鍨嬶級
	go service.NewProviderSyncService(24 * time.Hour).Start()
	// 浼佷笟鐢ㄩ噺缁熻瀹氭椂鑱氬悎锛堟瘡灏忔椂锛?
	go service.StartEnterpriseUsageStatsTask()
	service.LoadCustomOAuthProviders()

AFTER_SLAVE_SETUP:

	// Initialize i18n (only on master; slave has no web UI)
	if config.IsMasterNode {
		if err := i18n.Init(); err != nil {
			logger.FatalLog("failed to initialize i18n: " + err.Error())
		}
	}

	// Initialize HTTP server
	server := gin.New()
	server.Use(gin.Recovery())
	// This will cause SSE not to work!!!
	//server.Use(gzip.Gzip(gzip.DefaultCompression))
	server.Use(middleware.RequestId())
	middleware.SetUpLogger(server)
	// Initialize session store 鈥?Redis 鍏变韩锛堝鏈洪儴缃诧級鎴?Cookie 鍥為€€
	var store sessions.Store
	if common.RedisEnabled {
		redisAddr := os.Getenv("REDIS_HOST")
		redisPort := os.Getenv("REDIS_PORT")
		if redisPort == "" {
			redisPort = "6379"
		}
		redisPassword := os.Getenv("REDIS_PASSWORD")
		store, _ = redis.NewStore(10, "tcp", redisAddr+":"+redisPort, redisPassword,
			[]byte(config.SessionSecret))
		logger.SysLog("session store: Redis (multi-server mode)")
	}
	if store == nil {
		store = cookie.NewStore([]byte(config.SessionSecret))
		logger.SysLog("session store: Cookie (single-server mode)")
	}
	server.Use(sessions.Sessions("session", store))

	router.SetRouter(server, buildFS)
	var port = os.Getenv("PORT")
	if port == "" {
		port = strconv.Itoa(*common.Port)
	}
	logger.SysLogf("server started on http://localhost:%s", port)

	// 浼橀泤鍏抽棴锛氱洃鍚?SIGINT/SIGTERM
	srv := &http.Server{Addr: ":" + port, Handler: server}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.FatalLog("failed to start HTTP server: " + err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.SysLogf("received signal %v, shutting down gracefully...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.SysError("server forced to shutdown: " + err.Error())
	}

	if err := model.CloseDB(); err != nil {
		logger.SysError("failed to close database: " + err.Error())
	}
	logger.SysLog("server exited")
}
