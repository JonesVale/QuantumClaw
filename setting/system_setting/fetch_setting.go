package system_setting

// FetchSetting 由 system_setting.go 统一导出
// 此文件保留用于向后兼容引用路径

import "encoding/json"

// ==================== FetchSetting 核心结构体 ====================

// FetchSetting 全局请求转发与安全设置
type FetchSetting struct {
	// 服务器基础 URL（用于支付回调等）
	BaseURL string `json:"base_url"`

	// SSRF 防护
	EnableSSRFProtection bool     `json:"enable_ssrf_protection"`
	AllowPrivateIp        bool     `json:"allow_private_ip"`
	DomainFilterMode      string   `json:"domain_filter_mode"`    // "allowlist" | "denylist"
	DomainList            []string `json:"domain_list"`
	IpFilterMode          string   `json:"ip_filter_mode"`        // "allowlist" | "denylist"
	IpList                []string `json:"ip_list"`
	AllowedPorts          []int    `json:"allowed_ports"`         // 允许的目标端口

	// Waffo Pancake 支付配置
	WaffoPancakeEnabled        bool   `json:"waffo_pancake_enabled"`
	WaffoPancakeMerchantID     string `json:"waffo_pancake_merchant_id"`
	WaffoPancakeStoreID       string `json:"waffo_pancake_store_id"`
	WaffoPancakeProductID     string `json:"waffo_pancake_product_id"`
	WaffoPancakeCurrency      string `json:"waffo_pancake_currency"`
	WaffoPancakePrivateKey    string `json:"-"`                             // 不序列化
	WaffoPancakeSandbox       bool   `json:"waffo_pancake_sandbox"`
	WaffoPancakeWebhookSecret string `json:"-"`                             // HMAC 验签密钥，不序列化
	WaffoPancakeWebhookPublicKey string `json:"-"`                          // RSA 公钥，不序列化
	WaffoPancakeWebhookTestKey string `json:"waffo_pancake_webhook_test_key"`
}

// MarshalJSON 脱敏序列化：隐藏私钥
func (s FetchSetting) MarshalJSON() ([]byte, error) {
	type alias FetchSetting
	aux := struct {
		alias
		WaffoPancakePrivateKey    string `json:"waffo_pancake_private_key,omitempty"`
	}{
		alias:                    alias(s),
		WaffoPancakePrivateKey:   "", // 始终隐藏
	}
	return json.Marshal(aux)
}
