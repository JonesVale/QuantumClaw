package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/quantumclaw/quantumclaw/common/env"

	"github.com/google/uuid"
	_ "github.com/joho/godotenv/autoload"
)

var SystemName = "QuantumClaw"
var ServerAddress = func() string {
	if addr := os.Getenv("SERVER_ADDR"); addr != "" {
		port := os.Getenv("PORT")
		if port == "" {
			port = "3666"
		}
		return "http://localhost:" + port
	}
	return os.Getenv("SERVER_ADDRESS")
}()
var Footer = ""
var Logo = "/logo.webp"
var TopUpLink = ""
var ChatLink = ""
var QuotaPerUnit = 500 * 1000.0 // $0.002 / 1K tokens
var PlatformFeeRate = 0.1000 // default 10% platform fee
var DisplayInCurrencyEnabled = true
var DisplayTokenStatEnabled = true

// Any options with "Secret", "Token" in its key won't be return by GetOptions

var SessionSecret = func() string {
	if s := os.Getenv("SESSION_SECRET"); s != "" {
		return s
	}
	return uuid.New().String()
}()

var OptionMap map[string]string
var OptionMapRWMutex sync.RWMutex

var ItemsPerPage = 10
var MaxRecentItems = 100

var PasswordLoginEnabled = true
var PasswordRegisterEnabled = true
var EmailVerificationEnabled = false
var GitHubOAuthEnabled = false
var OidcEnabled = false
var WeChatAuthEnabled = false
var DiscordOAuthEnabled = false
var TurnstileCheckEnabled = false
var RegisterEnabled = true

var EmailDomainRestrictionEnabled = false
var EmailDomainWhitelist = []string{
	"gmail.com",
	"163.com",
	"126.com",
	"qq.com",
	"outlook.com",
	"hotmail.com",
	"icloud.com",
	"yahoo.com",
	"foxmail.com",
}

// PasswordStrength configuration
var PasswordMinLength = env.Int("PASSWORD_MIN_LENGTH", 6)
var PasswordRequireUpper = os.Getenv("PASSWORD_REQUIRE_UPPER") != "false"
var PasswordRequireNumber = os.Getenv("PASSWORD_REQUIRE_NUMBER") != "false"
var PasswordRequireSpecial = os.Getenv("PASSWORD_REQUIRE_SPECIAL") == "true"

var DebugEnabled = strings.ToLower(os.Getenv("DEBUG")) == "true"
var DebugSQLEnabled = strings.ToLower(os.Getenv("DEBUG_SQL")) == "true"
var MemoryCacheEnabled = strings.ToLower(os.Getenv("MEMORY_CACHE_ENABLED")) == "true"

var LogConsumeEnabled = true

var SMTPServer = ""
var SMTPPort = 587
var SMTPAccount = ""
var SMTPFrom = ""
var SMTPToken = ""

var GitHubClientId = ""
var GitHubClientSecret = ""

var LarkClientId = ""
var LarkClientSecret = ""

var OidcClientId = ""
var OidcClientSecret = ""
var OidcWellKnown = ""
var OidcAuthorizationEndpoint = ""
var OidcTokenEndpoint = ""
var OidcUserinfoEndpoint = ""

var AlipayOAuthEnabled = false
var AlipayAppId = ""
var AlipayPrivateKey = ""

var WeChatServerAddress = ""
var WeChatServerToken = ""
var WeChatAccountQRCodeImageURL = ""

var DiscordClientId = ""
var DiscordClientSecret = ""

var LinuxDOOAuthEnabled = false
var LinuxDOClientId = ""
var LinuxDOClientSecret = ""

var TelegramOAuthEnabled = false
var TelegramBotToken = ""
var TelegramBotUsername = ""

var WebAuthnEnabled = false
var WebAuthnRPDisplayName = "QuantumClaw"
var WebAuthnRPID = ""       // 留空则自动从 ServerAddress 提取
var WebAuthnOrigin = ""     // 留空则自动从 ServerAddress 提取

var MessagePusherAddress = ""
var MessagePusherToken = ""

var TurnstileSiteKey = ""
var TurnstileSecretKey = ""

var QuotaForNewUser int64 = 0 // 新用户注册赠送配额(0=关闭), 平台不承担赠送成本
var QuotaForInviter int64 = 0 // 邀请人奖励配额(0=关闭)
var QuotaForInvitee int64 = 0  // 被邀请人奖励配额(0=关闭)
var NewUserTrialBalance int64 = 0 // 新用户注册赠送试用金（分），默认 0 = 关闭
var ChannelDisableThreshold = 5.0
var AutomaticDisableChannelEnabled = true
var AutomaticEnableChannelEnabled = false
var QuotaRemindThreshold int64 = 1000
var PreConsumedQuota int64 = 500
var ApproximateTokenEnabled = false
var DebtDisableThreshold int64 = 50000 // 欠费超过 ¥500 自动禁用账号
var RetryTimes = 2

// MinChannelOwnerBalanceCents 渠道选择时的最低余额阈值（分）
// 当渠道商（channel owner）的现金余额低于此值时，
// 分发器会优先跳过该渠道，选择余额充足的其他渠道。
// 设为 0 表示不进行余额预检（默认行为，向后兼容）。
// 建议生产环境设为 50000（即 ¥500），防止选中余额即将耗尽的渠道。
var MinChannelOwnerBalanceCents = func() int64 {
	if v := os.Getenv("MIN_CHANNEL_OWNER_BALANCE_CENTS"); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return parsed
		}
	}
	return 0 // 默认不启用余额预检（向后兼容）
}()

var RootUserEmail = ""

var IsMasterNode = os.Getenv("NODE_TYPE") != "slave"

// Cascade config (used when NODE_TYPE=slave)
var CascadeMasterURL = os.Getenv("CASCADE_MASTER_URL")
var CascadeNodeName = os.Getenv("CASCADE_NODE_NAME")
var CascadeRegion = os.Getenv("CASCADE_REGION")

// Web search API keys
var BingSearchAPIKey = os.Getenv("BING_SEARCH_API_KEY")
var SerpAPIKey = os.Getenv("SERPAPI_API_KEY")

// Geo service API keys
var AmapAPIKey = os.Getenv("AMAP_API_KEY")
var AmapGeoCodeKey = os.Getenv("AMAP_GEOCODE_KEY")
var GoogleMapsAPIKey = os.Getenv("GOOGLE_MAPS_API_KEY")

// ForceHTTPS when true, HTTP requests are redirected to HTTPS
var ForceHTTPS = strings.ToLower(os.Getenv("FORCE_HTTPS")) == "true"

// ==================== 生产环境强化配置 ====================
// CryptoSecret: 共享Redis数据加密密钥(多机部署必须设置)
// 使用 init() 加载确保在 godotenv 加载之后（godotenv/autoload 的 init 优先于本包 var 初始化）
var CryptoSecret string

func init() {
	CryptoSecret = os.Getenv("CRYPTO_SECRET")
	// 如果环境变量为空，尝试从 .env 文件读取
	// 兜底 godotenv/autoload 未及时加载的情况
	if CryptoSecret == "" {
		data, err := os.ReadFile(".env")
		if err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "CRYPTO_SECRET=") {
					CryptoSecret = strings.TrimPrefix(line, "CRYPTO_SECRET=")
					break
				}
			}
		}
	}
	// 如果仍为空，打印警告（但不阻止启动，main() 会生成新密钥）
	if CryptoSecret == "" {
		fmt.Println("[WARN] CRYPTO_SECRET not set, will be generated at startup")
	}
}

// StreamingTimeout: SSE流式响应超时(秒),默认300
var StreamingTimeout = env.Int("STREAMING_TIMEOUT", 300)

// MaxRequestBodyMB: 请求体大小限制(MB),超限返回413
var MaxRequestBodyMB = env.Int("MAX_REQUEST_BODY_MB", 32)

// ErrorLogEnabled: 独立错误日志开关
var ErrorLogEnabled = strings.ToLower(os.Getenv("ERROR_LOG_ENABLED")) == "true"

// QRNGEnabled: use quantum random number generator for enhanced security (default: true)
var QRNGEnabled = os.Getenv("QRNG_ENABLED") != "false"
// QRNGSourceURL: quantum random number source URL (default: ANU, can point to domestic source)
var QRNGSourceURL = func() string {
	if v := os.Getenv("QRNG_SOURCE_URL"); v != "" {
		return v
	}
	return "https://qrng.anu.edu.au/API/jsonI.php"
}()

var requestInterval, _ = strconv.Atoi(os.Getenv("POLLING_INTERVAL"))
var RequestInterval = time.Duration(requestInterval) * time.Second

var SyncFrequency = env.Int("SYNC_FREQUENCY", 10*60) // unit is second

var BatchUpdateEnabled = false
var BatchUpdateInterval = env.Int("BATCH_UPDATE_INTERVAL", 5)

var RelayTimeout = env.Int("RELAY_TIMEOUT", 0) // unit is second

var GeminiSafetySetting = env.String("GEMINI_SAFETY_SETTING", "BLOCK_NONE")

var Theme = env.String("THEME", "default")
var ValidThemes = map[string]bool{
	"default": true,
	"berry":   true,
	"air":     true,
}

// All duration's unit is seconds
// Shouldn't larger then RateLimitKeyExpirationDuration
var (
	GlobalApiRateLimitNum            = env.Int("GLOBAL_API_RATE_LIMIT", 480)
	GlobalApiRateLimitDuration int64 = 3 * 60

	GlobalWebRateLimitNum            = env.Int("GLOBAL_WEB_RATE_LIMIT", 240)
	GlobalWebRateLimitDuration int64 = 3 * 60

	UploadRateLimitNum            = 10
	UploadRateLimitDuration int64 = 60

	DownloadRateLimitNum            = 10
	DownloadRateLimitDuration int64 = 60

	CriticalRateLimitNum            = 20
	CriticalRateLimitDuration int64 = 20 * 60
)

var RateLimitKeyExpirationDuration = 20 * time.Minute

var EnableMetric = env.Bool("ENABLE_METRIC", true)
var MetricQueueSize = env.Int("METRIC_QUEUE_SIZE", 10)
var MetricSuccessRateThreshold = env.Float64("METRIC_SUCCESS_RATE_THRESHOLD", 0.8)
var MetricSuccessChanSize = env.Int("METRIC_SUCCESS_CHAN_SIZE", 1024)
var MetricFailChanSize = env.Int("METRIC_FAIL_CHAN_SIZE", 128)

var InitialRootToken = os.Getenv("INITIAL_ROOT_TOKEN")

var InitialRootAccessToken = os.Getenv("INITIAL_ROOT_ACCESS_TOKEN")

var InitialRootPassword = os.Getenv("INITIAL_ROOT_PASSWORD")

var GeminiVersion = env.String("GEMINI_VERSION", "v1")

var OnlyOneLogFile = env.Bool("ONLY_ONE_LOG_FILE", false)

var RelayProxy = env.String("RELAY_PROXY", "")
var UserContentRequestProxy = env.String("USER_CONTENT_REQUEST_PROXY", "")
var UserContentRequestTimeout = env.Int("USER_CONTENT_REQUEST_TIMEOUT", 30)

var EnforceIncludeUsage = env.Bool("ENFORCE_INCLUDE_USAGE", false)
var TestPrompt = env.String("TEST_PROMPT", "Output only your specific model name with no additional text.")

// ==================== 安全配置 ====================

// CORS 允许的 Origins（逗号分隔）
// 例如: "https://example.com,https://app.example.com"
// 留空则禁止跨域请求（仅允许同源）
var AllowedOrigins = env.StringSlice("ALLOWED_ORIGINS", []string{})
var CORSAllowCredentials = env.Bool("CORS_ALLOW_CREDENTIALS", true)
var CORSMaxAge = env.Int("CORS_MAX_AGE", 86400)

func GetAllowedOrigins() []string {
	origins := AllowedOrigins
	if len(origins) == 0 {
		return []string{"*"}
	}
	return origins
}

// CSP 配置
var CSPReportOnly = env.Bool("CSP_REPORT_ONLY", false) // CSP 是否只报告不阻止

// HSTS 配置
var HSTSMaxAge = env.Int("HSTS_MAX_AGE", 31536000) // 默认 1 年 (秒)
var HSTSIncludeSubDomains = env.Bool("HSTS_INCLUDE_SUBDOMAINS", false)
var HSTSPreload = env.Bool("HSTS_PRELOAD", false)

// X-Frame-Options 配置
// 可选值: "DENY", "SAMEORIGIN"
var XFrameOptions = env.String("X_FRAME_OPTIONS", "DENY")

// X-XSS-Protection 配置
var XSSProtectionEnabled = env.Bool("XSS_PROTECTION_ENABLED", true)

// Referrer-Policy 配置
// 可选值: "no-referrer", "no-referrer-when-downgrade", "origin", "origin-when-cross-origin",
//          "same-origin", "strict-origin", "strict-origin-when-cross-origin", "unsafe-url"
var ReferrerPolicy = env.String("REFERRER_POLICY", "strict-origin-when-cross-origin")

// Permissions-Policy 配置 (原 Feature-Policy)
// 例如: "camera=(), microphone=(), geolocation=()"
var PermissionsPolicy = env.String("PERMISSIONS_POLICY", "camera=(), microphone=(), geolocation=()")

// Pyroscope 持续剖析配置
var PyroscopeURL = os.Getenv("PYROSCOPE_URL")
var PyroscopeAppName = env.String("PYROSCOPE_APP_NAME", "quantumclaw")
var PyroscopeBasicAuthUser = os.Getenv("PYROSCOPE_BASIC_AUTH_USER")
var PyroscopeBasicAuthPassword = os.Getenv("PYROSCOPE_BASIC_AUTH_PASSWORD")

// 支付 Webhook IP 白名单（逗号分隔）
var PaymentWebhookIPWhitelist = os.Getenv("PAYMENT_WEBHOOK_IP_WHITELIST")

// SSRF 防护配置
var EnableSSRFProtection = env.Bool("ENABLE_SSRF_PROTECTION", true)
var SSRFAllowedHosts = env.StringSlice("SSRF_ALLOWED_HOSTS", []string{}) // 允许的 Host 白名单
var SSRFAllowedCIDRs = env.StringSlice("SSRF_ALLOWED_CIDRS", []string{}) // 允许的 CIDR 白名单
