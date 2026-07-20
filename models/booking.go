// models/booking.go
package models

import "time"

type Booking struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    string    `gorm:"type:uuid;index;not null"` // เก็บ UUID ของ User จาก Supabase
	EventID   uint      `gorm:"not null"`
	SeatID    uint      `gorm:"uniqueIndex;not null"` // 1 ที่นั่ง โดนจองได้แค่ 1 ครั้ง
	Seat      Seat      `gorm:"foreignKey:SeatID"`
	CreatedAt time.Time
}