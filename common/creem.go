package common

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
)

// ==================== Creem.io 支付集成 ====================

const (
	creemAPI        = "https://api.creem.io"
	creemSandboxAPI = "https://test-api.creem.io"
	// Creem 创建 checkout 路径
	creemCheckoutPath = "/v1/checkouts"
)

// CreemCheckoutRequest Creem 创建 checkout 请求
type CreemCheckoutRequest struct {
	ProductID       string `json:"product_id"`
	SuccessURL      string `json:"success_url"`
	CancelURL       string `json:"cancel_url"`
	WebhookURL      string `json:"webhook_url,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	CustomerEmail   string `json:"customer_email,omitempty"`
	ClientReference string `json:"client_reference_id,omitempty"`
}

// CreemCheckoutResponse Creem 创建 checkout 响应
type CreemCheckoutResponse struct {
	ID          string `json:"id"`
	URL         string `json:"url"`
	Status      string `json:"status"`
	ProductID   string `json:"product_id"`
	Amount      int64  `json:"amount"`
	Currency    string `json:"currency"`
}

// CreemWebhookPayload Creem webhook payload
type CreemWebhookPayload struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

// CreateCreemCheckout 创建 Creem checkout session
// 使用 Creem REST API（非官方，HTTP 接口）
func CreateCreemCheckout(params *StripeCheckoutParams) (checkoutURL string, sessionID string, err error) {
	ps := GetPaymentSetting()

	if ps.CreemApiKey == "" {
		return "", "", fmt.Errorf("Creem API Key 未配置")
	}

	apiBase := creemAPI
	if ps.CreemTestMode {
		apiBase = creemSandboxAPI
	}

	notifyURL := GetPaymentNotifyURL()
	if notifyURL == "" {
		notifyURL = ps.PaymentNotifyURL
	}

	// Creem 需要先创建产品，这里用 metadata 传递自定义金额
	req := CreemCheckoutRequest{
		ProductID:       ps.CreemProductID,
		SuccessURL:      params.SuccessURL,
		CancelURL:       params.CancelURL,
		WebhookURL:      notifyURL,
		CustomerEmail:   params.UserEmail,
		ClientReference: params.TradeNo,
		Metadata: map[string]string{
			"trade_no":   params.TradeNo,
			"amount":     fmt.Sprintf("%.2f", params.PayMoney),
			"quota":      fmt.Sprintf("%d", params.Amount),
			"product":    params.ProductName,
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", "", fmt.Errorf("creem request marshal: %w", err)
	}

	httpReq, err := http.NewRequest("POST", apiBase+creemCheckoutPath, strings.NewReader(string(body)))
	if err != nil {
		return "", "", fmt.Errorf("creem request create: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", ps.CreemApiKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", "", fmt.Errorf("creem api request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	// Creem 可能返回 201 或 200
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("creem api error: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var result CreemCheckoutResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", "", fmt.Errorf("creem response parse: %w", err)
	}

	if result.URL == "" {
		return "", "", fmt.Errorf("creem response missing checkout URL")
	}

	logger.SysLog(fmt.Sprintf("Creem checkout 创建成功: session_id=%s trade_no=%s", result.ID, params.TradeNo))

	return result.URL, result.ID, nil
}

// VerifyCreemWebhook 验证 Creem webhook 签名
// Creem 使用 x-webhook-signature header (HMAC-SHA256 of payload)
func VerifyCreemWebhook(payload []byte, signature string) bool {
	ps := GetPaymentSetting()
	if ps.CreemWebhookSecret == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(ps.CreemWebhookSecret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expected))
}
