package common

import (
	"sync"
)

// ==================== 公司收款信息 ====================

var (
	companyInfoOnce sync.Once
	companyInfo     *CompanyPaymentInfo
)

// CompanyPaymentInfo 公司收款信息（对公转账）
type CompanyPaymentInfo struct {
	CompanyName    string `json:"company_name"`
	TaxID          string `json:"tax_id"`
	Address        string `json:"address"`
	Phone          string `json:"phone"`
	BankName       string `json:"bank_name"`
	BankAccount    string `json:"bank_account"`
	AlipayQRURL    string `json:"alipay_qr_url"`
}

// GetCompanyPaymentInfo 获取公司收款信息（单例）
func GetCompanyPaymentInfo() *CompanyPaymentInfo {
	companyInfoOnce.Do(func() {
		companyInfo = &CompanyPaymentInfo{
			CompanyName: getEnvOrDefault("COMPANY_NAME", "深圳市中科劲纬智能有限公司"),
			TaxID:       getEnvOrDefault("COMPANY_TAX_ID", "91440300MA5GH45W8C"),
			Address:     getEnvOrDefault("COMPANY_ADDRESS", "深圳市宝安区石岩街道塘头社区塘头大道33号东海创意园205A"),
			Phone:       getEnvOrDefault("COMPANY_PHONE", "15920005303"),
			BankName:    getEnvOrDefault("COMPANY_BANK", "深圳农村商业银行股份有限公司应人石支行"),
			BankAccount: getEnvOrDefault("COMPANY_BANK_ACCOUNT", "000396168236"),
			AlipayQRURL: getEnvOrDefault("COMPANY_ALIPAY_QR", "/payment/alipay-qr.jpg"),
		}
	})
	return companyInfo
}

// IsCompanyPaymentEnabled 检查公司收款信息是否完整
func IsCompanyPaymentEnabled() bool {
	info := GetCompanyPaymentInfo()
	return info.CompanyName != "" && info.BankAccount != ""
}
