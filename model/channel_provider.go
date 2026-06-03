package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/relay/channeltype"
)

// ChannelProvider 渠道提供商表 — 统一管理各品牌的基础信息
// 替代散落在 Go 源码中的 ChannelBaseURLs / ChannelTypeNames / DefaultChannelModels / brandConfigs
// 数据在启动时从已有硬编码 seed，之后可通过定时同步服务自动更新
type ChannelProvider struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	TypeID       int    `gorm:"type:int;not null;uniqueIndex" json:"type_id"`
	Name         string `gorm:"type:varchar(100);not null" json:"name"`
	BaseURL      string `gorm:"type:varchar(500)" json:"base_url"`
	Region       string `gorm:"type:varchar(20);default:'overseas'" json:"region"`
	Models       string `gorm:"type:text" json:"models"`     // JSON 数组，如 ["gpt-4o","gpt-4o-mini"]
	ProviderSlug string `gorm:"type:varchar(100)" json:"provider_slug"` // 对应 model_metadata.provider
	IsQuantum    bool   `gorm:"default:false" json:"is_quantum"`
	AutoSync     bool   `gorm:"default:true" json:"auto_sync"`  // 是否定时从 /v1/models 拉取最新模型
	LastSynced   int64  `json:"last_synced"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
}

// GetModels 返回解析后的模型列表
func (p *ChannelProvider) GetModels() []string {
	if p.Models == "" {
		return nil
	}
	var list []string
	if err := json.Unmarshal([]byte(p.Models), &list); err != nil {
		return nil
	}
	return list
}

// SeedChannelProviders 启动时从已有的硬编码数据 seed channel_providers 表
// 只插入不存在的记录，已有记录不覆盖（保留管理员的后台修改）
// 但如果某个 type_id 在老版本中新增了字段（如 URL），会更新空字段
func SeedChannelProviders() {
	now := time.Now().Unix()

	for typeID, name := range channeltype.ChannelTypeNames {
		if typeID <= 0 {
			continue
		}

		// 跳过 Sub2API / VLLM / SGLang（type >= 200，非标准渠道提供商）
		if typeID >= 200 {
			continue
		}

		baseURL := ""
		if typeID >= 0 && typeID < len(channeltype.ChannelBaseURLs) {
			baseURL = channeltype.ChannelBaseURLs[typeID]
		}

		region := channeltype.RegionOverseas
		if r, ok := channeltype.ChannelTypeRegion[typeID]; ok {
			region = r
		}

		providerSlug := ""
		if mapped, ok := channeltype.ChannelTypeNameToProvider[name]; ok {
			providerSlug = mapped
		}

		isQuantum := typeID >= 100

		// 默认模型列表
		var models []string
		if dm, ok := DefaultChannelModels[typeID]; ok && dm != "" {
			models = strings.Split(dm, ",")
		}
		modelsJSON := "[]"
		if len(models) > 0 {
			if b, err := json.Marshal(models); err == nil {
				modelsJSON = string(b)
			}
		}

		// 检查是否已存在
		var existing ChannelProvider
		result := DB.Where("type_id = ?", typeID).First(&existing)

		if result.Error != nil {
			// 不存在 → 插入
			provider := ChannelProvider{
				TypeID:       typeID,
				Name:         name,
				BaseURL:      baseURL,
				Region:       region,
				Models:       modelsJSON,
				ProviderSlug: providerSlug,
				IsQuantum:    isQuantum,
				AutoSync:     baseURL != "" && !isQuantum, // 量子渠道暂不自动同步
				LastSynced:   0,
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			if err := DB.Create(&provider).Error; err != nil {
				logger.SysError(fmt.Sprintf("seed channel_provider type_id=%d: %v", typeID, err))
			}
		} else {
			// 已存在 → 只更新空字段（URL、ProviderSlug 等）
			updates := make(map[string]interface{})
			if existing.BaseURL == "" && baseURL != "" {
				updates["base_url"] = baseURL
			}
			if existing.Region == "" && region != "" {
				updates["region"] = region
			}
			if existing.ProviderSlug == "" && providerSlug != "" {
				updates["provider_slug"] = providerSlug
			}
			if existing.Models == "[]" || existing.Models == "" {
				if modelsJSON != "[]" {
					updates["models"] = modelsJSON
				}
			}
			if len(updates) > 0 {
				updates["updated_at"] = now
				DB.Model(&existing).Updates(updates)
			}
		}
	}
	logger.SysLog(fmt.Sprintf("channel_providers seeded: %d types processed", len(channeltype.ChannelTypeNames)))
}
