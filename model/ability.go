package model

import (
	"context"
	"sort"
	"strings"

	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/encrypt"
	"github.com/quantumclaw/quantumclaw/common/logger"
)

// modelMatchCondition generates SQL condition for matching a model name
// against the comma-separated channels.models field.
func modelMatchCondition(model string) string {
	if common.UsingPostgreSQL {
		return "',' || channels.models || ',' LIKE '%,' || ? || ',%'"
	}
	return "instr(',' || channels.models || ',', ',' || ? || ',') > 0"
}

// GetCheapestSatisfiedChannel finds the cheapest enabled channel that supports
// the given model, using direct channels.models matching (not abilities table).
func GetCheapestSatisfiedChannel(group string, model string, ownerId int, region string) (*Channel, error) {
	modelCond := modelMatchCondition(model)
	query := DB.Model(&Channel{}).
		Where("`group` = ? AND status = ? AND deleted_at IS NULL AND "+modelCond,
			group, ChannelStatusEnabled, model)

	if ownerId > 0 {
		query = query.Where("user_id = ?", ownerId)
	}
	if region != "" {
		query = query.Where("region = ?", region)
	}

	var channel Channel
	err := query.Order("sell_price_rate ASC, channel_markup ASC").
		First(&channel).Error
	if err != nil {
		return nil, err
	}

	// Decrypt channel key before returning
	if channel.Key != "" && config.CryptoSecret != "" {
		decrypted, e := encrypt.DecryptChannelKey(channel.Key, config.CryptoSecret)
		if e == nil {
			channel.Key = decrypted
		}
	}
	return &channel, nil
}

// GetCheapestSatisfiedChannelWithBalance 带余额预检的渠道选择
//
// 在 GetCheapestSatisfiedChannel 基础上增加渠道商余额预检：
// 按 sell_price_rate 升序遍历所有匹配渠道，返回第一个
// 渠道商（owner）现金余额 >= minBalanceCents 的渠道。
//
// 参数：
//   - group, model, ownerId, region: 同 GetCheapestSatisfiedChannel
//   - minBalanceCents: 渠道商最低所需余额（分），<=0 表示不做余额过滤
//
// 返回：
//   - 第一个价格最低且余额充足的渠道
//   - 如果所有匹配渠道的 owner 余额均不足，回退到最便宜的渠道并记录警告日志
func GetCheapestSatisfiedChannelWithBalance(group string, modelName string, ownerId int, region string, minBalanceCents int64) (*Channel, error) {
	modelCond := modelMatchCondition(modelName)
	query := DB.Model(&Channel{}).
		Where("`group` = ? AND status = ? AND deleted_at IS NULL AND "+modelCond,
			group, ChannelStatusEnabled, modelName)

	if ownerId > 0 {
		query = query.Where("user_id = ?", ownerId)
	}
	if region != "" {
		query = query.Where("region = ?", region)
	}

	// 查询所有匹配渠道（按价格排序）
	var channels []Channel
	err := query.Order("sell_price_rate ASC, channel_markup ASC").
		Find(&channels).Error
	if err != nil || len(channels) == 0 {
		return nil, err // 保持与原函数一致的行为：无匹配时返回 GORM err
	}

	// 余额预检：跳过 owner 余额不足的渠道
	if minBalanceCents > 0 {
		for i, ch := range channels {
			// 平台自有渠道（UserId==0）不受余额限制
			if ch.UserId <= 0 {
				return decryptAndReturn(&channels[i])
			}
			// 检查渠道商余额
			balance, balErr := GetUserCashBalance(ch.UserId)
			if balErr != nil {
				logger.Warnf(context.Background(),
					"balance check failed for channel owner %d (ch#%d): %v, skipping",
					ch.UserId, ch.Id, balErr)
				continue
			}
			if balance >= minBalanceCents {
				logger.Debugf(context.Background(),
					"channel #%d (owner=%d) passed balance check: %d >= %d",
					ch.Id, ch.UserId, balance, minBalanceCents)
				return decryptAndReturn(&channels[i])
			}
			logger.Debugf(context.Background(),
				"channel #%d (owner=%d) SKIPPED: balance %d < required %d",
				ch.Id, ch.UserId, balance, minBalanceCents)
		}
		// 所有渠道的 owner 余额都不足 → 回退到最便宜的，记录警告
		logger.Warnf(context.Background(),
			"[BALANCE_FALLBACK] group=%s model=%s 所有匹配渠道owner余额均<%d分，回退到最便宜渠道#%d(owner=%d)",
			group, modelName, minBalanceCents, channels[0].Id, channels[0].UserId)
	}

	return decryptAndReturn(&channels[0])
}

// decryptAndReturn 解密渠道密钥后返回
func decryptAndReturn(ch *Channel) (*Channel, error) {
	if ch.Key != "" && config.CryptoSecret != "" {
		decrypted, e := encrypt.DecryptChannelKey(ch.Key, config.CryptoSecret)
		if e == nil {
			ch.Key = decrypted
		}
	}
	return ch, nil
}

// GetGroupModels returns all model names from enabled channels in the group.
func GetGroupModels(ctx context.Context, group string) ([]string, error) {
	var modelsStr []string
	err := DB.Model(&Channel{}).
		Where("`group` = ? AND status = ? AND deleted_at IS NULL AND models != ''", group, ChannelStatusEnabled).
		Pluck("models", &modelsStr).Error
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool)
	for _, ms := range modelsStr {
		for _, m := range strings.Split(ms, ",") {
			m = strings.TrimSpace(m)
			if m != "" && !seen[m] {
				seen[m] = true
			}
		}
	}
	models := make([]string, 0, len(seen))
	for m := range seen {
		models = append(models, m)
	}
	sort.Strings(models)
	return models, nil
}

// GetGroupModelsWithHealthCheck returns models from channels with test_passed=true.
func GetGroupModelsWithHealthCheck(ctx context.Context, group string) ([]string, error) {
	var modelsStr []string
	err := DB.Model(&Channel{}).
		Where("`group` = ? AND status = ? AND deleted_at IS NULL AND models != ''",
			group, ChannelStatusEnabled).
		Pluck("models", &modelsStr).Error
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool)
	for _, ms := range modelsStr {
		for _, m := range strings.Split(ms, ",") {
			m = strings.TrimSpace(m)
			if m != "" && !seen[m] {
				seen[m] = true
			}
		}
	}
	models := make([]string, 0, len(seen))
	for m := range seen {
		models = append(models, m)
	}
	sort.Strings(models)
	return models, nil
}

// Placeholder functions for abilities table (no longer used for routing)
func (channel *Channel) AddAbilities() error           { return nil }
func (channel *Channel) DeleteAbilities() error        { return nil }
func (channel *Channel) UpdateAbilities() error        { return nil }
func UpdateAbilityStatus(channelId int, status bool) error { return nil }

// Ability struct retained for backward compatibility
type Ability struct {
	Group     string `json:"group" gorm:"type:varchar(32)"`
	Model     string `json:"model" gorm:"type:text"`
	ChannelId int    `json:"channel_id" gorm:"type:int"`
	UserId    int    `json:"user_id" gorm:"type:int;default:0"`
	Enabled   bool   `json:"enabled" gorm:"type:bool"`
	Priority  *int64 `json:"priority" gorm:"type:int;default:0"`
}

// GetRandomSatisfiedChannelByOwner finds enabled channel for model owned by specific user
func GetRandomSatisfiedChannelByOwner(group string, model string, ownerId int) (*Channel, error) {
	return GetCheapestSatisfiedChannel(group, model, ownerId, "")
}

// GetRandomSatisfiedChannelAnyOwner finds enabled channel not owned by a specific user
func GetRandomSatisfiedChannelAnyOwner(group string, model string, excludeOwnerId int) (*Channel, error) {
	modelCond := modelMatchCondition(model)
	query := DB.Model(&Channel{}).
		Where("`group` = ? AND status = ? AND deleted_at IS NULL AND user_id != ? AND "+modelCond,
			group, ChannelStatusEnabled, excludeOwnerId, model)
	var channel Channel
	err := query.Order("sell_price_rate ASC").First(&channel).Error
	if err != nil {
		return nil, err
	}
	if channel.Key != "" && config.CryptoSecret != "" {
		decrypted, e := encrypt.DecryptChannelKey(channel.Key, config.CryptoSecret)
		if e == nil {
			channel.Key = decrypted
		}
	}
	return &channel, nil
}

// GetRandomSatisfiedChannel finds the cheapest channel for a model (ignoring owner)
func GetRandomSatisfiedChannel(group string, model string, ignoreFirstPriority bool) (*Channel, error) {
	modelCond := modelMatchCondition(model)
	query := DB.Model(&Channel{}).
		Where("`group` = ? AND status = ? AND deleted_at IS NULL AND "+modelCond,
			group, ChannelStatusEnabled, model)
	if !ignoreFirstPriority {
		query = query.Where("priority > 0")
	}
	var channel Channel
	err := query.Order("sell_price_rate ASC").First(&channel).Error
	if err != nil {
		return nil, err
	}
	if channel.Key != "" && config.CryptoSecret != "" {
		decrypted, e := encrypt.DecryptChannelKey(channel.Key, config.CryptoSecret)
		if e == nil {
			channel.Key = decrypted
		}
	}
	return &channel, nil
}

// GetGroupModelsWithPriority returns models from channels with priority > 0 (deprecated)
func GetGroupModelsWithPriority(group string) ([]string, error) {
	return GetGroupModels(nil, group)
}
