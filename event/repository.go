package event

import (
	"context"
	"conkub-backend/models"
	"gorm.io/gorm"
)

// 🔴 1. สร้าง Struct พิเศษสำหรับรับค่าที่มี Count มาด้วย
type EventWithTicketCount struct {
	models.Event
	RemainingTickets int64 `gorm:"column:remaining_tickets"`
}

type Repository interface {
	FindAll(ctx context.Context) ([]EventWithTicketCount, error)
	// 🔴 2. เพิ่มฟังก์ชันสำหรับดึง Event รายตัวตาม ID (รองรับ GET /api/v1/events/:id)
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

	// ใช้ SQL Subquery นับตั๋วที่ AVAILABLE
	err := r.db.WithContext(ctx).
		Table("events").
		Select("events.*, (SELECT COUNT(*) FROM seats WHERE seats.event_id = events.id AND seats.status = 'AVAILABLE') AS remaining_tickets").
		Scan(&events).Error

	return events, err
}

// 🔴 3. เพิ่ม Implementation ของ FindByID โดยใช้ Subquery นับตั๋วเช่นเดียวกับ FindAll
// เมื่อไม่พบข้อมูล GORM จะคืนค่า gorm.ErrRecordNotFound ซึ่ง Handler จะนำไปแปลงเป็น 404
func (r *repository) FindByID(ctx context.Context, id uint) (*EventWithTicketCount, error) {
	var event EventWithTicketCount

	err := r.db.WithContext(ctx).
		Table("events").
		Select("events.*, (SELECT COUNT(*) FROM seats WHERE seats.event_id = events.id AND seats.status = 'AVAILABLE') AS remaining_tickets").
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