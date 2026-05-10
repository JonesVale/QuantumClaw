package system_setting

import (
	"encoding/json"
	"math/big"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

// ==================== SSRF 防护配置 ====================

// PasskeySetting Passkey/WebAuthn 设置
type PasskeySetting struct {
	EnablePasskey  bool   `json:"enable_passkey"`
	RelyingPartyID string `json:"relying_party_id"` // 通常是域名
	RPName         string `json:"rp_name"`           // 显示名称
}

// DiscordSetting Discord OAuth 设置
type DiscordSetting struct {
	EnableDiscord  bool     `json:"enable_discord"`
	ClientID       string   `json:"client_id"`
	ClientSecret   string   `json:"-"`                // 不序列化
	RedirectURI    string   `json:"redirect_uri"`
	GuildSync      bool     `json:"guild_sync"`      // 同步 Discord 服务器成员
	AllowedGuilds  []string `json:"allowed_guilds"`  // 允许的服务器 ID
}

var (
	fetchSettingVal  *FetchSetting
	fetchSettingMu   sync.RWMutex
	passkeySetting   *PasskeySetting
	discordSetting   *DiscordSetting
	ssrfSettingMu    sync.RWMutex
	passkeyMu         sync.RWMutex
	discordMu         sync.RWMutex
)

// ==================== FetchSetting 单例访问 ====================

func GetFetchSetting() *FetchSetting {
	fetchSettingMu.RLock()
	defer fetchSettingMu.RUnlock()
	if fetchSettingVal == nil {
		fetchSettingVal = &FetchSetting{
			EnableSSRFProtection: true,
			AllowPrivateIp:       false,
			DomainFilterMode:     "denylist",
			IpFilterMode:         "denylist",
			DomainList:           []string{},
			IpList:               []string{},
			AllowedPorts:         []int{},
		}
	}
	return fetchSettingVal
}

func SetFetchSetting(s *FetchSetting) {
	fetchSettingMu.Lock()
	defer fetchSettingMu.Unlock()
	if s == nil {
		s = &FetchSetting{}
	}
	fetchSettingVal = s
}

func ParseFetchSetting(data string) (*FetchSetting, error) {
	if data == "" {
		return &FetchSetting{
			EnableSSRFProtection: true,
			AllowPrivateIp:       false,
			DomainFilterMode:     "denylist",
			IpFilterMode:         "denylist",
			DomainList:           []string{},
			IpList:               []string{},
			AllowedPorts:         []int{},
		}, nil
	}
	var s FetchSetting
	if err := json.Unmarshal([]byte(data), &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// ==================== PasskeySetting 单例访问 ====================

func GetPasskeySetting() *PasskeySetting {
	passkeyMu.RLock()
	defer passkeyMu.RUnlock()
	if passkeySetting == nil {
		passkeySetting = &PasskeySetting{
			EnablePasskey: false,
			RPName:         "QuantumClaw",
		}
	}
	return passkeySetting
}

func SetPasskeySetting(s *PasskeySetting) {
	passkeyMu.Lock()
	defer passkeyMu.Unlock()
	if s == nil {
		s = &PasskeySetting{}
	}
	passkeySetting = s
}

func ParsePasskeySetting(data string) (*PasskeySetting, error) {
	if data == "" {
		return &PasskeySetting{
			EnablePasskey: false,
			RPName:        "QuantumClaw",
		}, nil
	}
	var s PasskeySetting
	if err := json.Unmarshal([]byte(data), &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// ==================== DiscordSetting 单例访问 ====================

func GetDiscordSetting() *DiscordSetting {
	discordMu.RLock()
	defer discordMu.RUnlock()
	if discordSetting == nil {
		discordSetting = &DiscordSetting{
			EnableDiscord: false,
			AllowedGuilds: []string{},
		}
	}
	return discordSetting
}

func SetDiscordSetting(s *DiscordSetting) {
	discordMu.Lock()
	defer discordMu.Unlock()
	if s == nil {
		s = &DiscordSetting{}
	}
	discordSetting = s
}

func ParseDiscordSetting(data string) (*DiscordSetting, error) {
	if data == "" {
		return &DiscordSetting{
			EnableDiscord: false,
			AllowedGuilds: []string{},
		}, nil
	}
	var s DiscordSetting
	if err := json.Unmarshal([]byte(data), &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// ==================== SSRF 防护检查逻辑 ====================

// privateIPBlocks 私有 IP CIDR 块
var privateIPBlocks []*net.IPNet

func init() {
	for _, cidr := range []string{
		"127.0.0.0/8",   // IPv4 loopback
		"10.0.0.0/8",    // RFC1918
		"172.16.0.0/12", // RFC1918
		"192.168.0.0/16", // RFC1918
		"169.254.0.0/16", // RFC3927 link-local
		"::1/128",       // IPv6 loopback
		"fc00::/7",      // IPv6 unique local
		"fe80::/10",     // IPv6 link-local
	} {
		_, block, _ := net.ParseCIDR(cidr)
		privateIPBlocks = append(privateIPBlocks, block)
	}
}

func isPrivateIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

// ValidateURL 检查 URL 是否符合 SSRF 防护策略
// 返回 (ok bool, reason string)
func ValidateURL(rawURL string) (bool, string) {
	setting := GetFetchSetting()
	if !setting.EnableSSRFProtection {
		return true, ""
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return false, "invalid URL"
	}

	host := u.Hostname()
	if host == "" {
		return false, "empty host"
	}

	// 如果是 IP 地址
	ip := net.ParseIP(host)
	if ip != nil {
		if isPrivateIP(ip) && !setting.AllowPrivateIp {
			return false, "private IP not allowed: " + host
		}
		if setting.IpFilterMode == "allowlist" {
			for _, allowed := range setting.IpList {
				if allowed == host {
					ok, reason := checkPort(u.Port(), setting)
					return ok, reason
				}
			}
			return false, "IP not in allowlist: " + host
		}
		if setting.IpFilterMode == "denylist" {
			for _, denied := range setting.IpList {
				if denied == host {
					return false, "IP in denylist: " + host
				}
			}
		}
		return checkPort(u.Port(), setting)
	}

	// 域名: DNS 解析后检查私有 IP
	addrs, err := net.LookupHost(host)
	if err == nil && len(addrs) > 0 {
		for _, addr := range addrs {
			resolvedIP := net.ParseIP(addr)
			if isPrivateIP(resolvedIP) && !setting.AllowPrivateIp {
				return false, "domain resolves to private IP: " + host
			}
		}
	}

	// 域名 allowlist/denylist
	if setting.DomainFilterMode == "allowlist" {
		for _, allowed := range setting.DomainList {
			if globMatch(allowed, host) {
				return checkPort(u.Port(), setting)
			}
		}
		return false, "domain not in allowlist: " + host
	}
	if setting.DomainFilterMode == "denylist" {
		for _, denied := range setting.DomainList {
			if globMatch(denied, host) {
				return false, "domain in denylist: " + host
			}
		}
	}

	return checkPort(u.Port(), setting)
}

func checkPort(port string, setting *FetchSetting) (bool, string) {
	if port == "" || len(setting.AllowedPorts) == 0 {
		return true, ""
	}
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return false, "invalid port: " + port
	}
	for _, allowed := range setting.AllowedPorts {
		if portNum == allowed {
			return true, ""
		}
	}
	return false, "port not allowed: " + port
}

func globMatch(pattern, host string) bool {
	if pattern == host {
		return true
	}
	// 支持通配符 *.example.com
	if strings.HasPrefix(pattern, "*.") {
		suffix := pattern[2:]
		return strings.HasSuffix(host, suffix) || host == suffix[1:]
	}
	return false
}

// SafeHostHeader 检测 Host header 攻击
func SafeHostHeader(host string) (bool, string) {
	setting := GetFetchSetting()
	if !setting.EnableSSRFProtection {
		return true, ""
	}
	ip := net.ParseIP(host)
	if ip != nil && isPrivateIP(ip) && !setting.AllowPrivateIp {
		return false, "Host header contains private IP"
	}
	return true, ""
}

// ==================== IP 范围解析辅助 ====================

// ParseCIDRRange 解析 CIDR 或单独 IP，支持 ! 前缀表示排除
func ParseCIDRRange(entries []string) (allow []*net.IPNet, deny []*net.IPNet, err error) {
	for _, entry := range entries {
		negate := strings.HasPrefix(entry, "!")
		val := strings.TrimPrefix(entry, "!")
		val = strings.TrimSpace(val)

		_, ipNet, err := net.ParseCIDR(val)
		if err != nil {
			// 尝试解析为单个 IP
			ip := net.ParseIP(val)
			if ip == nil {
				return nil, nil, err
			}
			if negate {
				deny = append(deny, &net.IPNet{IP: ip, Mask: ip.DefaultMask()})
			} else {
				allow = append(allow, &net.IPNet{IP: ip, Mask: ip.DefaultMask()})
			}
			continue
		}
		if negate {
			deny = append(deny, ipNet)
		} else {
			allow = append(allow, ipNet)
		}
	}
	return allow, deny, nil
}

// BigInt to avoid import conflict
var _ = big.NewInt(0)

// ==================== 性能监控设置 ====================

// PerformanceSetting 性能监控设置
type PerformanceSetting struct {
	// 是否启用 Prometheus 指标端点
	EnablePrometheusMetrics bool `json:"enable_prometheus_metrics"`
	// Prometheus 端点路径
	PrometheusPath string `json:"prometheus_path"`
	// 是否启用运行时性能日志
	EnableRuntimeLogs bool `json:"enable_runtime_logs"`
	// GC 日志阈值（MB）
	GCLogThresholdMB int `json:"gc_log_threshold_mb"`
}

var (
	performanceSetting   *PerformanceSetting
	performanceSettingMu sync.RWMutex
)

// GetPerformanceSetting 获取性能监控设置（单例）
func GetPerformanceSetting() *PerformanceSetting {
	performanceSettingMu.RLock()
	defer performanceSettingMu.RUnlock()
	if performanceSetting == nil {
		performanceSetting = &PerformanceSetting{
			EnablePrometheusMetrics: false,
			PrometheusPath:         "/metrics",
			EnableRuntimeLogs:      true,
			GCLogThresholdMB:       50,
		}
	}
	return performanceSetting
}

// SetPerformanceSetting 更新性能监控设置
func SetPerformanceSetting(s *PerformanceSetting) {
	performanceSettingMu.Lock()
	defer performanceSettingMu.Unlock()
	if s == nil {
		s = &PerformanceSetting{}
	}
	performanceSetting = s
}

// ParsePerformanceSetting 从 JSON 解析性能监控设置
func ParsePerformanceSetting(data string) (*PerformanceSetting, error) {
	var s PerformanceSetting
	if data == "" {
		return &PerformanceSetting{
			EnablePrometheusMetrics: false,
			PrometheusPath:         "/metrics",
			EnableRuntimeLogs:      true,
			GCLogThresholdMB:       50,
		}, nil
	}
	if err := json.Unmarshal([]byte(data), &s); err != nil {
		return nil, err
	}
	return &s, nil
}
