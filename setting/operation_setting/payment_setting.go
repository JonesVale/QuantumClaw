package operation_setting

import "sync"

// PayMethods 是可展示的支付方式列表（管理员可配置）
var (
	payMethodsMu sync.RWMutex
	PayMethods   = []map[string]string{}
)

func GetPayMethods() []map[string]string {
	payMethodsMu.RLock()
	defer payMethodsMu.RUnlock()
	result := make([]map[string]string, len(PayMethods))
	for i, m := range PayMethods {
		mc := make(map[string]string, len(m))
		for k, v := range m {
			mc[k] = v
		}
		result[i] = mc
	}
	return result
}

func SetPayMethods(methods []map[string]string) {
	payMethodsMu.Lock()
	defer payMethodsMu.Unlock()
	PayMethods = methods
}
