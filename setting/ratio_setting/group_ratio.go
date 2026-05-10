package ratio_setting

import (
	"encoding/json"
	"errors"
	"sync"
)

// 分组比率管理（适配 QuantumClaw，不依赖 types.RWMap）
var (
	groupRatioMu sync.RWMutex
	groupRatio   = map[string]float64{
		"default": 1,
		"vip":     1,
		"svip":    1,
	}

	// 充值分组比率（不同分组的充值额度倍率）
	topupGroupRatioMu sync.RWMutex
	topupGroupRatio   = map[string]float64{
		"default": 1.0,
		"vip":     1.1,
		"svip":    1.2,
	}
)

// GetGroupRatioCopy 获取分组比率的只读副本
func GetGroupRatioCopy() map[string]float64 {
	groupRatioMu.RLock()
	defer groupRatioMu.RUnlock()
	result := make(map[string]float64, len(groupRatio))
	for k, v := range groupRatio {
		result[k] = v
	}
	return result
}

// ContainsGroupRatio 检查分组是否存在
func ContainsGroupRatio(name string) bool {
	groupRatioMu.RLock()
	defer groupRatioMu.RUnlock()
	_, ok := groupRatio[name]
	return ok
}

// GetGroupRatio 获取分组比率，不存在时返回1
func GetGroupRatio(name string) float64 {
	groupRatioMu.RLock()
	defer groupRatioMu.RUnlock()
	if ratio, ok := groupRatio[name]; ok {
		return ratio
	}
	return 1
}

// GroupRatio2JSONString 序列化分组比率为 JSON 字符串
func GroupRatio2JSONString() string {
	groupRatioMu.RLock()
	defer groupRatioMu.RUnlock()
	data, _ := json.Marshal(groupRatio)
	return string(data)
}

// UpdateGroupRatioByJSONString 从 JSON 字符串更新分组比率
func UpdateGroupRatioByJSONString(jsonStr string) error {
	newMap := make(map[string]float64)
	if err := json.Unmarshal([]byte(jsonStr), &newMap); err != nil {
		return err
	}
	for name, ratio := range newMap {
		if ratio < 0 {
			return errors.New("group ratio must be >= 0: " + name)
		}
	}
	groupRatioMu.Lock()
	defer groupRatioMu.Unlock()
	for k, v := range newMap {
		groupRatio[k] = v
	}
	return nil
}

// GetTopupGroupRatio 获取充值分组倍率（不同分组购买额度的倍数）
func GetTopupGroupRatio(group string) float64 {
	topupGroupRatioMu.RLock()
	defer topupGroupRatioMu.RUnlock()
	if ratio, ok := topupGroupRatio[group]; ok {
		return ratio
	}
	return 1.0
}

// SetTopupGroupRatio 设置充值分组倍率
func SetTopupGroupRatio(group string, ratio float64) {
	topupGroupRatioMu.Lock()
	defer topupGroupRatioMu.Unlock()
	topupGroupRatio[group] = ratio
}

// GetTopupGroupRatioCopy 获取充值分组比率副本
func GetTopupGroupRatioCopy() map[string]float64 {
	topupGroupRatioMu.RLock()
	defer topupGroupRatioMu.RUnlock()
	result := make(map[string]float64, len(topupGroupRatio))
	for k, v := range topupGroupRatio {
		result[k] = v
	}
	return result
}

// TopupGroupRatio2JSONString 序列化充值分组比率
func TopupGroupRatio2JSONString() string {
	topupGroupRatioMu.RLock()
	defer topupGroupRatioMu.RUnlock()
	data, _ := json.Marshal(topupGroupRatio)
	return string(data)
}

// UpdateTopupGroupRatioByJSONString 从 JSON 字符串更新充值分组比率
func UpdateTopupGroupRatioByJSONString(jsonStr string) error {
	newMap := make(map[string]float64)
	if err := json.Unmarshal([]byte(jsonStr), &newMap); err != nil {
		return err
	}
	topupGroupRatioMu.Lock()
	defer topupGroupRatioMu.Unlock()
	for k, v := range newMap {
		topupGroupRatio[k] = v
	}
	return nil
}
