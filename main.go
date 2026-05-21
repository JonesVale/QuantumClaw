package main

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/grafana/pyroscope-go"
	_ "github.com/joho/godotenv/autoload"

	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/client"
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

func main() {
	common.Init()
	logger.SetupLogger()
	logger.SysLogf("QuantumClaw %s started", common.Version)

	// 验证频道类型定义一致性
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
	model.InitLogDB()

	// Initialize language types
	model.InitLanguageTypes()

	// Initialize T_Languages translation tables (seed if empty)
	model.InitLanguageTables()

	// Initialize commission tables
	model.InitCommissionTables()

	// Initialize distributor tables
	model.InitDistributorTables()

	// 启动入驻费自动结算定时器（每小时检查，次月1号凌晨2点执行）
	go func() {
		time.Sleep(30 * time.Second)
		for {
			now := time.Now()
			if now.Day() == 1 && now.Hour() == 2 {
				logger.SysLog("[CRON] auto settling monthly platform fees...")
				model.AutoSettleMonthlyFees()
				logger.SysLog("[CRON] monthly platform fee settlement completed")
			}
			time.Sleep(1 * time.Hour)
		}
	}()

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
		go model.SyncOptions(config.SyncFrequency)
		go model.SyncChannelCache(config.SyncFrequency)
	}
	if os.Getenv("CHANNEL_TEST_FREQUENCY") != "" {
		frequency, err := strconv.Atoi(os.Getenv("CHANNEL_TEST_FREQUENCY"))
		if err != nil {
			logger.FatalLog("failed to parse CHANNEL_TEST_FREQUENCY: " + err.Error())
		}
		go controller.AutomaticallyTestChannels(frequency)
	}
	if os.Getenv("BATCH_UPDATE_ENABLED") == "true" {
		config.BatchUpdateEnabled = true
		logger.SysLog("batch update enabled with interval " + strconv.Itoa(config.BatchUpdateInterval) + "s")
		model.InitBatchUpdater()
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
	go service.StartRssService(context.Background())
	service.LoadCustomOAuthProviders()

	// Initialize i18n
	if err := i18n.Init(); err != nil {
		logger.FatalLog("failed to initialize i18n: " + err.Error())
	}

	// Initialize HTTP server
	server := gin.New()
	server.Use(gin.Recovery())
	// This will cause SSE not to work!!!
	//server.Use(gzip.Gzip(gzip.DefaultCompression))
	server.Use(middleware.RequestId())
	server.Use(middleware.Language())
	middleware.SetUpLogger(server)
	// Initialize session store
	store := cookie.NewStore([]byte(config.SessionSecret))
	server.Use(sessions.Sessions("session", store))

	router.SetRouter(server, buildFS)
	var port = os.Getenv("PORT")
	if port == "" {
		port = strconv.Itoa(*common.Port)
	}
	logger.SysLogf("server started on http://localhost:%s", port)

	// 优雅关闭：监听 SIGINT/SIGTERM
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