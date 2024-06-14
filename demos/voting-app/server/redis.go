package server

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type RedisConnection struct {
	rdb *redis.Client
}

func NewRedisConnection(ctx context.Context) (*RedisConnection, error) {
	// get redis uri from environment
	rdb := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:64433",
		Password: "",
		DB:       0,
	})
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	return &RedisConnection{
		rdb: rdb,
	}, nil
}

func (r *RedisConnection) Persist() {
	return
}
