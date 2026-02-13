package queue

import (
	"context"
	"github.com/redis/go-redis/v9"
	"os"
	"fmt"
	"time"
)

// RedisClient обертка над redis.Client
type RedisClient struct {
	Client *redis.Client
}

// NewRedisClient создает новое подключение к Redis
func NewRedisClient() (*RedisClient, error) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	// Проверка подключения
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
    defer cancel()

    if err := client.Ping(ctx).Err(); err != nil {
        return nil, fmt.Errorf("Redis недоступен: %w", err)
    }

	return &RedisClient{Client: client}, nil
}
