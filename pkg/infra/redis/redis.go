package redis_wrapper

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisConfig struct {
	ConnectionURL       string `yaml:"connection_url"`
	PoolSize            int    `yaml:"pool_size"`
	DialTimeoutSeconds  int    `yaml:"dial_timeout_seconds"`
	ReadTimeoutSeconds  int    `yaml:"read_timeout_seconds"`
	WriteTimeoutSeconds int    `yaml:"write_timeout_seconds"`
	IdleTimeoutSeconds  int    `yaml:"idle_timeout_seconds"`
}

// InitRedis create a redis from config
func InitRedis(redisCfg *RedisConfig) (*redis.Client, error) {
	var redisClient *redis.Client

	opts, err := redis.ParseURL(redisCfg.ConnectionURL)
	if err != nil {
		zap.S().Debugf("parse redis url fail: %+v", err)
		return nil, err
	}

	opts.PoolSize = redisCfg.PoolSize
	opts.DialTimeout = time.Duration(redisCfg.DialTimeoutSeconds) * time.Second
	opts.ReadTimeout = time.Duration(redisCfg.ReadTimeoutSeconds) * time.Second
	opts.WriteTimeout = time.Duration(redisCfg.WriteTimeoutSeconds) * time.Second
	opts.ConnMaxIdleTime = time.Duration(redisCfg.IdleTimeoutSeconds) * time.Second

	redisClient = redis.NewClient(opts)

	cmd := redisClient.Ping(context.Background())
	if cmd.Err() != nil {
		return nil, cmd.Err()
	}

	zap.S().Debug("connect to redis successful")
	return redisClient, nil
}
