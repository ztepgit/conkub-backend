package models

import "time"

type SeatStatus string

const (
	SeatStatusAvailable SeatStatus = "AVAILABLE"
	SeatStatusPending   SeatStatus = "PENDING" // เพิ่มสถานะกำลังรอจ่ายเงิน
	SeatStatusBooked    SeatStatus = "BOOKED"
)

type Seat struct {
	ID        uint       `gorm:"primaryKey"`
	EventID   uint       `gorm:"index;not null"`
	Event     Event      `gorm:"foreignKey:EventID"` // Relation
	Row       string     `gorm:"type:varchar(10);not null"`
	Number    int        `gorm:"not null"`
	Price     float64    `gorm:"not null"`
	Status    SeatStatus `gorm:"type:varchar(20);default:'AVAILABLE';not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}