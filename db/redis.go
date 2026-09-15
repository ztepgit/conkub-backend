// db/redis.go
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func InitRedis(url, password string) (*redis.Client, error) {
	// เช็ค URL ว่าง ให้ return error ทันที
	if url == "" {
		return nil, fmt.Errorf("Redis URL is empty")
	}

	client := redis.NewClient(&redis.Options{
		Addr:         url,
		Password:     password,
		DB:           0,
		PoolSize:     100, // รองรับ High Concurrency (คงโค้ดเดิมของคุณไว้)
		MinIdleConns: 10,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// เช็ค Ping ถ้าพังให้ return error ทันที
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}