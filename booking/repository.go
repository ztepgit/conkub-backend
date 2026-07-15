package booking

import (
	"context"
	"errors"

	"conkub-backend/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	BookSeatTx(ctx context.Context, userID string, eventID uint, seatID uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// BookSeatTx จัดการ Transaction และ Database Row Lock
func (r *repository) BookSeatTx(ctx context.Context, userID string, eventID uint, seatID uint) error {
	// เริ่ม Database Transaction
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var seat models.Seat

		// 1. SELECT ... FOR UPDATE (ล็อค Row นี้ไว้จนกว่า Transaction จะจบ)
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&seat, seatID).Error; err != nil {
			return errors.New("seat not found")
		}

		// 2. เช็คว่าที่นั่งถูกจองไปหรือยัง (กันหลุด)
		if seat.Status == models.SeatStatusBooked {
			return errors.New("seat already booked")
		}

		// 3. อัปเดตสถานะที่นั่งเป็น BOOKED
		seat.Status = models.SeatStatusBooked
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

		// Return nil หมายถึง Commit Transaction สำเร็จ!
		return nil
	})
}