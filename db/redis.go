package db

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

func InitRedis(url, password string) *redis.Client {
	// 1. เพิ่มเช็ค URL ว่าง ให้ข้ามการเชื่อมต่อ
	if url == "" {
		log.Println("Redis disabled: URL is empty")
		return nil
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

	// 2. เปลี่ยนจาก Fatalf เป็น Printf และ return nil เพื่อไม่ให้ Backend แครช
	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("Redis unavailable, running without Redis: %v\n", err)
		return nil
	}

	log.Println("Redis connected successfully")
	return client
}