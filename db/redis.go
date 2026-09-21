// db/redis.go
package db

import (
	"context"
	"crypto/tls" // 🔴 1. เพิ่ม import crypto/tls
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
		PoolSize:     100, // รองรับ High Concurrency (โค้ดเดิม)
		MinIdleConns: 10,
		// 🔴 2. เพิ่ม TLSConfig เพื่อให้เชื่อมต่อกับ Upstash (บน Cloud) ได้
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// เช็ค Ping ถ้าพังให้ return error ทันที
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}