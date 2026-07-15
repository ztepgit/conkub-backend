package event

import (
	"context"
	"conkub-backend/models"
	"gorm.io/gorm"
)

type Repository interface {
	FindAll(ctx context.Context) ([]models.Event, error)
	FindSeatsByEventID(ctx context.Context, eventID uint) ([]models.Seat, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context) ([]models.Event, error) {
	var events []models.Event
	// ใช้ WithContext เพื่อให้รองรับ Context Timeout
	err := r.db.WithContext(ctx).Find(&events).Error
	return events, err
}

func (r *repository) FindSeatsByEventID(ctx context.Context, eventID uint) ([]models.Seat, error) {
	var seats []models.Seat
	err := r.db.WithContext(ctx).Where("event_id = ?", eventID).Find(&seats).Error
	return seats, err
}