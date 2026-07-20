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
	// 🔴 2. ปรับ Type ที่ Return ให้เป็น EventWithTicketCount
	FindAll(ctx context.Context) ([]EventWithTicketCount, error)
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
	
	// 🔴 3. ใช้ SQL Subquery นับตั๋วที่ AVAILABLE พร้อมกับ WithContext ของคุณ
	err := r.db.WithContext(ctx).
		Table("events").
		Select("events.*, (SELECT COUNT(*) FROM seats WHERE seats.event_id = events.id AND seats.status = 'AVAILABLE') AS remaining_tickets").
		Scan(&events).Error
		
	return events, err
}

// คงฟังก์ชันนี้ของคุณไว้ 100% ไม่ดัดแปลง
func (r *repository) FindSeatsByEventID(ctx context.Context, eventID uint) ([]models.Seat, error) {
	var seats []models.Seat
	err := r.db.WithContext(ctx).Where("event_id = ?", eventID).Find(&seats).Error
	return seats, err
}