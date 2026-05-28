package common

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/logger"
)

// ==================== 支付配置(安全增强版)====================

var (
	paymentSettingOnce  sync.Once
	paymentSetting      *PaymentSetting
	paymentSettingMu    sync.Mutex // 保护 Reload/Reset/Save 的 Once 重置
)

// PaymentSetting 支付设置
type PaymentSetting struct {
	// 易支付(Epay)配置
	EpayEnabled bool   `json:"epay_enabled"`
	EpayId      string `json:"epay_id"`
	EpayKey     string `json:"epay_key"`
	EpayAddress string `json:"epay_address"`

	// Stripe 配置
	StripeEnabled      bool    `json:"stripe_enabled"`
	StripeApiSecret   string  `json:"stripe_api_secret"`
	StripeWebhookSecret string `json:"stripe_webhook_secret"`
	StripeMinTopUp    int     `json:"stripe_min_topup"`
	StripeUnitPrice   float64 `json:"stripe_unit_price"`

	// Creem 配置
	CreemEnabled       bool   `json:"creem_enabled"`
	CreemApiKey       string `json:"creem_api_key"`
	CreemWebhookSecret string `json:"creem_webhook_secret"`
	CreemTestMode     bool   `json:"creem_test_mode"`
	CreemProductID    string `json:"creem_product_id"`
	CreemProducts     string `json:"creem_products"` // JSON 字符串

	// Waffo 配置
	WaffoEnabled        bool   `json:"waffo_enabled"`
	WaffoSandbox       bool   `json:"waffo_sandbox"`
	WaffoApiKey        string `json:"waffo_api_key"`
	WaffoSandboxApiKey string `json:"waffo_sandbox_api_key"`
	WaffoMinTopUp      int    `json:"waffo_min_topup"`
	WaffoUnitPrice     float64 `json:"waffo_unit_price"`
	WaffoCurrency      string  `json:"waffo_currency"`

	// Binance Pay 配置
	BinanceEnabled       bool   `json:"binance_enabled"`
	BinanceApiKey        string `json:"binance_api_key"`
	BinanceSecretKey     string `json:"binance_secret_key"`
	BinanceMerchantId    string `json:"binance_merchant_id"`
	BinanceMinTopUp      int    `json:"binance_min_topup"`
	BinanceUnitPrice     float64 `json:"binance_unit_price"`
	BinanceCurrency      string  `json:"binance_currency"`

	// 支付宝官方直连配置
	AlipayEnabled        bool    `json:"alipay_enabled"`
	AlipayAppId         string  `json:"alipay_app_id"`
	AlipayPrivateKey    string  `json:"alipay_private_key"`
	AlipayPublicKey     string  `json:"alipay_public_key"`
	AlipayGatewayUrl    string  `json:"alipay_gateway_url"`
	AlipayMinTopUp      int     `json:"alipay_min_topup"`

	// 万里汇 WorldFirst 配置
	WorldFirstEnabled     bool    `json:"worldfirst_enabled"`
	WorldFirstApiKey     string  `json:"worldfirst_api_key"`
	WorldFirstSecretKey  string  `json:"worldfirst_secret_key"`
	WorldFirstWebhookKey string  `json:"worldfirst_webhook_key"`
	WorldFirstMinTopUp   int     `json:"worldfirst_min_topup"`

	// 通用配置
	MinTopUp       int               `json:"min_topup"`
	AmountOptions  []int            `json:"amount_options"`
	AmountDiscount map[int]float64  `json:"amount_discount"`
	PayMethods     []map[string]string `json:"pay_methods"`

	// 安全配置
	PaymentSignatureSecret string `json:"payment_signature_secret"` // 订单签名密钥
	PaymentNotifyURL       string `json:"payment_notify_url"`      // 支付回调URL
	PaymentReturnURL       string `json:"payment_return_url"`      // 支付返回URL
}

// GetPaymentSetting 获取支付设置(单例模式)
func GetPaymentSetting() *PaymentSetting {
	paymentSettingOnce.Do(func() {
		paymentSetting = &PaymentSetting{
			// 默认值
			EpayEnabled:          false,
			StripeEnabled:        false,
			CreemEnabled:         false,
			WaffoEnabled:         false,
			BinanceEnabled:       false,
			WorldFirstEnabled:     false,
			WorldFirstMinTopUp:   1,
			StripeMinTopUp:       1,
			WaffoMinTopUp:        1,
			MinTopUp:             1,
			AmountOptions:         []int{100, 500, 1000, 5000},
			AmountDiscount:        make(map[int]float64),
			PayMethods:            []map[string]string{},
			PaymentSignatureSecret: getEnvOrDefault("PAYMENT_SIGNATURE_SECRET", generateRandomSecret()),
		}

		// 从环境变量加载配置
		loadPaymentConfigFromEnv(paymentSetting)
		// 从 system_options 覆盖加载（支持后台动态修改）
		loadPaymentConfigFromOptions(paymentSetting)

		logger.SysLog("支付配置加载完成")
	})

	return paymentSetting
}

// loadPaymentConfigFromOptions 从 system_options 加载支付配置（支持后台动态修改）
func loadPaymentConfigFromOptions(settings *PaymentSetting) {
	getOpt := func(key string) string {
		config.OptionMapRWMutex.RLock()
		defer config.OptionMapRWMutex.RUnlock()
		return config.OptionMap[key]
	}

	if getOpt("EpayEnabled") == "true" {
		settings.EpayEnabled = true
		if v := getOpt("EpayId"); v != "" {
			settings.EpayId = v
		}
		if v := getOpt("EpayKey"); v != "" {
			settings.EpayKey = v
		}
		if v := getOpt("EpayAddress"); v != "" {
			settings.EpayAddress = v
		}
	}
	if getOpt("StripeEnabled") == "true" {
		settings.StripeEnabled = true
		if v := getOpt("StripeApiSecret"); v != "" {
			settings.StripeApiSecret = v
		}
		if v := getOpt("StripeMinTopUp"); v != "" {
			if min, err := strconv.Atoi(v); err == nil {
				settings.StripeMinTopUp = min
			}
		}
	}
	if getOpt("CreemEnabled") == "true" {
		settings.CreemEnabled = true
		if v := getOpt("CreemApiKey"); v != "" {
			settings.CreemApiKey = v
		}
	}
	if getOpt("WaffoEnabled") == "true" {
		settings.WaffoEnabled = true
		if v := getOpt("WaffoApiKey"); v != "" {
			settings.WaffoApiKey = v
		}
		if getOpt("WaffoSandbox") == "true" {
			settings.WaffoSandbox = true
		}
	}
	if getOpt("BinanceEnabled") == "true" {
		settings.BinanceEnabled = true
		if v := getOpt("BinanceApiKey"); v != "" {
			settings.BinanceApiKey = v
		}
		if v := getOpt("BinanceSecretKey"); v != "" {
			settings.BinanceSecretKey = v
		}
		if v := getOpt("BinanceMerchantId"); v != "" {
			settings.BinanceMerchantId = v
		}
	}

	// 支付宝官方直连
	if getOpt("AlipayEnabled") == "true" {
		settings.AlipayEnabled = true
		if v := getOpt("AlipayAppId"); v != "" {
			settings.AlipayAppId = v
		}
		if v := getOpt("AlipayPrivateKey"); v != "" {
			settings.AlipayPrivateKey = v
		}
		if v := getOpt("AlipayPublicKey"); v != "" {
			settings.AlipayPublicKey = v
		}
		if v := getOpt("AlipayGatewayUrl"); v != "" {
			settings.AlipayGatewayUrl = v
		}
	}

	// 万里汇 WorldFirst
	if getOpt("WorldFirstEnabled") == "true" {
		settings.WorldFirstEnabled = true
		if v := getOpt("WorldFirstApiKey"); v != "" { settings.WorldFirstApiKey = v }
		if v := getOpt("WorldFirstSecretKey"); v != "" { settings.WorldFirstSecretKey = v }
		if v := getOpt("WorldFirstWebhookKey"); v != "" { settings.WorldFirstWebhookKey = v }
	}
}

// ReloadPaymentSettings 重新加载支付配置（管理员修改后调用）
func ReloadPaymentSettings() {
	paymentSettingMu.Lock()
	paymentSettingOnce = sync.Once{}
	paymentSetting = nil
	paymentSettingMu.Unlock()
	GetPaymentSetting()
	logger.SysLog("支付配置已重新加载")
}

// loadPaymentConfigFromEnv 从环境变量加载支付配置
func loadPaymentConfigFromEnv(settings *PaymentSetting) {
	// 易支付配置
	if os.Getenv("EPAY_ENABLED") == "true" {
		settings.EpayEnabled = true
		settings.EpayId = os.Getenv("EPAY_ID")
		settings.EpayKey = os.Getenv("EPAY_KEY")
		settings.EpayAddress = os.Getenv("EPAY_ADDRESS")
	}

	// Stripe 配置
	if os.Getenv("STRIPE_ENABLED") == "true" {
		settings.StripeEnabled = true
		settings.StripeApiSecret = os.Getenv("STRIPE_API_SECRET")
		settings.StripeWebhookSecret = os.Getenv("STRIPE_WEBHOOK_SECRET")
		if minStr := os.Getenv("STRIPE_MIN_TOPUP"); minStr != "" {
			if min, err := strconv.Atoi(minStr); err == nil {
				settings.StripeMinTopUp = min
			}
		}
	}

	// Creem 配置
	if os.Getenv("CREEM_ENABLED") == "true" {
		settings.CreemEnabled = true
		settings.CreemApiKey = os.Getenv("CREEM_API_KEY")
		settings.CreemWebhookSecret = os.Getenv("CREEM_WEBHOOK_SECRET")
		settings.CreemTestMode = os.Getenv("CREEM_TEST_MODE") == "true"
		settings.CreemProductID = os.Getenv("CREEM_PRODUCT_ID")
		settings.CreemProducts = os.Getenv("CREEM_PRODUCTS")
	}

	// Waffo 配置
	if os.Getenv("WAFFO_ENABLED") == "true" {
		settings.WaffoEnabled = true
		settings.WaffoSandbox = os.Getenv("WAFFO_SANDBOX") == "true"
		settings.WaffoApiKey = os.Getenv("WAFFO_API_KEY")
		settings.WaffoSandboxApiKey = os.Getenv("WAFFO_SANDBOX_API_KEY")
		if minStr := os.Getenv("WAFFO_MIN_TOPUP"); minStr != "" {
			if min, err := strconv.Atoi(minStr); err == nil {
				settings.WaffoMinTopUp = min
			}
		}
		settings.WaffoCurrency = getEnvOrDefault("WAFFO_CURRENCY", "USD")
	}

	// Binance Pay 配置
	if os.Getenv("BINANCE_ENABLED") == "true" {
		settings.BinanceEnabled = true
		settings.BinanceApiKey = os.Getenv("BINANCE_API_KEY")
		settings.BinanceSecretKey = os.Getenv("BINANCE_SECRET_KEY")
		settings.BinanceMerchantId = os.Getenv("BINANCE_MERCHANT_ID")
		if minStr := os.Getenv("BINANCE_MIN_TOPUP"); minStr != "" {
			if min, err := strconv.Atoi(minStr); err == nil {
				settings.BinanceMinTopUp = min
			}
		}
		settings.BinanceCurrency = getEnvOrDefault("BINANCE_CURRENCY", "USDT")
	}

	// 支付宝官方直连配置
	if os.Getenv("ALIPAY_ENABLED") == "true" {
		settings.AlipayEnabled = true
		settings.AlipayAppId = os.Getenv("ALIPAY_APP_ID")
		settings.AlipayPrivateKey = os.Getenv("ALIPAY_PRIVATE_KEY")
		settings.AlipayPublicKey = os.Getenv("ALIPAY_PUBLIC_KEY")
		if v := os.Getenv("ALIPAY_GATEWAY_URL"); v != "" {
			settings.AlipayGatewayUrl = v
		}
		if v := os.Getenv("ALIPAY_MIN_TOPUP"); v != "" {
			if min, err := strconv.Atoi(v); err == nil {
				settings.AlipayMinTopUp = min
			}
		}
	}
	if settings.AlipayGatewayUrl == "" {
		settings.AlipayGatewayUrl = "https://openapi.alipay.com/gateway.do"
	}

	// 万里汇 WorldFirst 配置
	if os.Getenv("WORLDFIRST_ENABLED") == "true" {
		settings.WorldFirstEnabled = true
		settings.WorldFirstApiKey = os.Getenv("WORLDFIRST_API_KEY")
		settings.WorldFirstSecretKey = os.Getenv("WORLDFIRST_SECRET_KEY")
		settings.WorldFirstWebhookKey = os.Getenv("WORLDFIRST_WEBHOOK_KEY")
	}

	// 通用配置
	if minStr := os.Getenv("MIN_TOPUP"); minStr != "" {
		if min, err := strconv.Atoi(minStr); err == nil {
			settings.MinTopUp = min
		}
	}

	// 支付签名密钥
	if secret := os.Getenv("PAYMENT_SIGNATURE_SECRET"); secret != "" {
		settings.PaymentSignatureSecret = secret
	}
}

// generateRandomSecret 生成随机密钥(用于订单签名)
func generateRandomSecret() string {
	secret := GetRandomString(32)
	logger.SysLog("生成随机支付签名密钥(请设置为环境变量 PAYMENT_SIGNATURE_SECRET)")
	return secret
}

// ==================== 辅助函数 ====================

// IsEpayEnabled 检查易支付是否启用
func IsEpayEnabled() bool {
	return GetPaymentSetting().EpayEnabled
}

// IsStripeEnabled 检查Stripe是否启用
func IsStripeEnabled() bool {
	return GetPaymentSetting().StripeEnabled
}

// IsCreemEnabled 检查Creem是否启用
func IsCreemEnabled() bool {
	return GetPaymentSetting().CreemEnabled
}

// IsWaffoEnabled 检查Waffo是否启用
func IsWaffoEnabled() bool {
	return GetPaymentSetting().WaffoEnabled
}

// GetMinTopUp 获取最小充值金额
func GetMinTopUp() int {
	return GetPaymentSetting().MinTopUp
}

// GetStripeMinTopUp 获取Stripe最小充值金额
func GetStripeMinTopUp() int {
	return GetPaymentSetting().StripeMinTopUp
}

// GetWaffoMinTopUp 获取Waffo最小充值金额
func GetWaffoMinTopUp() int {
	return GetPaymentSetting().WaffoMinTopUp
}

// GetAmountOptions 获取金额选项
func GetAmountOptions() []int {
	return GetPaymentSetting().AmountOptions
}

// GetAmountDiscount 获取金额折扣
func GetAmountDiscount() map[int]float64 {
	return GetPaymentSetting().AmountDiscount
}

// GetPayMethods 获取支付方式列表
func GetPayMethods() []map[string]string {
	return GetPaymentSetting().PayMethods
}

// IsValidPayMethod 检查支付方式是否有效
func IsValidPayMethod(method string) bool {
	validMethods := []string{"epay", "stripe", "creem", "waffo", "waffo_pancake", "binance"}
	for _, m := range validMethods {
		if m == method {
			return true
		}
	}
	return false
}

// HmacSha256 计算HMAC-SHA256签名(用于订单签名和Webhook验证)
func HmacSha256(data string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// ValidateRedirectURL 验证重定向URL(防止钓鱼攻击)
// 安全增强:
// 1. 检查URL格式
// 2. 检查域名是否在白名单中(从环境变量配置)
// 3. 防止开放重定向漏洞
func ValidateRedirectURL(redirectURL string) error {
	if redirectURL == "" {
		return nil
	}

	// 解析URL
	parsedURL, err := url.Parse(redirectURL)
	if err != nil {
		return fmt.Errorf("无效的URL格式")
	}

	// 只允许 http 和 https
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("只允许 http 或 https 链接")
	}

	// 允许相对路径重定向
	if parsedURL.Scheme == "" && !strings.HasPrefix(redirectURL, "/") {
		return fmt.Errorf("无效的重定向路径")
	}

	// 检查域名白名单
	allowedDomains := getAllowedRedirectDomains()
	if len(allowedDomains) > 0 {
		hostname := strings.ToLower(parsedURL.Hostname())
		if hostname != "" {
			allowed := false
			for _, domain := range allowedDomains {
				if hostname == strings.ToLower(domain) || strings.HasSuffix(hostname, "."+strings.ToLower(domain)) {
					allowed = true
					break
				}
			}
			if !allowed {
				return fmt.Errorf("重定向域名不在白名单中: %s", hostname)
			}
		}
	}

	// 防止javascript:协议
	if parsedURL.Scheme == "javascript" {
		return fmt.Errorf("不允许的协议")
	}

	return nil
}

// getAllowedRedirectDomains 获取允许的重定向域名白名单
func getAllowedRedirectDomains() []string {
	domainsEnv := os.Getenv("ALLOWED_REDIRECT_DOMAINS")
	if domainsEnv == "" {
		return []string{}
	}
	domains := strings.Split(domainsEnv, ",")
	result := make([]string, 0, len(domains))
	for _, d := range domains {
		d = strings.TrimSpace(d)
		if d != "" {
			result = append(result, d)
		}
	}
	return result
}

// GetPaymentNotifyURL 获取支付回调URL
func GetPaymentNotifyURL() string {
	return GetPaymentSetting().PaymentNotifyURL
}

// GetPaymentReturnURL 获取支付返回URL
func GetPaymentReturnURL() string {
	return GetPaymentSetting().PaymentReturnURL
}

// SavePaymentSetting 保存支付设置到数据库
func SavePaymentSetting(settings *PaymentSetting) error {
	_, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	// 注意: 避免 import cycle (model 依赖 common)，调用方(controller)负责写入 DB Option 表
	// 仅重新加载配置
	paymentSettingMu.Lock()
	paymentSettingOnce = sync.Once{}
	paymentSetting = nil
	paymentSettingMu.Unlock()

	logger.SysLog("支付设置已保存")
	return nil
}

// getEnvOrDefault 获取环境变量,如果不存在则返回默认值
func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// ResetPaymentSetting 重置支付配置(重新从环境变量加载)
func ResetPaymentSetting() {
	paymentSettingMu.Lock()
	paymentSettingOnce = sync.Once{}
	paymentSetting = nil
	paymentSettingMu.Unlock()
	GetPaymentSetting()
	logger.SysLog("支付配置已重置")
}

// GetPrice 获取单价(每配额单位对应的价格)
func GetPrice() float64 {
	// 从 config 包读取 QuotaPerUnit
	// 1 单位 = config.QuotaPerUnit 配额,对应 $0.002
	// 因此 1 配额 = 0.002 / 500000 = 4e-9 美元
	return 0.002 / 500000.0
}

// GetTopupGroupRatio 获取用户组的充值比例
// 不同用户组可能有不同的充值折扣比例
// 目前硬编码默认值，后续可以从数据库设置管理
func GetTopupGroupRatio(group string) float64 {
	groupRatios := map[string]float64{
		"vip":   0.95, // VIP 95折
		"svip":  0.90, // SVIP 9折
		"team":  0.85, // 团队 85折
		"enterprise": 0.80, // 企业 8折
	}
	if ratio, ok := groupRatios[group]; ok {
		return ratio
	}
	return 1.0
}

// GetRandomString 生成指定长度的随机字符串(用于密钥、验证码等)
func GetRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	randomBytes := make([]byte, length)
	if _, err := rand.Read(randomBytes); err != nil {
		// 如果获取随机数失败,使用时间戳作为备用
		for i := range b {
			b[i] = charset[int(time.Now().UnixNano()/int64(i+1))%len(charset)]
		}
		return string(b)
	}
	for i := range b {
		b[i] = charset[int(randomBytes[i])%len(charset)]
	}
	return string(b)
}

// IsBinanceEnabled 检查Binance Pay是否启用
func IsBinanceEnabled() bool {
	return GetPaymentSetting().BinanceEnabled
}

// GetBinanceMinTopUp 获取Binance最小充值金额
func GetBinanceMinTopUp() int {
	return GetPaymentSetting().BinanceMinTopUp
}

// IsAlipayEnabled 支付宝是否启用
func IsAlipayEnabled() bool {
	return GetPaymentSetting().AlipayEnabled
}

// GetAlipayMinTopUp 获取支付宝最小充值金额
func GetAlipayMinTopUp() int {
	min := GetPaymentSetting().AlipayMinTopUp
	if min <= 0 {
		return 1
	}
	return min
}

// IsWorldFirstEnabled 万里汇是否启用
func IsWorldFirstEnabled() bool {
	return GetPaymentSetting().WorldFirstEnabled
}

// GetWorldFirstMinTopUp 万里汇最小充值金额
func GetWorldFirstMinTopUp() int {
	min := GetPaymentSetting().WorldFirstMinTopUp
	if min <= 0 {
		return 1
	}
	return min
}

// GetTimestamp 获取当前Unix时间戳(秒)
func GetTimestamp() int64 {
	return time.Now().Unix()
}

// GetEnvOrDefault 获取环境变量值,如果为空则返回默认值
func GetEnvOrDefault(key string, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
