package ratio_setting

import "sync"

// ==================== 货币汇率管理 ====================
// 用于分层计费时将不同货币的 API 价格转换为统一额度单位

var (
	exchangeMu sync.RWMutex
	// 美元对各主要货币汇率（USD = 1.0 为基准）
	usdRates = map[string]float64{
		"USD": 1.0,
		"CNY": 7.24,
		"JPY": 149.5,
		"EUR": 0.92,
		"GBP": 0.79,
		"KRW": 1320.0,
		"AUD": 1.53,
		"CAD": 1.36,
		"SGD": 1.34,
		"HKD": 7.81,
		"TWD": 31.5,
		"INR": 83.2,
	}
)

// SetExchangeRate 设置美元对某货币的汇率
func SetExchangeRate(currency string, rate float64) {
	exchangeMu.Lock()
	defer exchangeMu.Unlock()
	usdRates[currency] = rate
}

// SetUSDCNY 设置美元兑人民币汇率
func SetUSDCNY(rate float64) { SetExchangeRate("CNY", rate) }

// SetUSDJPY 设置美元兑日元汇率
func SetUSDJPY(rate float64) { SetExchangeRate("JPY", rate) }

// SetUSDEUR 设置美元兑欧元汇率
func SetUSDEUR(rate float64) { SetExchangeRate("EUR", rate) }

// SetUSDGBP 设置美元兑英镑汇率
func SetUSDGBP(rate float64) { SetExchangeRate("GBP", rate) }

// GetExchangeRate 获取美元对某货币的汇率
func GetExchangeRate(currency string) float64 {
	exchangeMu.RLock()
	defer exchangeMu.RUnlock()
	if rate, ok := usdRates[currency]; ok {
		return rate
	}
	return 1.0
}

// ConvertToQuotaUSD 将任意货币金额转换为美元（用于统一计费）
func ConvertToQuotaUSD(amount float64, currency string) float64 {
	rate := GetExchangeRate(currency)
	if rate <= 0 {
		return amount
	}
	return amount / rate
}

// GetExchangeRatesCopy 获取所有汇率副本
func GetExchangeRatesCopy() map[string]float64 {
	exchangeMu.RLock()
	defer exchangeMu.RUnlock()
	result := make(map[string]float64, len(usdRates))
	for k, v := range usdRates {
		result[k] = v
	}
	return result
}

// UpdateExchangeRatesFromMap 批量更新汇率
func UpdateExchangeRatesFromMap(rates map[string]float64) {
	exchangeMu.Lock()
	defer exchangeMu.Unlock()
	for currency, rate := range rates {
		if rate > 0 {
			usdRates[currency] = rate
		}
	}
}
