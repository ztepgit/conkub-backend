// booking/service.go
package booking

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/webhook"
)

type Service interface {
	// เปลี่ยนให้ return (string, error) เพื่อส่ง Stripe Checkout URL กลับไป
	BookSeat(ctx context.Context, userID string, req BookSeatRequest) (string, error)
	
	// สำหรับให้ Webhook เรียกใช้เมื่อจ่ายเงินสำเร็จ
	ConfirmBooking(ctx context.Context, seatID uint) error 

	// 🔴 เพิ่ม Interface สำหรับประมวลผล Webhook
	ProcessStripeWebhook(ctx context.Context, payload []byte, signature string) error
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

	
	

	// 3. ไปทำรายการ Database Transaction (สร้าง Booking และเปลี่ยนสถานะที่นั่งเป็น PENDING)
	// 🔴 สังเกต: เราปรับให้ BookSeatTx คืนค่า booking object กลับมาด้วยเพื่อนำ ID ไปใช้
	booking, err := s.repo.BookSeatTx(ctx, userID, req.EventID, req.SeatID)
	if err != nil {
		return "", err
	}

	// 🔴 ดึงราคาที่แท้จริงจาก Database (Source of Truth)
	price := booking.Seat.Price

	// 4. สร้าง Stripe Checkout URL
	checkoutURL, err := CreateStripeCheckout(booking.ID, req.EventID, req.SeatID, price, userID)
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

	// 🔴 [A] ตรวจสอบว่า CreateStripeCheckout คืนค่า URL กลับมาได้หรือไม่
	log.Printf("[Stripe] checkoutURL=%s", checkoutURL)
	
	return checkoutURL, nil
}

func (s *service) ConfirmBooking(ctx context.Context, seatID uint) error {
	return s.repo.ConfirmBooking(ctx, seatID)
}

// 🔴 Implement ฟังก์ชันสำหรับประมวลผล Webhook พร้อมแกะ Metadata
func (s *service) ProcessStripeWebhook(ctx context.Context, payload []byte, signature string) error {
	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if webhookSecret == "" {
		return errors.New("missing STRIPE_WEBHOOK_SECRET")
	}

	// 1. Verify Signature และแปลงเป็น Event ของ Stripe
	// 🔴 เปลี่ยนมาใช้ ConstructEventWithOptions เพื่อตั้งค่า IgnoreAPIVersionMismatch: true
	event, err := webhook.ConstructEventWithOptions(payload, signature, webhookSecret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if err != nil {
		return fmt.Errorf("invalid stripe signature: %w", err)
	}

	// 2. สนใจแค่ Event checkout.session.completed เท่านั้น
	if event.Type != "checkout.session.completed" {
		// Event อื่นๆ เราเพิกเฉย ไม่นับเป็น Error
		return nil
	}

	// 3. แปลง Raw Data กลับมาเป็น Checkout Session Object เพื่ออ่าน Metadata
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		return fmt.Errorf("failed to unmarshal checkout session: %w", err)
	}

	// 4. ดึงข้อมูลจาก Metadata
	bookingIDStr, ok1 := session.Metadata["booking_id"]
	seatIDStr, ok2 := session.Metadata["seat_id"]
	if !ok1 || !ok2 {
		return errors.New("missing booking_id or seat_id in metadata")
	}

	// 5. แปลง String เป็น uint
	bookingID, err1 := strconv.ParseUint(bookingIDStr, 10, 32)
	seatID, err2 := strconv.ParseUint(seatIDStr, 10, 32)
	if err1 != nil || err2 != nil {
		return errors.New("invalid booking_id or seat_id format in metadata")
	}

	// 6. ส่งไปอัปเดต Database ผ่าน Transaction พร้อมระบบ Idempotency
	return s.repo.ConfirmBookingTx(ctx, event.ID, uint(bookingID), uint(seatID))
}