package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/quantumclaw/quantumclaw/common"
)

func cacheGetTokenByKey(key string) (*Token, error) {
	objStr, err := common.RedisGet(fmt.Sprintf("token:%s", key))
	if err != nil {
		return nil, err
	}
	var token Token
	err = json.Unmarshal([]byte(objStr), &token)
	return &token, err
}

func cacheSetToken(token Token) error {
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}
	return common.RedisSet(fmt.Sprintf("token:%s", token.Key), string(data), time.Duration(TokenCacheSeconds)*time.Second)
}

func cacheDeleteToken(key string) error {
	return common.RedisDel(fmt.Sprintf("token:%s", key))
}

func cacheIncrTokenQuota(key string, delta int64) error {
	return common.RedisIncr(fmt.Sprintf("token_quota:%s", key), delta)
}

func cacheDecrTokenQuota(key string, delta int64) error {
	return common.RedisIncr(fmt.Sprintf("token_quota:%s", key), -delta)
}

func shouldUpdateRedis(fromDB bool, err error) bool {
	if !common.RedisEnabled {
		return false
	}
	if fromDB && err == nil {
		return true
	}
	return false
}
