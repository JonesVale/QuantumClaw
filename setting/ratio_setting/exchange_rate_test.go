package ratio_setting

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetExchangeRate_KnownCurrencies(t *testing.T) {
	tests := []struct {
		currency string
		wantRate float64
	}{
		{"USD", 1.0},
		{"CNY", 7.24},
		{"JPY", 149.5},
		{"EUR", 0.92},
		{"GBP", 0.79},
		{"KRW", 1320.0},
		{"AUD", 1.53},
		{"CAD", 1.36},
		{"SGD", 1.34},
		{"HKD", 7.81},
		{"TWD", 31.5},
		{"INR", 83.2},
	}

	for _, tt := range tests {
		t.Run(tt.currency, func(t *testing.T) {
			got := GetExchangeRate(tt.currency)
			assert.Equal(t, tt.wantRate, got, "USD 对 %s 的汇率应为 %f", tt.currency, tt.wantRate)
		})
	}
}

func TestGetExchangeRate_UnknownCurrency(t *testing.T) {
	// 未知货币默认返回 1.0（视为与 USD 等价）
	got := GetExchangeRate("UNKNOWN")
	assert.Equal(t, float64(1.0), got, "未知货币应默认返回 1.0")
}

func TestSetExchangeRate(t *testing.T) {
	// 设置新的汇率
	SetExchangeRate("BTC", 60000.0)
	got := GetExchangeRate("BTC")
	assert.Equal(t, float64(60000.0), got)

	// 覆盖已有汇率
	SetExchangeRate("CNY", 7.5)
	got = GetExchangeRate("CNY")
	assert.Equal(t, float64(7.5), got, "汇率应被覆盖")
}

func TestSetUSDCNY(t *testing.T) {
	SetUSDCNY(7.8)
	got := GetExchangeRate("CNY")
	assert.Equal(t, float64(7.8), got)
}

func TestSetUSDJPY(t *testing.T) {
	SetUSDJPY(150.0)
	got := GetExchangeRate("JPY")
	assert.Equal(t, float64(150.0), got)
}

func TestConvertToQuotaUSD(t *testing.T) {
	// 将 100 CNY 转换为 USD：100 / 7.24 = ~13.81 USD
	quota := ConvertToQuotaUSD(100.0, "CNY")
	assert.InDelta(t, 13.812, quota, 0.01, "100 CNY 应约为 13.81 USD")

	// USD 汇率是 1.0，转换后不变
	quota = ConvertToQuotaUSD(50.0, "USD")
	assert.Equal(t, float64(50.0), quota)

	// JPY 转换
	quota = ConvertToQuotaUSD(14950.0, "JPY")
	assert.InDelta(t, 100.0, quota, 0.1, "14950 JPY 应约为 100 USD")
}

func TestConvertToQuotaUSD_ZeroRate(t *testing.T) {
	// 零汇率应返回原值
	quota := ConvertToQuotaUSD(100.0, "ZERO")
	assert.Equal(t, float64(100.0), quota)
}

func TestConvertToQuotaUSD_NegativeAmount(t *testing.T) {
	// 负数金额应正常处理
	quota := ConvertToQuotaUSD(-50.0, "CNY")
	assert.InDelta(t, -6.91, quota, 0.01)
}

func TestGetExchangeRatesCopy(t *testing.T) {
	rates := GetExchangeRatesCopy()

	assert.Greater(t, len(rates), 0, "应有至少一个汇率")
	assert.Equal(t, float64(1.0), rates["USD"], "USD 应为 1.0")
	assert.Equal(t, float64(7.24), rates["CNY"], "CNY 应为 7.24")

	// 验证返回的是副本，修改副本不影响原数据
	rates["USD"] = 999.0
	original := GetExchangeRate("USD")
	assert.Equal(t, float64(1.0), original, "副本修改不应影响原数据")
}

func TestUpdateExchangeRatesFromMap(t *testing.T) {
	UpdateExchangeRatesFromMap(map[string]float64{
		"CNY": 7.5,
		"EUR": 0.95,
		"NEW_CURRENCY": 123.45,
	})

	assert.Equal(t, float64(7.5), GetExchangeRate("CNY"))
	assert.Equal(t, float64(0.95), GetExchangeRate("EUR"))
	assert.Equal(t, float64(123.45), GetExchangeRate("NEW_CURRENCY"))

	// 已有汇率保持不变
	assert.Equal(t, float64(149.5), GetExchangeRate("JPY"))
}

func TestUpdateExchangeRatesFromMap_ZeroRate(t *testing.T) {
	// 零或负数汇率不应更新
	initialCNY := GetExchangeRate("CNY")
	UpdateExchangeRatesFromMap(map[string]float64{
		"CNY": 0.0,
		"EUR": -1.0,
	})

	assert.Equal(t, initialCNY, GetExchangeRate("CNY"), "零汇率不应更新")
}

func TestConvertToQuotaUSD_RealWorldScenario(t *testing.T) {
	// 模拟：DeepSeek API 收费 0.1 USD/M token
	// 用户使用 CNY 充值，系统按汇率换算
	deepseekInputCost := 0.1       // USD per 1M tokens
	userCurrency := "CNY"
	userCurrencyRate := GetExchangeRate(userCurrency)

	costInUserCurrency := deepseekInputCost * userCurrencyRate
	assert.InDelta(t, 0.724, costInUserCurrency, 0.01, "0.1 USD 应约为 0.724 CNY")

	// 验证反向计算
	usdBack := costInUserCurrency / userCurrencyRate
	assert.InDelta(t, 0.1, usdBack, 0.001, "反向计算应回到原值")
}
