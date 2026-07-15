package booking

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Service interface {
	BookSeat(ctx context.Context, userID string, req BookSeatRequest) error
}

type service struct {
	repo        Repository
	redisClient *redis.Client
}

func NewService(repo Repository, redisClient *redis.Client) Service {
	return &service{repo: repo, redisClient: redisClient}
}

func (s *service) BookSeat(ctx context.Context, userID string, req BookSeatRequest) error {
	// สร้าง Key สำหรับล็อคที่นั่งนี้ (เช่น "lock:seat:105")
	lockKey := fmt.Sprintf("lock:seat:%d", req.SeatID)

	// 1. Acquire Redis Lock (SET NX PX)
	// SET NX = เซ็ตค่าถ้า Key นี้ยังไม่มี, PX = หมดอายุใน 5 วินาที (กันระบบค้างแล้วล็อคคา)
	locked, err := s.redisClient.SetNX(ctx, lockKey, userID, 5*time.Second).Result()
	if err != nil {
		return errors.New("internal server error during locking")
	}
	if !locked {
		// ถ้า set ไม่สำเร็จ แปลว่ามีคนอื่นกำลังจองที่นั่งนี้อยู่ (Fast fail)
		return errors.New("seat is currently being booked by someone else")
	}

	// 2. Release Redis Lock เสมอเมื่อจบการทำงาน (ใช้ defer)
	// ใช้ context.Background() เพื่อให้แน่ใจว่าคำสั่งลบจะทำงานแม้ ctx หลักจะ timeout ไปแล้ว
	defer s.redisClient.Del(context.Background(), lockKey)

	// 3. ไปทำรายการ Database Transaction (ที่มี FOR UPDATE อีกชั้น)
	err = s.repo.BookSeatTx(ctx, userID, req.EventID, req.SeatID)
	if err != nil {
		return err // หาก Database คืนค่า error (เช่น จองแล้ว) ก็ให้จบการทำงานแค่นี้
	}

	return nil
}