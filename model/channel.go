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

// ============================================================
// Channel — AI 渠道模型（API Provider 连接配置）
//
// 字段按功能域分组说明（31 个字段）：
//
//	[Core]        Id, Type, Key, Status, Name, Weight, CreatedTime, DeletedAt       (8 基础字段)
//	[Connection]  BaseURL, Models, Config, SystemPrompt, Category, Other             (6 连接字段)
//	[Pricing]     CostPerUnit, SellPriceRate, CostPrice, ProfitSplit, ChannelMarkup  (5 定价字段)
//	[Balance]     Balance, BalanceUpdatedTime, UsedQuota, BalanceAlertThreshold,
//	              BalanceDisableThreshold                                           (5 余额字段)
//	[Routing]     Group, Priority, ModelMapping                                      (3 路由字段)
//	[Ownership]   UserId, StoreID, IsPlatformPool, Region                            (4 归属字段)
//	[Test]        TestTime, ResponseTime, LastTestPassed, LastErrorMessage           (4 测试字段)
//
// TODO: 后续版本考虑将 Pricing/Balance/Routing/Ownership/Test 抽取为 embedded 子结构体，
//       使用 GORM 的 embedded tag 保持数据库列名兼容。参见下方的 ChannelPricing 等类型定义。
// ============================================================
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
	UserId             int     `json:"user_id" gorm:"type:int;default:0;index"` // 娓犻亾褰掑睘锛?=骞冲彴锛?0=渚涘簲鍟?
	CostPrice          float64 `json:"cost_price" gorm:"type:decimal(10,6);default:0"` // Key 璐＄尞鑰呭疄闄呮垚鏈?

	// 浣欓棰勮涓庤嚜鍔ㄧ鐢ㄩ槇鍊硷紙鍗曚綅涓?USD锛?
	BalanceAlertThreshold   float64 `json:"balance_alert_threshold" gorm:"type:decimal(10,2);default:0"`
	BalanceDisableThreshold float64 `json:"balance_disable_threshold" gorm:"type:decimal(10,2);default:0"`
	ChannelMarkup          float64 `json:"channel_markup" gorm:"type:decimal(5,2);default:1.0"`      // 娓犻亾鍔犱环鍊嶇巼锛?.0=鍘熶环, 1.2=+20%锛?

	// 鍖哄煙鏍囪瘑锛歝hina / overseas / ""锛堣嚜鍔ㄥ垽瀹氾級
	Region string `json:"region" gorm:"type:varchar(20);default:''"`

	// 杞垹闄わ紙渚涘簲鍟嗗垹闄ゆ椂鏍囪锛屼笉鐗╃悊鍒犻櫎锛?
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty"`

	// 娴嬭瘯鐘舵€?
	LastTestPassed   bool   `json:"last_test_passed" gorm:"default:false"`      // 鏈€杩戜竴娆℃祴璇曟槸鍚﹂€氳繃
	LastErrorMessage string `json:"last_error_message" gorm:"type:text;default:''"` // 鏈€杩戜竴娆￠敊璇俊鎭?

	// 鍒嗚处姣斾緥锛?.0~1.0锛屾笭閬撳晢鎵€寰楁瘮渚嬶紝鍓╀綑涓哄钩鍙版娊鎴?
	// 榛樿 0.85 = 娓犻亾鍟嗗緱 85%锛屽钩鍙板緱 15%
	ProfitSplit float64 `json:"profit_split" gorm:"type:decimal(4,3);default:0.85"`

	// 鎵€灞炲簵閾猴紙甯傚満鐗堢敤锛夛細绌?= 浼犵粺娓犻亾锛岄潪绌?= 鍏宠仈搴楅摵
	StoreID int `json:"store_id" gorm:"type:int;default:0;index"`

	// 骞冲彴鍏滃簳姹犳爣璁帮細true = 骞冲彴鑷惀淇濆簳璧勬簮锛屼笉璁″叆甯傚満鎺掑悕
	IsPlatformPool bool `json:"is_platform_pool" gorm:"default:false"`
}

type ChannelConfig struct {
	Region            string  `json:"region,omitempty"`
	SK                string  `json:"sk,omitempty"`
	AK                string  `json:"ak,omitempty"`
	UserID            string  `json:"user_id,omitempty"`
	APIVersion        string  `json:"api_version,omitempty"`
	LibraryID         string  `json:"library_id,omitempty"`
	Plugin            string  `json:"plugin,omitempty"`
	VertexAIProjectID string  `json:"vertex_ai_project_id,omitempty"`
	VertexAIADC       string  `json:"vertex_ai_adc,omitempty"`
	CacheBillingRatio float64 `json:"cache_billing_ratio,omitempty"`
	PromptCacheEnabled *bool   `json:"prompt_cache_enabled,omitempty"`
	ThinkingToContent bool    `json:"thinking_to_content,omitempty"`
}

// ============================================================
// Channel 子结构体定义（语义分组）
// 这些类型用于未来重构时替代 Channel 中的扁平字段。
// 当前 Channel 结构体保持扁平以确保 GORM/JSON 向后兼容。
// ============================================================

// ChannelPricing 定价相关字段
type ChannelPricing struct {
	CostPerUnit   float64 `json:"cost_per_unit"` // 每单位成本（USD）
	SellPriceRate float64 `json:"sell_price_rate"` // 加价倍率
	CostPrice     float64 `json:"cost_price"` // Key 提供商实际成本
	ProfitSplit   float64 `json:"profit_split"` // 渠道商分成比例 (0~1)
	ChannelMarkup float64 `json:"channel_markup"` // 渠道加价率 (1.0=原价, 1.2=+20%)
}

// ChannelBalance 余额与用量字段
type ChannelBalance struct {
	Balance                float64 `json:"balance"` // 账户余额（USD）
	BalanceUpdatedTime     int64   `json:"balance_updated_time"`
	UsedQuota              int64   `json:"used_quota"`
	BalanceAlertThreshold  float64 `json:"balance_alert_threshold"`
	BalanceDisableThreshold float64 `json:"balance_disable_threshold"`
}

// ChannelRouting 路由和分组字段
type ChannelRouting struct {
	Group        string  `json:"group"`
	Priority     *int64  `json:"priority"`
	ModelMapping *string `json:"model_mapping"`
}

// ChannelOwnership 归属信息
type ChannelOwnership struct {
	UserId         int    `json:"user_id"` // 0=平台自有, >0=供应商
	StoreID        int    `json:"store_id"`
	IsPlatformPool bool   `json:"is_platform_pool"`
	Region         string `json:"region"` // china / overseas / ""
}

// ChannelTest 测试状态字段
type ChannelTest struct {
	TestTime         int64  `json:"test_time"`
	ResponseTime     int   `json:"response_time"` // ms
	LastTestPassed   bool   `json:"last_test_passed"`
	LastErrorMessage string `json:"last_error_message"`
}

// ToPricing 从 Channel 实例提取定价子集（用于计费逻辑中传递定价信息）
func (ch *Channel) ToPricing() ChannelPricing {
	return ChannelPricing{
		CostPerUnit:   ch.CostPerUnit,
		SellPriceRate: ch.SellPriceRate,
		CostPrice:     ch.CostPrice,
		ProfitSplit:   ch.ProfitSplit,
		ChannelMarkup: ch.ChannelMarkup,
	}
}

// ToBalance 从 Channel 实例提取余额子集
func (ch *Channel) ToBalance() ChannelBalance {
	return ChannelBalance{
		Balance:                ch.Balance,
		BalanceUpdatedTime:     ch.BalanceUpdatedTime,
		UsedQuota:              ch.UsedQuota,
		BalanceAlertThreshold:  ch.BalanceAlertThreshold,
		BalanceDisableThreshold: ch.BalanceDisableThreshold,
	}
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
					decrypted, e := encrypt.DecryptChannelKey(channels[i].Key, config.CryptoSecret)
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
		// 瑙ｅ瘑 API Key锛坰electAll 琛ㄧず闇€瑕佸畬鏁翠俊鎭紝濡?relay 鍦烘櫙锛?
		if err == nil && channel.Key != "" && config.CryptoSecret != "" {
			decrypted, e := encrypt.DecryptChannelKey(channel.Key, config.CryptoSecret)
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
	// 鎵归噺鍔犲瘑 API Key
	for i := range channels {
		if channels[i].Key != "" && config.CryptoSecret != "" {
			encrypted, e := encrypt.EncryptChannelKey(channels[i].Key, config.CryptoSecret)
			if e != nil {
				logger.SysError("batch encrypt channel key: " + e.Error())
				return fmt.Errorf("failed to encrypt channel key: %w", e)
			}
			channels[i].Key = encrypted
		}
		channels[i].LastTestPassed = true
	}
	if err := DB.Create(&channels).Error; err != nil {
		return err
	}
	for _, channel_ := range channels {
		if err := channel_.AddAbilities(); err != nil {
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

	count := 0
	for _, ch := range channels {
		// Check: is this key already encrypted? Try decrypt first
		_, decErr := encrypt.DecryptChannelKey(ch.Key, config.CryptoSecret)
		if decErr == nil {
			continue // Already encrypted
		}
		// Key is plaintext (sk-..., gsk_..., or any non-base64-encrypted string)
		// Only proceed if the key looks like a real API key (not a placeholder)
		if strings.HasPrefix(ch.Key, "PUT_YOUR") || ch.Key == "" {
			continue
		}
		// Encrypt the plaintext key
		encrypted, err := encrypt.EncryptChannelKey(ch.Key, config.CryptoSecret)
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
	// 鍔犲瘑 API Key
	if channel.Key != "" && config.CryptoSecret != "" {
		encrypted, e := encrypt.EncryptChannelKey(channel.Key, config.CryptoSecret)
		if e != nil {
			logger.SysError("encrypt channel key: " + e.Error())
			return fmt.Errorf("failed to encrypt channel key: %w", e)
		}
		channel.Key = encrypted
	}
	channel.LastTestPassed = true
	if err := DB.Create(channel).Error; err != nil {
		return err
	}
	return channel.AddAbilities()
}

func (channel *Channel) Update() error {
	// 濡傛灉 Key 鍙樹簡锛屽姞瀵嗗悗瀛樺偍
	if channel.Key != "" && config.CryptoSecret != "" {
		existing, err := GetChannelById(channel.Id, true)
		if err != nil {
			return fmt.Errorf("failed to get existing channel: %w", err)
		}
		if channel.Key != existing.Key {
			encrypted, e := encrypt.EncryptChannelKey(channel.Key, config.CryptoSecret)
			if e != nil {
				logger.SysError("encrypt channel key on update: " + e.Error())
				return fmt.Errorf("failed to encrypt channel key: %w", e)
			}
			channel.Key = encrypted
		}
	}
	if err := DB.Model(channel).Updates(channel).Error; err != nil {
		return err
	}
	DB.Model(channel).First(channel, "id = ?", channel.Id)
	return channel.UpdateAbilities()
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
