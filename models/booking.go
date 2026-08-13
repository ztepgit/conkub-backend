// models/booking.go
package models

import "time"
// สร้าง Custom Type สำหรับ Booking
type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "PENDING"
	BookingStatusConfirmed BookingStatus = "CONFIRMED"
	BookingStatusCancelled BookingStatus = "CANCELLED"
)

type Booking struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    string    `gorm:"type:uuid;index;not null"` // เก็บ UUID ของ User จาก Supabase
	EventID   uint      `gorm:"not null"`
	SeatID    uint      `gorm:"uniqueIndex;not null"` // 1 ที่นั่ง โดนจองได้แค่ 1 ครั้ง
	Seat      Seat      `gorm:"foreignKey:SeatID"`
	Status    BookingStatus `gorm:"column:status;not null;default:'PENDING'" json:"status"`
	CreatedAt time.Time
}