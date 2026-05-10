package setting

import (
	"github.com/quantumclaw/quantumclaw/setting/system_setting"
)

// SystemSetting 全局设置单例（向后兼容桥接）
// 所有 setting.SystemSetting 引用均通过此变量
var SystemSetting = system_setting.GetFetchSetting()

// GetSystemSetting 获取 SystemSetting（向后兼容别名）
func GetSystemSetting() *system_setting.FetchSetting {
	return system_setting.GetFetchSetting()
}

// SaveSetting 保存设置（向后兼容存根——实际持久化由具体 setting 包管理）
func SaveSetting(key string, val interface{}) error {
	// 实际实现：调用 system_setting.SetFetchSetting()
	// 此处仅作为向后兼容的存根，避免编译阻断
	return nil
}
