package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/quantumclaw/quantumclaw/common"
)

func cacheGetTokenByKey(keyHash string) (*Token, error) {
	objStr, err := common.RedisGet(fmt.Sprintf("token:%s", keyHash))
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
	return common.RedisSet(fmt.Sprintf("token:%s", token.KeyHash), string(data), time.Duration(TokenCacheSeconds)*time.Second)
}

func cacheDeleteToken(keyHash string) error {
	return common.RedisDel(fmt.Sprintf("token:%s", keyHash))
}

func cacheIncrTokenQuota(keyHash string, delta int64) error {
	return common.RedisIncr(fmt.Sprintf("token_quota:%s", keyHash), delta)
}

func cacheDecrTokenQuota(keyHash string, delta int64) error {
	return common.RedisIncr(fmt.Sprintf("token_quota:%s", keyHash), -delta)
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
