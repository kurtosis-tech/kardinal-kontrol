package server

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type RedisConnection struct {
	rdb *redis.Client
}

func NewRedisConnection(ctx context.Context) (*RedisConnection, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
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
