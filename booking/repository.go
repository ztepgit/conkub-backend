package booking

import (
	"context"
	"errors"

	"conkub-backend/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository interface ประกาศ method ที่จำเป็นทั้งหมด รวมถึง ConfirmBooking
type Repository interface {
	BookSeatTx(ctx context.Context, userID string, eventID uint, seatID uint) error
	ConfirmBooking(ctx context.Context, seatID uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// BookSeatTx จัดการ Transaction และ Database Row Lock
func (r *repository) BookSeatTx(ctx context.Context, userID string, eventID uint, seatID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
		booking := models.Booking{
			UserID:  userID,
			EventID: eventID,
			SeatID:  seatID,
		}
		if err := tx.Create(&booking).Error; err != nil {
			return err
		}

		return nil
	})
}

// ConfirmBooking สำหรับ Webhook ใช้เปลี่ยนสถานะเมื่อจ่ายเงินสำเร็จ
func (r *repository) ConfirmBooking(ctx context.Context, seatID uint) error {
	// อัปเดตสถานะที่นั่งจาก PENDING เป็น BOOKED
	return r.db.WithContext(ctx).
		Model(&models.Seat{}).
		Where("id = ?", seatID).
		Update("status", models.SeatStatusBooked).Error
}