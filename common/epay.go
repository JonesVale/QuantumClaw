package common

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/quantumclaw/quantumclaw/common/logger"
)

// ==================== Epay 易支付客户端 ====================

// EpayConfig Epay配置
type EpayConfig struct {
	PID     string // 商户ID
	Key     string // 商户密钥
	Address string // 支付平台地址 (如 https://pay.example.com)
}

// EpayParams 支付参数
type EpayParams struct {
	PID        string  `json:"pid"`
	Type       string  `json:"type"`
	OutTradeNo string  `json:"out_trade_no"`
	NotifyURL  string  `json:"notify_url"`
	ReturnURL  string  `json:"return_url"`
	Name       string  `json:"name"`
	Money      float64 `json:"money"`
	ClientIP   string  `json:"clientip"`
}

// GetEpayConfig 获取Epay配置
func GetEpayConfig() *EpayConfig {
	ps := GetPaymentSetting()
	if !ps.EpayEnabled || ps.EpayId == "" || ps.EpayKey == "" || ps.EpayAddress == "" {
		return nil
	}
	return &EpayConfig{
		PID:     ps.EpayId,
		Key:     ps.EpayKey,
		Address: strings.TrimRight(ps.EpayAddress, "/"),
	}
}

// EpayPaymentType 支付类型常量
const (
	EpayTypeAlipay = "alipay" // 支付宝
	EpayTypeWxpay  = "wxpay"  // 微信支付
	EpayTypeQQPay  = "qqpay"  // QQ钱包
)

// epayBuildSign 构建Epay签名
// 规则: 参数按key排序 → 拼接为 key=value&key=value → 末尾追加商户密钥 → MD5
func epayBuildSign(params map[string]string, key string) string {
	// 排序key
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "sign" || k == "sign_type" || params[k] == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建签名字符串
	var signParts []string
	for _, k := range keys {
		signParts = append(signParts, fmt.Sprintf("%s=%s", k, params[k]))
	}
	signStr := strings.Join(signParts, "&") + key

	// MD5
	hash := md5.Sum([]byte(signStr))
	return hex.EncodeToString(hash[:])
}

// BuildEpayPayURL 构建Epay支付URL
// paymentType: alipay/wxpay/qqpay
// clientIP: 用户IP
func BuildEpayPayURL(cfg *EpayConfig, params EpayParams) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("Epay未配置")
	}

	payParams := map[string]string{
		"pid":          cfg.PID,
		"type":         params.Type,
		"out_trade_no": params.OutTradeNo,
		"notify_url":   params.NotifyURL,
		"return_url":   params.ReturnURL,
		"name":         params.Name,
		"money":        fmt.Sprintf("%.2f", params.Money),
		"clientip":     params.ClientIP,
	}

	// 生成签名
	payParams["sign"] = epayBuildSign(payParams, cfg.Key)
	payParams["sign_type"] = "MD5"

	// 构建URL参数
	q := url.Values{}
	for k, v := range payParams {
		q.Set(k, v)
	}

	// 优先使用mapi.php (API提交), 兼容submit.php
	payURL := fmt.Sprintf("%s/mapi.php?%s", cfg.Address, q.Encode())

	logger.SysLog(fmt.Sprintf("Epay支付链接生成: trade_no=%s type=%s amount=%s", 
		params.OutTradeNo, params.Type, payParams["money"]))

	return payURL, nil
}

// VerifyEpaySign 验证Epay回调签名
func VerifyEpaySign(params map[string]string, key string) bool {
	if params["sign"] == "" {
		return false
	}
	expectedSign := epayBuildSign(params, key)
	actualSign := params["sign"]
	return strings.EqualFold(expectedSign, actualSign)
}

// ParseEpayNotifyParams 解析Epay回调参数 (从URL query或form body)
func ParseEpayNotifyParams(queryStr string) map[string]string {
	values, err := url.ParseQuery(queryStr)
	if err != nil {
		return nil
	}
	params := make(map[string]string)
	for k, v := range values {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}
	return params
}
