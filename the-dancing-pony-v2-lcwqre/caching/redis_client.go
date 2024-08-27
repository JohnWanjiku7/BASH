package cache

import (
	"context"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

// RedisClient is a global Redis client instance
var RedisClient *redis.Client

func InitRedisClient(addr, password string) error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
	return nil
}
