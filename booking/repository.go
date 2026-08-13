// booking/repository.go
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
	BookSeatTx(ctx context.Context, userID string, eventID uint, seatID uint) (*models.Booking, error)
	ConfirmBooking(ctx context.Context, seatID uint) error
	// 🔴 1. เพิ่มฟังก์ชัน CancelBooking สำหรับยกเลิกการจอง
	CancelBooking(ctx context.Context, bookingID uint, seatID uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}
// BookSeatTx จัดการ Transaction และ Database Row Lock
func (r *repository) BookSeatTx(ctx context.Context, userID string, eventID uint, seatID uint) (*models.Booking, error) {
	// 🔴 ประกาศตัวแปร booking ไว้ด้านนอก เพื่อให้ดึงค่าออกไปรีเทิร์นตอนจบได้
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
			Status:  models.BookingStatusPending, // 🔴 เพิ่มฟิลด์ Status เป็น PENDING เพื่อแก้ปัญหา NOT NULL constraint
		}

		// *หมายเหตุ: หากใน models/booking.go มีการประกาศ Constant ไว้ เช่น models.BookingStatusPending 
		// ให้เปลี่ยนคำว่า "PENDING" เป็น models.BookingStatusPending เพื่อความสม่ำเสมอของ Type แทนได้ครับ

		if err := tx.Create(&booking).Error; err != nil {
			return err
		}

		return nil
	})

	// 🔴 ถ้าระหว่าง Transaction มี Error ให้คืนค่า nil คู่กับ Error นั้น
	if err != nil {
		return nil, err
	}

	// 🔴 ถ้าสำเร็จ คืนค่า pointer ของ booking กลับไปให้ Service
	return &booking, nil
}

// ConfirmBooking สำหรับ Webhook ใช้เปลี่ยนสถานะเมื่อจ่ายเงินสำเร็จ
func (r *repository) ConfirmBooking(ctx context.Context, seatID uint) error {
	// อัปเดตสถานะที่นั่งจาก PENDING เป็น BOOKED
	return r.db.WithContext(ctx).
		Model(&models.Seat{}).
		Where("id = ?", seatID).
		Update("status", models.SeatStatusBooked).Error
}

// 🔴 2. Implement ฟังก์ชัน CancelBooking สำหรับลบ Booking ออกและคืนที่นั่ง
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