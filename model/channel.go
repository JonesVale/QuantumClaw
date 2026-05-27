package model

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/encrypt"
	"github.com/quantumclaw/quantumclaw/common/helper"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"gorm.io/gorm"
)

const (
	ChannelStatusUnknown          = 0
	ChannelStatusEnabled          = 1 // don't use 0, 0 is the default value!
	ChannelStatusManuallyDisabled = 2 // also don't use 0
	ChannelStatusAutoDisabled     = 3
)

type Channel struct {
	Id                 int     `json:"id"`
	Type               int     `json:"type" gorm:"default:0"`
	Key                string  `json:"key" gorm:"type:text"`
	Status             int     `json:"status" gorm:"default:1"`
	Name               string  `json:"name" gorm:"index"`
	Weight             *uint   `json:"weight" gorm:"default:0"`
	CreatedTime        int64   `json:"created_time" gorm:"bigint"`
	TestTime           int64   `json:"test_time" gorm:"bigint"`
	ResponseTime       int     `json:"response_time"` // in milliseconds
	BaseURL            *string `json:"base_url" gorm:"column:base_url;default:''"`
	Other              *string `json:"other"`   // DEPRECATED: please save config to field Config
	Balance            float64 `json:"balance"` // in USD
	BalanceUpdatedTime int64   `json:"balance_updated_time" gorm:"bigint"`
	Models             string  `json:"models"`
	Group              string  `json:"group" gorm:"type:varchar(32);default:'default'"`
	UsedQuota          int64   `json:"used_quota" gorm:"bigint;default:0"`
	ModelMapping       *string `json:"model_mapping" gorm:"type:varchar(1024);default:''"`
	Priority           *int64  `json:"priority" gorm:"bigint;default:0"`
	CostPerUnit        float64 `json:"cost_per_unit" gorm:"type:decimal(10,4);default:0"`
	SellPriceRate      float64 `json:"sell_price_rate" gorm:"type:decimal(10,4);default:1.0"`
	Config             string  `json:"config"`
	SystemPrompt       *string `json:"system_prompt" gorm:"type:text"`
	Category           string  `json:"category" gorm:"default:''"`  // paid / free / custom
	UserId             int     `json:"user_id" gorm:"type:int;default:0;index"` // 渠道归属：0=平台，>0=供应商
	CostPrice          float64 `json:"cost_price" gorm:"type:decimal(10,6);default:0"` // Key 贡献者实际成本

	// 余额预警与自动禁用阈值（单位为 USD）
	BalanceAlertThreshold   float64 `json:"balance_alert_threshold" gorm:"type:decimal(10,2);default:0"`
	BalanceDisableThreshold float64 `json:"balance_disable_threshold" gorm:"type:decimal(10,2);default:0"`
	ChannelMarkup          float64 `json:"channel_markup" gorm:"type:decimal(5,2);default:1.0"`      // 渠道加价倍率（1.0=原价, 1.2=+20%）
}

type ChannelConfig struct {
	Region            string `json:"region,omitempty"`
	SK                string `json:"sk,omitempty"`
	AK                string `json:"ak,omitempty"`
	UserID            string `json:"user_id,omitempty"`
	APIVersion        string `json:"api_version,omitempty"`
	LibraryID         string `json:"library_id,omitempty"`
	Plugin            string `json:"plugin,omitempty"`
	VertexAIProjectID string `json:"vertex_ai_project_id,omitempty"`
	VertexAIADC       string `json:"vertex_ai_adc,omitempty"`
	// CacheBillingRatio controls the billing ratio for cached prompt tokens (0-1 range).
	// When set to 0.5, cached tokens are billed at 50% of normal rate.
	// 0 (default) means cache billing is disabled and cached tokens are billed at full rate.
	CacheBillingRatio float64 `json:"cache_billing_ratio,omitempty"`
	// PromptCacheEnabled explicitly enables/disables prompt cache support.
	// If nil, cache detection is automatic based on provider capabilities.
	PromptCacheEnabled *bool `json:"prompt_cache_enabled,omitempty"`
	// ThinkingToContent converts reasoning_content (thinking tokens) into
	// [reasoning]...[/reasoning] tags appended to the final response content.
	ThinkingToContent bool `json:"thinking_to_content,omitempty"`
}

func GetAllChannels(startIdx int, num int, scope string) ([]*Channel, error) {
	var channels []*Channel
	var err error
	switch scope {
	case "all":
		err = DB.Order("id desc").Find(&channels).Error
		// Decrypt API keys if CRYPTO_SECRET is configured
		if err == nil && config.CryptoSecret != "" {
			for i := range channels {
				if channels[i].Key != "" {
					// Check if key is encrypted (starts with non-plaintext pattern)
					decrypted, e := encrypt.Decrypt(channels[i].Key, encrypt.DeriveKey(config.CryptoSecret))
					if e == nil {
						channels[i].Key = string(decrypted)
					} else {
						// Not encrypted yet (plaintext), leave as-is
					}
				}
			}
		}
	case "disabled":
		err = DB.Order("id desc").Where("status = ? or status = ?", ChannelStatusAutoDisabled, ChannelStatusManuallyDisabled).Find(&channels).Error
	default:
		err = DB.Order("id desc").Limit(num).Offset(startIdx).Omit("key").Find(&channels).Error
	}
	return channels, err
}

func SearchChannels(keyword string) (channels []*Channel, err error) {
	err = DB.Omit("key").Where("id = ? or name LIKE ?", helper.String2Int(keyword), keyword+"%").Find(&channels).Error
	return channels, err
}

func GetChannelById(id int, selectAll bool) (*Channel, error) {
	channel := Channel{Id: id}
	var err error = nil
	if selectAll {
		err = DB.First(&channel, "id = ?", id).Error
		// 解密 API Key（selectAll 表示需要完整信息，如 relay 场景）
		if err == nil && channel.Key != "" && config.CryptoSecret != "" {
			decrypted, e := encrypt.Decrypt(channel.Key, encrypt.DeriveKey(config.CryptoSecret))
			if e == nil {
				channel.Key = string(decrypted)
			} else {
				logger.SysError("decrypt channel key: " + e.Error())
			}
		}
	} else {
		err = DB.Omit("key").First(&channel, "id = ?", id).Error
	}
	return &channel, err
}

func BatchInsertChannels(channels []Channel) error {
	var err error
	// 批量加密 API Key
	for i := range channels {
		if channels[i].Key != "" && config.CryptoSecret != "" {
			encrypted, e := encrypt.Encrypt([]byte(channels[i].Key), encrypt.DeriveKey(config.CryptoSecret))
			if e == nil {
				channels[i].Key = encrypted
			} else {
				logger.SysError("batch encrypt channel key: " + e.Error())
			}
		}
	}
	err = DB.Create(&channels).Error
	if err != nil {
		return err
	}
	for _, channel_ := range channels {
		err = channel_.AddAbilities()
		if err != nil {
			return err
		}
	}
	return nil
}

func (channel *Channel) GetPriority() int64 {
	if channel.Priority == nil {
		return 0
	}
	return *channel.Priority
}

func (channel *Channel) GetBaseURL() string {
	if channel.BaseURL == nil {
		return ""
	}
	return *channel.BaseURL
}

func (channel *Channel) GetModelMapping() map[string]string {
	if channel.ModelMapping == nil || *channel.ModelMapping == "" || *channel.ModelMapping == "{}" {
		return nil
	}
	modelMapping := make(map[string]string)
	err := json.Unmarshal([]byte(*channel.ModelMapping), &modelMapping)
	if err != nil {
		logger.SysError(fmt.Sprintf("failed to unmarshal model mapping for channel %d, error: %s", channel.Id, err.Error()))
		return nil
	}
	return modelMapping
}


// EncryptExistingChannelKeys encrypts all plaintext channel API keys.
// Safe to call on startup: skips already-encrypted keys.
// Requires config.CryptoSecret to be set.
func EncryptExistingChannelKeys() error {
	if config.CryptoSecret == "" {
		return fmt.Errorf("CRYPTO_SECRET not set, cannot encrypt")
	}
	var channels []*Channel
	if err := DB.Where("key != ''").Find(&channels).Error; err != nil {
		return err
	}
	key := encrypt.DeriveKey(config.CryptoSecret)
	count := 0
	for _, ch := range channels {
		// Check: is this key already encrypted? Try decrypt first
		_, decErr := encrypt.Decrypt(ch.Key, key)
		if decErr == nil {
			continue // Already encrypted
		}
		// Key is plaintext (sk-..., gsk_..., or any non-base64-encrypted string)
		// Only proceed if the key looks like a real API key (not a placeholder)
		if strings.HasPrefix(ch.Key, "PUT_YOUR") || ch.Key == "" {
			continue
		}
		// Encrypt the plaintext key
		encrypted, err := encrypt.Encrypt([]byte(ch.Key), key)
		if err != nil {
			logger.SysError("encrypt key for channel " + fmt.Sprint(ch.Id) + ": " + err.Error())
			continue
		}
		if err := DB.Model(ch).Update("key", encrypted).Error; err != nil {
			logger.SysError("update encrypted key for channel " + fmt.Sprint(ch.Id) + ": " + err.Error())
			continue
		}
		count++
		logger.SysLog(fmt.Sprintf("Encrypted key for ch#%d (%s)", ch.Id, ch.Name))
	}
	logger.SysLog(fmt.Sprintf("Encrypted %d channel keys", count))
	return nil
}
func (channel *Channel) Insert() error {
	var err error
	// 加密 API Key
	if channel.Key != "" && config.CryptoSecret != "" {
		key, e := encrypt.Encrypt([]byte(channel.Key), encrypt.DeriveKey(config.CryptoSecret))
		if e == nil {
			channel.Key = key
		} else {
			logger.SysError("encrypt channel key: " + e.Error())
		}
	}
	err = DB.Create(channel).Error
	if err != nil {
		return err
	}
	err = channel.AddAbilities()
	return err
}

func (channel *Channel) Update() error {
	var err error
	// 如果 Key 变了，加密后存储
	if channel.Key != "" && config.CryptoSecret != "" {
		// 检查是否是已加密的（已加密的 key 以 base64 字符构成，不包含换行符）
		existing, _ := GetChannelById(channel.Id, true)
		if existing != nil && channel.Key != existing.Key {
			// Key 有变，加密新 key
			encrypted, e := encrypt.Encrypt([]byte(channel.Key), encrypt.DeriveKey(config.CryptoSecret))
			if e == nil {
				channel.Key = encrypted
			} else {
				logger.SysError("encrypt channel key on update: " + e.Error())
			}
		}
	}
	err = DB.Model(channel).Updates(channel).Error
	if err != nil {
		return err
	}
	DB.Model(channel).First(channel, "id = ?", channel.Id)
	err = channel.UpdateAbilities()
	return err
}

func (channel *Channel) UpdateResponseTime(responseTime int64) {
	err := DB.Model(channel).Select("response_time", "test_time").Updates(Channel{
		TestTime:     helper.GetTimestamp(),
		ResponseTime: int(responseTime),
	}).Error
	if err != nil {
		logger.SysError("failed to update response time: " + err.Error())
	}
}

func (channel *Channel) UpdateBalance(balance float64) {
	err := DB.Model(channel).Select("balance_updated_time", "balance").Updates(Channel{
		BalanceUpdatedTime: helper.GetTimestamp(),
		Balance:            balance,
	}).Error
	if err != nil {
		logger.SysError("failed to update balance: " + err.Error())
		return
	}

	// Check balance thresholds (logged; actual disable handled by caller)
	if channel.BalanceDisableThreshold > 0 && balance < channel.BalanceDisableThreshold {
		if channel.Status != ChannelStatusAutoDisabled {
			logger.SysLog(fmt.Sprintf("channel #%d [%s] balance %.2f below disable threshold %.2f",
				channel.Id, channel.Name, balance, channel.BalanceDisableThreshold))
		}
	} else if channel.BalanceAlertThreshold > 0 && balance < channel.BalanceAlertThreshold {
		logger.SysLog(fmt.Sprintf("channel #%d [%s] balance %.2f below alert threshold %.2f",
			channel.Id, channel.Name, balance, channel.BalanceAlertThreshold))
	}
}

func (channel *Channel) Delete() error {
	var err error
	err = DB.Delete(channel).Error
	if err != nil {
		return err
	}
	err = channel.DeleteAbilities()
	return err
}

func (channel *Channel) LoadConfig() (ChannelConfig, error) {
	var cfg ChannelConfig
	if channel.Config == "" {
		return cfg, nil
	}
	err := json.Unmarshal([]byte(channel.Config), &cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}

func UpdateChannelStatusById(id int, status int) {
	err := UpdateAbilityStatus(id, status == ChannelStatusEnabled)
	if err != nil {
		logger.SysError("failed to update ability status: " + err.Error())
	}
	err = DB.Model(&Channel{}).Where("id = ?", id).Update("status", status).Error
	if err != nil {
		logger.SysError("failed to update channel status: " + err.Error())
	}
}

func UpdateChannelUsedQuota(id int, quota int64) {
	if config.BatchUpdateEnabled {
		addNewRecord(BatchUpdateTypeChannelUsedQuota, id, quota)
		return
	}
	updateChannelUsedQuota(id, quota)
}

func updateChannelUsedQuota(id int, quota int64) {
	err := DB.Model(&Channel{}).Where("id = ?", id).Update("used_quota", gorm.Expr("used_quota + ?", quota)).Error
	if err != nil {
		logger.SysError("failed to update channel used quota: " + err.Error())
	}
}

func DeleteChannelByStatus(status int64) (int64, error) {
	result := DB.Where("status = ?", status).Delete(&Channel{})
	return result.RowsAffected, result.Error
}

func DeleteDisabledChannel() (int64, error) {
	result := DB.Where("status = ? or status = ?", ChannelStatusAutoDisabled, ChannelStatusManuallyDisabled).Delete(&Channel{})
	return result.RowsAffected, result.Error
}
