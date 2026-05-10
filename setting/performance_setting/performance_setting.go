package performance_setting

import (
	"encoding/json"
	"sync"

	"github.com/quantumclaw/quantumclaw/setting/system_setting"
)

// Re-export PerformanceSetting from system_setting
type PerformanceSetting = system_setting.PerformanceSetting

var (
	perfSetting   *PerformanceSetting
	perfSettingMu sync.RWMutex
)

// GetPerformanceSetting 获取性能监控设置（单例）
func GetPerformanceSetting() *PerformanceSetting {
	perfSettingMu.RLock()
	defer perfSettingMu.RUnlock()
	if perfSetting == nil {
		perfSetting = &PerformanceSetting{
			EnablePrometheusMetrics: false,
			PrometheusPath:          "/metrics",
			EnableRuntimeLogs:       true,
			GCLogThresholdMB:        50,
		}
	}
	return perfSetting
}

// SetPerformanceSetting 更新性能监控设置
func SetPerformanceSetting(s *PerformanceSetting) {
	perfSettingMu.Lock()
	defer perfSettingMu.Unlock()
	if s == nil {
		s = &PerformanceSetting{}
	}
	perfSetting = s
}

// ParsePerformanceSetting 从 JSON 解析性能监控设置
func ParsePerformanceSetting(data string) (*PerformanceSetting, error) {
	var s PerformanceSetting
	if data == "" {
		return &PerformanceSetting{
			EnablePrometheusMetrics: false,
			PrometheusPath:          "/metrics",
			EnableRuntimeLogs:       true,
			GCLogThresholdMB:        50,
		}, nil
	}
	if err := json.Unmarshal([]byte(data), &s); err != nil {
		return nil, err
	}
	return &s, nil
}
