package booking

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type Service interface {
	// เปลี่ยนให้ return (string, error) เพื่อส่ง Stripe Checkout URL กลับไป
	BookSeat(ctx context.Context, userID string, req BookSeatRequest) (string, error)
	
	// สำหรับให้ Webhook เรียกใช้เมื่อจ่ายเงินสำเร็จ
	ConfirmBooking(ctx context.Context, seatID uint) error 
}

type service struct {
	repo        Repository
	redisClient *redis.Client
}

func NewService(repo Repository, redisClient *redis.Client) Service {
	return &service{repo: repo, redisClient: redisClient}
}

func (s *service) BookSeat(ctx context.Context, userID string, req BookSeatRequest) (string, error) {
	// สร้าง Key สำหรับล็อคที่นั่งนี้ (เช่น "lock:seat:105")
	lockKey := fmt.Sprintf("lock:seat:%d", req.SeatID)

	// 1. เช็คว่ามี Redis ให้ใช้ไหม ถ้าไม่มีให้ข้ามไป (Bypass)
	if s.redisClient != nil {
		// Acquire Redis Lock (SET NX PX)
		// SET NX = เซ็ตค่าถ้า Key นี้ยังไม่มี, PX = หมดอายุใน 5 วินาที
		locked, err := s.redisClient.SetNX(ctx, lockKey, userID, 5*time.Second).Result()
		if err != nil {
			return "", errors.New("internal server error during locking")
		}
		if !locked {
			// ถ้า set ไม่สำเร็จ แปลว่ามีคนอื่นกำลังจองที่นั่งนี้อยู่ (Fast fail)
			return "", errors.New("seat is currently being booked by someone else")
		}

		// Release Redis Lock เสมอเมื่อจบการทำงาน (เฉพาะตอนที่มี Redis)
		defer s.redisClient.Del(context.Background(), lockKey)
	} else {
		// ข้าม Redis lock ชั่วคราว
		log.Println("Redis disabled, skipping seat lock")
	}

	// สมมติว่าดึงราคามาจาก DB ได้ 2500 บาท
	// (ในโปรเจกต์จริง ควรดึงราคาของที่นั่งรหัส req.SeatID จากตาราง seats)
	price := 2500.00

	// 3. ไปทำรายการ Database Transaction (สร้าง Booking และเปลี่ยนสถานะที่นั่งเป็น PENDING)
	// 🔴 สังเกต: เราปรับให้ BookSeatTx คืนค่า booking object กลับมาด้วยเพื่อนำ ID ไปใช้
	booking, err := s.repo.BookSeatTx(ctx, userID, req.EventID, req.SeatID)
	if err != nil {
		return "", err
	}

	// 4. สร้าง Stripe Checkout URL
	checkoutURL, err := CreateStripeCheckout(req.EventID, req.SeatID, price, userID)
	if err != nil {
		// 🔴 หากสร้าง URL จ่ายเงินไม่สำเร็จ ให้ทำการชดเชย (Compensate) โดยเรียก CancelBooking
		log.Printf("[BookingService] Stripe checkout failed for booking %d, seat %d. Reverting: %v", booking.ID, req.SeatID, err)
		
		cancelErr := s.repo.CancelBooking(ctx, booking.ID, req.SeatID)
		if cancelErr != nil {
			// หากยกเลิกไม่สำเร็จ ต้องมี Log ที่ชัดเจนเพื่อให้ Admin ตรวจสอบ (Manual Intervention)
			log.Printf("[CRITICAL] Failed to revert seat %d to AVAILABLE for booking %d: %v", req.SeatID, booking.ID, cancelErr)
		}

		// Return error ส่งกลับไปหา Client
		return "", fmt.Errorf("failed to create payment session, booking cancelled")
	}
	
	return checkoutURL, nil
}

func (s *service) ConfirmBooking(ctx context.Context, seatID uint) error {
	return s.repo.ConfirmBooking(ctx, seatID)
}