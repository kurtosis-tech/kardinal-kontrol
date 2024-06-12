package server

type RedisConnection struct {
}

func NewRedisConnection() (*RedisConnection, error) {
	// connect to redis instance
	// return
	return &RedisConnection{}, nil
}

func (r *RedisConnection) Persist() {
	return
}
