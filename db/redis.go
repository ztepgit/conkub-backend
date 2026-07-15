package db

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

func InitRedis(url, password string) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:         url,
		Password:     password,
		DB:           0,
		PoolSize:     100, // รองรับ High Concurrency
		MinIdleConns: 10,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Redis connected successfully")
	return client
}