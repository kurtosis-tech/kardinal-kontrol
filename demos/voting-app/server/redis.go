package server

import (
	"context"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/redis/go-redis/v9"
	"os"
)

const (
	REDIS_URI_ENV_VAR = "REDIS_HOST"
)

type RedisConnection struct {
	rdb *redis.Client
}

func NewRedisConnection(ctx context.Context) (*RedisConnection, error) {
	redisUri := os.Getenv(REDIS_URI_ENV_VAR)
	opt, err := redis.ParseURL(redisUri)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred parsing redis URI: %v", redisUri)
	}
	rdb := redis.NewClient(opt)

	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred attempting to ping redis service at uri: %v", redisUri)
	}
	return &RedisConnection{
		rdb: rdb,
	}, nil
}

func (r *RedisConnection) Persist() {
	return
}
