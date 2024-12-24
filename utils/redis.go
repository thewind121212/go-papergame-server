package utils

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

// Redis client instance
var redisClient *redis.Client

// Initialize Redis client once
func InitRedisClient(addr string, password string, db int) {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
}

func PingRedis() error {
	result, err := redisClient.Ping(context.Background()).Result()
	fmt.Println(result)
	return err
}

func SetKey(key string, value string, exp time.Duration) error {
	return redisClient.Set(context.Background(), key, value, exp).Err()
}

func DelKey(key string) error {
	return redisClient.Del(context.Background(), key).Err()
}

func GetKey(key string) (string, error) {
	return redisClient.Get(context.Background(), key).Result()
}
