package model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/encrypt"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/common/random"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	TokenCacheSeconds         = config.SyncFrequency
	UserId2GroupCacheSeconds  = config.SyncFrequency
	UserId2QuotaCacheSeconds  = config.SyncFrequency
	UserId2StatusCacheSeconds = config.SyncFrequency
	GroupModelsCacheSeconds   = config.SyncFrequency
)

func CacheGetTokenByKey(keyHash string) (*Token, error) {
	// 所有调用方已传入 SHA256(raw_key)，直接用哈希查询
	var token Token
	if !common.RedisEnabled {
		err := DB.Where("key_hash = ?", keyHash).First(&token).Error
		return &token, err
	}
	tokenObjectString, err := common.RedisGet(fmt.Sprintf("token:%s", keyHash))
	if err != nil {
		err := DB.Where("key_hash = ?", keyHash).First(&token).Error
		if err != nil {
			return nil, err
		}
		jsonBytes, err := json.Marshal(token)
		if err != nil {
			return nil, err
		}
		err = common.RedisSet(fmt.Sprintf("token:%s", keyHash), string(jsonBytes), time.Duration(TokenCacheSeconds)*time.Second)
		if err != nil {
			logger.SysError("Redis set token error: " + err.Error())
		}
		return &token, nil
	}
	err = json.Unmarshal([]byte(tokenObjectString), &token)
	return &token, err
}

func CacheGetUserGroup(id int) (group string, err error) {
	if !common.RedisEnabled {
		return GetUserGroup(id)
	}
	group, err = common.RedisGet(fmt.Sprintf("user_group:%d", id))
	if err != nil {
		group, err = GetUserGroup(id)
		if err != nil {
			return "", err
		}
		err = common.RedisSet(fmt.Sprintf("user_group:%d", id), group, time.Duration(UserId2GroupCacheSeconds)*time.Second)
		if err != nil {
			logger.SysError("Redis set user group error: " + err.Error())
		}
	}
	return group, err
}

func fetchAndUpdateUserQuota(ctx context.Context, id int) (quota int64, err error) {
	quota, err = GetUserQuota(id)
	if err != nil {
		return 0, err
	}
	err = common.RedisSet(fmt.Sprintf("user_quota:%d", id), fmt.Sprintf("%d", quota), time.Duration(UserId2QuotaCacheSeconds)*time.Second)
	if err != nil {
		logger.Error(ctx, "Redis set user quota error: "+err.Error())
	}
	return
}

func CacheGetUserQuota(ctx context.Context, id int) (quota int64, err error) {
	if !common.RedisEnabled {
		return GetUserQuota(id)
	}
	quotaString, err := common.RedisGet(fmt.Sprintf("user_quota:%d", id))
	if err != nil {
		return fetchAndUpdateUserQuota(ctx, id)
	}
	quota, err = strconv.ParseInt(quotaString, 10, 64)
	if err != nil {
		return 0, nil
	}
	if quota <= config.PreConsumedQuota { // when user's quota is less than pre-consumed quota, we need to fetch from db
		logger.Infof(ctx, "user %d's cached quota is too low: %d, refreshing from db", quota, id)
		return fetchAndUpdateUserQuota(ctx, id)
	}
	return quota, nil
}

func CacheUpdateUserQuota(ctx context.Context, id int) error {
	if !common.RedisEnabled {
		return nil
	}
	quota, err := CacheGetUserQuota(ctx, id)
	if err != nil {
		return err
	}
	err = common.RedisSet(fmt.Sprintf("user_quota:%d", id), fmt.Sprintf("%d", quota), time.Duration(UserId2QuotaCacheSeconds)*time.Second)
	return err
}

func CacheDecreaseUserQuota(id int, quota int64) error {
	if !common.RedisEnabled {
		return nil
	}
	err := common.RedisDecrease(fmt.Sprintf("user_quota:%d", id), int64(quota))
	return err
}

func CacheIsUserEnabled(userId int) (bool, error) {
	if !common.RedisEnabled {
		return IsUserEnabled(userId)
	}
	enabled, err := common.RedisGet(fmt.Sprintf("user_enabled:%d", userId))
	if err == nil {
		return enabled == "1", nil
	}

	userEnabled, err := IsUserEnabled(userId)
	if err != nil {
		return false, err
	}
	enabled = "0"
	if userEnabled {
		enabled = "1"
	}
	err = common.RedisSet(fmt.Sprintf("user_enabled:%d", userId), enabled, time.Duration(UserId2StatusCacheSeconds)*time.Second)
	if err != nil {
		logger.SysError("Redis set user enabled error: " + err.Error())
	}
	return userEnabled, err
}

func CacheGetGroupModels(ctx context.Context, group string) ([]string, error) {
	if !common.RedisEnabled {
		return GetGroupModels(ctx, group)
	}
	modelsStr, err := common.RedisGet(fmt.Sprintf("group_models:%s", group))
	if err == nil {
		return strings.Split(modelsStr, ","), nil
	}
	models, err := GetGroupModels(ctx, group)
	if err != nil {
		return nil, err
	}
	err = common.RedisSet(fmt.Sprintf("group_models:%s", group), strings.Join(models, ","), time.Duration(GroupModelsCacheSeconds)*time.Second)
	if err != nil {
		logger.SysError("Redis set group models error: " + err.Error())
	}
	return models, nil
}

var group2model2channels map[string]map[string][]*Channel
var group2model2ownerChannels map[string]map[string]map[int][]*Channel
var channelSyncLock sync.RWMutex

func InitChannelCache() {
	newChannelId2channel := make(map[int]*Channel)
	var channels []*Channel
	DB.Where("status = ?", ChannelStatusEnabled).Find(&channels)

	// Decrypt API keys for relay usage
	// Keys in DB are AES-256-GCM encrypted if CRYPTO_SECRET is set
	if config.CryptoSecret != "" {
		decryptCount := 0
		for i := range channels {
			if channels[i].Key != "" {
				decrypted, err := encrypt.Decrypt(channels[i].Key, encrypt.DeriveKey(config.CryptoSecret))
				if err == nil {
					channels[i].Key = string(decrypted)
					decryptCount++
				} else if strings.HasPrefix(channels[i].Key, "sk-") || strings.HasPrefix(channels[i].Key, "gsk_") {
					// Plaintext key that starts with expected pattern - already plaintext
					logger.SysLog("InitChannelCache: ch#" + fmt.Sprint(channels[i].Id) + " key is already plaintext")
				} else if strings.Contains(channels[i].Key, "PUT_YOUR") {
					// Placeholder - skip
				} else {
					logger.SysWarn("InitChannelCache: decrypt failed for ch#" + fmt.Sprint(channels[i].Id) + ": " + err.Error())
				}
			}
		}
		logger.SysLog(fmt.Sprintf("InitChannelCache: decrypted %d keys", decryptCount))
	} else {
		logger.SysWarn("InitChannelCache: CRYPTO_SECRET is empty, keys will not be decrypted")
	}

	// 过滤掉未配置 API Key 的渠道（占位符或空 key）
	var activeChannels []*Channel
	for _, ch := range channels {
		if ch.Key == "" || strings.HasPrefix(ch.Key, "PUT_YOUR") {
			continue
		}
		activeChannels = append(activeChannels, ch)
	}
	channels = activeChannels

	for _, channel := range channels {
		newChannelId2channel[channel.Id] = channel
	}
	var abilities []*Ability
	DB.Find(&abilities)
	groups := make(map[string]bool)
	for _, ability := range abilities {
		groups[ability.Group] = true
	}
	newGroup2model2channels := make(map[string]map[string][]*Channel)
	newGroup2model2ownerChannels := make(map[string]map[string]map[int][]*Channel)
	for group := range groups {
		newGroup2model2channels[group] = make(map[string][]*Channel)
		newGroup2model2ownerChannels[group] = make(map[string]map[int][]*Channel)
	}
	for _, channel := range channels {
		channelGroups := strings.Split(channel.Group, ",")
		for _, group := range channelGroups {
			// 确保外层 map 存在
			if _, ok := newGroup2model2channels[group]; !ok {
				newGroup2model2channels[group] = make(map[string][]*Channel)
				newGroup2model2ownerChannels[group] = make(map[string]map[int][]*Channel)
			}
			models := strings.Split(channel.Models, ",")
			for _, model := range models {
				// 全量索引
				if _, ok := newGroup2model2channels[group][model]; !ok {
					newGroup2model2channels[group][model] = make([]*Channel, 0)
				}
				newGroup2model2channels[group][model] = append(newGroup2model2channels[group][model], channel)

				// 按 owner 索引
				if _, ok := newGroup2model2ownerChannels[group][model]; !ok {
					newGroup2model2ownerChannels[group][model] = make(map[int][]*Channel)
				}
				ownerId := channel.UserId
				newGroup2model2ownerChannels[group][model][ownerId] = append(
					newGroup2model2ownerChannels[group][model][ownerId], channel)
			}
		}
	}

	// sort by priority
	for group, model2channels := range newGroup2model2channels {
		for model, channels := range model2channels {
			sort.Slice(channels, func(i, j int) bool {
				return channels[i].GetPriority() > channels[j].GetPriority()
			})
			newGroup2model2channels[group][model] = channels
		}
	}
	for group, model2owner := range newGroup2model2ownerChannels {
		for model, ownerMap := range model2owner {
			for ownerId, channels := range ownerMap {
				sort.Slice(channels, func(i, j int) bool {
					return channels[i].GetPriority() > channels[j].GetPriority()
				})
				newGroup2model2ownerChannels[group][model][ownerId] = channels
			}
		}
	}

	channelSyncLock.Lock()
	group2model2channels = newGroup2model2channels
	group2model2ownerChannels = newGroup2model2ownerChannels
	channelSyncLock.Unlock()
	logger.SysLog("channels synced from database")
}

func SyncChannelCache(frequency int) {
	for {
		time.Sleep(time.Duration(frequency) * time.Second)
		logger.SysLog("syncing channels from database")
		InitChannelCache()
	}
}

func CacheGetRandomSatisfiedChannel(group string, model string, ignoreFirstPriority bool) (*Channel, error) {
	if !config.MemoryCacheEnabled {
		return GetRandomSatisfiedChannel(group, model, ignoreFirstPriority)
	}
	channelSyncLock.RLock()
	defer channelSyncLock.RUnlock()
	channels := group2model2channels[group][model]
	if len(channels) == 0 {
		return nil, errors.New("channel not found")
	}
	endIdx := len(channels)
	// choose by priority
	firstChannel := channels[0]
	if firstChannel.GetPriority() > 0 {
		for i := range channels {
			if channels[i].GetPriority() != firstChannel.GetPriority() {
				endIdx = i
				break
			}
		}
	}
	idx := rand.Intn(endIdx)
	if ignoreFirstPriority {
		if endIdx < len(channels) { // which means there are more than one priority
			idx = random.RandRange(endIdx, len(channels))
		}
	}
	return channels[idx], nil
}

// CacheGetRandomSatisfiedChannelByOwner 按 owner 查缓存
// ownerId=0 → 平台渠道, ownerId>0 → 渠道商
func CacheGetRandomSatisfiedChannelByOwner(group string, model string, ownerId int) (*Channel, error) {
	if !config.MemoryCacheEnabled {
		return GetRandomSatisfiedChannelByOwner(group, model, ownerId)
	}
	channelSyncLock.RLock()
	defer channelSyncLock.RUnlock()

	ownerMap := group2model2ownerChannels[group][model]
	if ownerMap == nil {
		return nil, errors.New("channel not found for model")
	}
	channels := ownerMap[ownerId]
	if len(channels) == 0 {
		return nil, errors.New("channel not found for owner")
	}
	// 取最高 priority 中的随机一个
	endIdx := len(channels)
	first := channels[0]
	if first.GetPriority() > 0 {
		for i := range channels {
			if channels[i].GetPriority() != first.GetPriority() {
				endIdx = i
				break
			}
		}
	}
	return channels[rand.Intn(endIdx)], nil
}

// CacheGetRandomSatisfiedChannelAnyOwner 从全资源池查（不限 owner）
func CacheGetRandomSatisfiedChannelAnyOwner(group string, model string, excludeOwnerId int) (*Channel, error) {
	if !config.MemoryCacheEnabled {
		return GetRandomSatisfiedChannelAnyOwner(group, model, excludeOwnerId)
	}
	channelSyncLock.RLock()
	defer channelSyncLock.RUnlock()

	allChannels := group2model2channels[group][model]
	if len(allChannels) == 0 {
		return nil, errors.New("channel not found")
	}

	// 排除已试过的 owner
	var candidates []*Channel
	for _, ch := range allChannels {
		if ch.UserId != excludeOwnerId {
			candidates = append(candidates, ch)
		}
	}
	if len(candidates) == 0 {
		return nil, errors.New("no alternate channel available")
	}

	// 取最高 priority 中的随机一个
	endIdx := len(candidates)
	first := candidates[0]
	if first.GetPriority() > 0 {
		for i := range candidates {
			if candidates[i].GetPriority() != first.GetPriority() {
				endIdx = i
				break
			}
		}
	}
	return candidates[rand.Intn(endIdx)], nil
}
