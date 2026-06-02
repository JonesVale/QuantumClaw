package operation_setting

import "sync"

// CheckinSetting 签到功能配置
type CheckinSetting struct {
	Enabled  bool `json:"enabled"`   // 是否启用签到功能
	MinQuota int  `json:"min_quota"` // 签到最小额度奖励
	MaxQuota int  `json:"max_quota"` // 签到最大额度奖励
}

var (
	checkinSettingMu sync.RWMutex
	checkinSetting   = CheckinSetting{
		Enabled:  true,
		MinQuota: 10000,  // 约 0.02 USD
		MaxQuota: 50000, // 约 0.1 USD
	}
)

// GetCheckinSetting 获取签到配置
func GetCheckinSetting() *CheckinSetting {
	checkinSettingMu.RLock()
	defer checkinSettingMu.RUnlock()
	copy := checkinSetting
	return &copy
}

// SetCheckinSetting 更新签到配置
func SetCheckinSetting(s CheckinSetting) {
	checkinSettingMu.Lock()
	defer checkinSettingMu.Unlock()
	checkinSetting = s
}

// IsCheckinEnabled 是否启用签到功能
func IsCheckinEnabled() bool {
	checkinSettingMu.RLock()
	defer checkinSettingMu.RUnlock()
	return checkinSetting.Enabled
}

// GetCheckinQuotaRange 获取签到额度范围
func GetCheckinQuotaRange() (min, max int) {
	checkinSettingMu.RLock()
	defer checkinSettingMu.RUnlock()
	return checkinSetting.MinQuota, checkinSetting.MaxQuota
}
