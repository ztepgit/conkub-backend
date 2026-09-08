// event/repository.go
package event

import (
	"context"
	"conkub-backend/models"
	"time"

	"gorm.io/gorm"
)

// 🔴 1. สร้าง Struct พิเศษสำหรับรับค่าที่มี Count มาด้วย
type EventWithTicketCount struct {
	models.Event
	RemainingTickets int64   `gorm:"column:remaining_tickets"`
	Price            float64 `gorm:"column:price"` // 🔴 เพิ่มฟิลด์รับราคาจาก Subquery
}

type Repository interface {
	FindAll(ctx context.Context, search, location string, parsedDate *time.Time) ([]EventWithTicketCount, error)
	FindByID(ctx context.Context, id uint) (*EventWithTicketCount, error)
	FindSeatsByEventID(ctx context.Context, eventID uint) ([]models.Seat, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context, search, location string, parsedDate *time.Time) ([]EventWithTicketCount, error) {
	var events []EventWithTicketCount

	// เริ่มสร้าง Query พื้นฐานที่คง Logic เดิมเอาไว้ 100%
	query := r.db.WithContext(ctx).Model(&models.Event{}).
		Select(`
			events.*, 
			(SELECT COUNT(*) FROM seats WHERE seats.event_id = events.id AND seats.status = 'AVAILABLE') AS remaining_tickets,
			(SELECT price FROM seats WHERE seats.event_id = events.id LIMIT 1) AS price
		`)

	// เช็คเงื่อนไข Search (ใช้ ILIKE เพื่อไม่สนใจตัวพิมพ์เล็ก-ใหญ่)
	if search != "" {
		query = query.Where("events.artist ILIKE ? OR events.name ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// เช็คเงื่อนไข Location
	if location != "" {
		query = query.Where("events.venue ILIKE ?", "%"+location+"%")
	}

	// เช็คเงื่อนไข Date (ทำ Range ค้นหาตั้งแต่เริ่มวัน จนถึงก่อนเริ่มวันถัดไป)
	if parsedDate != nil {
		startOfDay := *parsedDate
		startOfNextDay := startOfDay.AddDate(0, 0, 1)
		query = query.Where("events.show_time >= ? AND events.show_time < ?", startOfDay, startOfNextDay)
	}

	// 🔴 รัน Query พร้อม Subquery นับตั๋วที่ AVAILABLE และดึงราคา
	err := query.Find(&events).Error

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