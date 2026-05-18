package common

import (
	"fmt"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/webhook"
)

// ==================== Stripe 支付集成 ====================

// PaymentCheckoutParams 通用的支付 checkout 参数
type PaymentCheckoutParams struct {
	TradeNo       string
	Amount        int64   // 配额数量
	PayMoney      float64 // 实际支付金额（元）
	UserEmail     string
	SuccessURL    string
	CancelURL     string
	NotifyURL     string
	ProductName   string
}

// CreateStripeCheckoutSession 创建 Stripe Checkout Session
// 返回 checkout URL 和 session ID
func CreateStripeCheckoutSession(params *PaymentCheckoutParams) (checkoutURL string, sessionID string, err error) {
	ps := GetPaymentSetting()
	if ps.StripeApiSecret == "" {
		return "", "", fmt.Errorf("Stripe API Secret 未配置")
	}

	stripe.Key = ps.StripeApiSecret

	// 金额单位转换：元 -> 分（Stripe 使用最小货币单位）
	unitAmount := int64(params.PayMoney * 100)

	domain := ps.PaymentReturnURL
	if domain == "" {
		domain = params.SuccessURL
	}
	if domain == "" {
		domain = "http://localhost:3666"
	}

	sessionParams := &stripe.CheckoutSessionParams{
		Params: stripe.Params{
			Metadata: map[string]string{
				"trade_no": params.TradeNo,
			},
		},
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("cny"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(params.ProductName),
					},
					UnitAmount: stripe.Int64(unitAmount),
				},
				Quantity: stripe.Int64(params.Amount),
			},
		},
		ClientReferenceID: stripe.String(params.TradeNo),
		CustomerEmail:     stripe.String(params.UserEmail),
		SuccessURL:        stripe.String(params.SuccessURL),
		CancelURL:         stripe.String(params.CancelURL),
	}

	if params.NotifyURL != "" {
		if sessionParams.Params.Metadata == nil {
			sessionParams.Params.Metadata = make(map[string]string)
		}
		sessionParams.Params.Metadata["notify_url"] = params.NotifyURL
	}

	s, err := session.New(sessionParams)
	if err != nil {
		return "", "", fmt.Errorf("Stripe checkout session 创建失败: %w", err)
	}

	return s.URL, s.ID, nil
}

// VerifyStripeWebhook 验证 Stripe Webhook 签名
// 返回解析后的事件类型和 trade_no
func VerifyStripeWebhook(payload []byte, sigHeader string, webhookSecret string) (eventType string, tradeNo string, err error) {
	if webhookSecret == "" {
		return "", "", fmt.Errorf("Stripe webhook secret 未配置")
	}

	event, err := webhook.ConstructEvent(payload, sigHeader, webhookSecret)
	if err != nil {
		return "", "", fmt.Errorf("Stripe webhook 验签失败: %w", err)
	}

	tradeNo = ""
	if event.Data != nil && event.Data.Object != nil {
		if refID, ok := event.Data.Object["client_reference_id"]; ok {
			if s, ok := refID.(string); ok {
				tradeNo = s
			}
		}
	}

	return string(event.Type), tradeNo, nil
}

// GetStripeEventType 从事件中提取 type（别名）
func GetStripeEventType(payload []byte, sigHeader string, webhookSecret string) (string, string, error) {
	return VerifyStripeWebhook(payload, sigHeader, webhookSecret)
}
