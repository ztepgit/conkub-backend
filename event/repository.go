// event/repository.go
package event

import (
	"context"
	"conkub-backend/models"
	"gorm.io/gorm"
)

// 🔴 1. สร้าง Struct พิเศษสำหรับรับค่าที่มี Count มาด้วย
type EventWithTicketCount struct {
	models.Event
	RemainingTickets int64   `gorm:"column:remaining_tickets"`
	Price            float64 `gorm:"column:price"` // 🔴 เพิ่มฟิลด์รับราคาจาก Subquery
}

type Repository interface {
	FindAll(ctx context.Context) ([]EventWithTicketCount, error)
	FindByID(ctx context.Context, id uint) (*EventWithTicketCount, error)
	FindSeatsByEventID(ctx context.Context, eventID uint) ([]models.Seat, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context) ([]EventWithTicketCount, error) {
	var events []EventWithTicketCount

	// 🔴 ใช้ SQL Subquery นับตั๋วที่ AVAILABLE และดึงราคา 1 ค่าจากตาราง seats
	err := r.db.WithContext(ctx).Model(&models.Event{}).
		Select(`
			events.*, 
			(SELECT COUNT(*) FROM seats WHERE seats.event_id = events.id AND seats.status = 'AVAILABLE') AS remaining_tickets,
			(SELECT price FROM seats WHERE seats.event_id = events.id LIMIT 1) AS price
		`).
		Find(&events).Error

	return events, err
}

func (r *repository) FindByID(ctx context.Context, id uint) (*EventWithTicketCount, error) {
	var event EventWithTicketCount

	err := r.db.WithContext(ctx).Model(&models.Event{}).
		Select(`
			events.*, 
			(SELECT COUNT(*) FROM seats WHERE seats.event_id = events.id AND seats.status = 'AVAILABLE') AS remaining_tickets,
			(SELECT price FROM seats WHERE seats.event_id = events.id LIMIT 1) AS price
		`).
		Where("events.id = ?", id).
		First(&event).Error

	if err != nil {
		return nil, err
	}

	return &event, nil
}

// คงฟังก์ชันนี้ไว้ 100% ไม่ดัดแปลง
func (r *repository) FindSeatsByEventID(ctx context.Context, eventID uint) ([]models.Seat, error) {
	var seats []models.Seat
	err := r.db.WithContext(ctx).Where("event_id = ?", eventID).Find(&seats).Error
	return seats, err
}