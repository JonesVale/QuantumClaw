package common

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/quantumclaw/quantumclaw/common/logger"
)

var RDB redis.Cmdable
var RedisEnabled = true

// InitRedisClient This function is called after init()
func InitRedisClient() (err error) {
	// 优先使用 REDIS_CONN_STRING，否则根据 REDIS_HOST/PORT/PASSWORD 构造
	redisConnString := os.Getenv("REDIS_CONN_STRING")
	if redisConnString == "" {
		redisHost := os.Getenv("REDIS_HOST")
		redisPort := os.Getenv("REDIS_PORT")
		redisPassword := os.Getenv("REDIS_PASSWORD")
		
		if redisHost == "" {
			RedisEnabled = false
			logger.SysLog("REDIS_HOST not set, Redis is not enabled")
			return nil
		}
		if redisPort == "" {
			redisPort = "6379"
		}
		if redisPassword == "" {
			redisConnString = "redis://" + redisHost + ":" + redisPort + "/0"
		} else {
			redisConnString = "redis://:" + redisPassword + "@" + redisHost + ":" + redisPort + "/0"
		}
	}
	
	// 设置 REDIS_CONN_STRING 供后续使用
	os.Setenv("REDIS_CONN_STRING", redisConnString)
	if os.Getenv("SYNC_FREQUENCY") == "" {
		RedisEnabled = false
		logger.SysLog("SYNC_FREQUENCY not set, Redis is disabled")
		return nil
	}
	if os.Getenv("REDIS_MASTER_NAME") == "" {
		logger.SysLog("Redis is enabled")
		opt, err := redis.ParseURL(redisConnString)
		if err != nil {
			logger.FatalLog("failed to parse Redis connection string: " + err.Error())
		}
		RDB = redis.NewClient(opt)
	} else {
		// cluster mode
		logger.SysLog("Redis cluster mode enabled")
		RDB = redis.NewUniversalClient(&redis.UniversalOptions{
			Addrs:      strings.Split(redisConnString, ","),
			Password:   os.Getenv("REDIS_PASSWORD"),
			MasterName: os.Getenv("REDIS_MASTER_NAME"),
		})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = RDB.Ping(ctx).Result()
	if err != nil {
		logger.FatalLog("Redis ping test failed: " + err.Error())
	}
	return err
}

func ParseRedisOption() *redis.Options {
	opt, err := redis.ParseURL(os.Getenv("REDIS_CONN_STRING"))
	if err != nil {
		logger.FatalLog("failed to parse Redis connection string: " + err.Error())
	}
	return opt
}

func RedisSet(key string, value string, expiration time.Duration) error {
	ctx := context.Background()
	return RDB.Set(ctx, key, value, expiration).Err()
}

func RedisGet(key string) (string, error) {
	ctx := context.Background()
	return RDB.Get(ctx, key).Result()
}

func RedisDel(key string) error {
	ctx := context.Background()
	return RDB.Del(ctx, key).Err()
}

func RedisIncr(key string, value int64) error {
	ctx := context.Background()
	return RDB.IncrBy(ctx, key, value).Err()
}

func RedisDecrease(key string, value int64) error {
	ctx := context.Background()
	return RDB.DecrBy(ctx, key, value).Err()
}
