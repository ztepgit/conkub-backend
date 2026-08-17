// booking/repository.go
package booking

import (
	"context"
	"errors"
	"fmt" // 🔴 เพิ่ม fmt สำหรับใช้คืนค่า error ด้วย fmt.Errorf

	"conkub-backend/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository interface ประกาศ method ที่จำเป็นทั้งหมด
type Repository interface {
	BookSeatTx(ctx context.Context, userID string, eventID uint, seatID uint) (*models.Booking, error)
	ConfirmBooking(ctx context.Context, seatID uint) error
	CancelBooking(ctx context.Context, bookingID uint, seatID uint) error
	// 🔴 เพิ่มฟังก์ชัน ConfirmBookingTx แบบมี Idempotency และเช็ค State
	ConfirmBookingTx(ctx context.Context, stripeEventID string, bookingID uint, seatID uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// BookSeatTx จัดการ Transaction และ Database Row Lock
func (r *repository) BookSeatTx(ctx context.Context, userID string, eventID uint, seatID uint) (*models.Booking, error) {
	// ประกาศตัวแปร booking ไว้ด้านนอก เพื่อให้ดึงค่าออกไปรีเทิร์นตอนจบได้
	var booking models.Booking

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var seat models.Seat

		// 1. SELECT ... FOR UPDATE (ล็อค Row นี้ไว้จนกว่า Transaction จะจบ)
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&seat, seatID).Error; err != nil {
			return errors.New("seat not found")
		}

		// 2. เช็คว่าที่นั่งถูกจอง หรือ กำลังรอจ่ายเงินอยู่หรือไม่
		if seat.Status == models.SeatStatusBooked || seat.Status == models.SeatStatusPending {
			return errors.New("seat already booked or pending payment")
		}

		// 3. อัปเดตสถานะเป็น PENDING (รอชำระเงินจาก Stripe)
		seat.Status = models.SeatStatusPending
		if err := tx.Save(&seat).Error; err != nil {
			return err
		}

		// 4. บันทึกประวัติการจองลงตาราง bookings
		booking = models.Booking{
			UserID:  userID,
			EventID: eventID,
			SeatID:  seatID,
			Status:  models.BookingStatusPending,
		}

		if err := tx.Create(&booking).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &booking, nil
}

// ConfirmBooking สำหรับ Webhook ใช้เปลี่ยนสถานะเมื่อจ่ายเงินสำเร็จ (ฟังก์ชันเดิม)
func (r *repository) ConfirmBooking(ctx context.Context, seatID uint) error {
	// อัปเดตสถานะที่นั่งจาก PENDING เป็น BOOKED
	return r.db.WithContext(ctx).
		Model(&models.Seat{}).
		Where("id = ?", seatID).
		Update("status", models.SeatStatusBooked).Error
}

// CancelBooking สำหรับลบ Booking ออกและคืนที่นั่ง
func (r *repository) CancelBooking(ctx context.Context, bookingID uint, seatID uint) error {
	// เปิด Transaction สำหรับการลบและการอัปเดตให้เป็น Atomic
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// ลบ Booking ขยะที่สร้างไว้ทิ้ง (ใช้ Unscoped เพื่อ Hard Delete ป้องกันข้อมูลตกค้าง)
	if err := tx.Unscoped().Where("id = ?", bookingID).Delete(&models.Booking{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// อัปเดต Seat กลับเป็น AVAILABLE (กำหนดเงื่อนไข AND status = PENDING เพื่อความปลอดภัย)
	result := tx.Model(&models.Seat{}).
		Where("id = ? AND status = ?", seatID, models.SeatStatusPending).
		Update("status", models.SeatStatusAvailable)

	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	return tx.Commit().Error
}

// 🔴 Implement ฟังก์ชัน ConfirmBookingTx
func (r *repository) ConfirmBookingTx(ctx context.Context, stripeEventID string, bookingID uint, seatID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Idempotency Check: เคยทำ Event นี้ไปหรือยัง?
		var webhookEvent models.StripeWebhookEvent
		err := tx.Where("stripe_event_id = ?", stripeEventID).First(&webhookEvent).Error
		if err == nil {
			// ทำไปแล้ว ให้ถือว่าสำเร็จ (Idempotent)
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("database error checking idempotency: %w", err)
		}

		// 2. Lock the Booking row
		var booking models.Booking
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&booking, bookingID).Error; err != nil {
			return fmt.Errorf("booking not found: %w", err)
		}

		// 🔴 3. Verify SeatID จาก Metadata
		if booking.SeatID != seatID {
			return fmt.Errorf("metadata seat_id %d does not match booking seat_id %d", seatID, booking.SeatID)
		}

		// 🔴 4. State Machine Validation
		if booking.Status == models.BookingStatusConfirmed {
			// ถ้าชำระแล้ว (CONFIRMED) ให้ข้ามไปและบันทึก Event ไว้ได้เลย (Idempotent)
			webhookEvent = models.StripeWebhookEvent{StripeEventID: stripeEventID, EventType: "checkout.session.completed"}
			return tx.Create(&webhookEvent).Error
		} else if booking.Status != models.BookingStatusPending {
			// ถ้าถูก Cancelled หรือ Expired ไปแล้ว ห้าม Confirm เด็ดขาด
			return fmt.Errorf("cannot confirm booking with status: %s", booking.Status)
		}

		// 5. Lock the Seat row associated with this booking
		var seat models.Seat
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&seat, booking.SeatID).Error; err != nil {
			return fmt.Errorf("seat not found: %w", err)
		}

		// 6. Update Statuses
		booking.Status = models.BookingStatusConfirmed
		if err := tx.Save(&booking).Error; err != nil {
			return fmt.Errorf("failed to update booking: %w", err)
		}

		seat.Status = models.SeatStatusBooked
		if err := tx.Save(&seat).Error; err != nil {
			return fmt.Errorf("failed to update seat: %w", err)
		}

		// 7. Save Idempotency Record
		webhookEvent = models.StripeWebhookEvent{
			StripeEventID: stripeEventID,
			EventType:     "checkout.session.completed",
		}
		if err := tx.Create(&webhookEvent).Error; err != nil {
			return fmt.Errorf("failed to save webhook event: %w", err)
		}

		return nil
	})
}