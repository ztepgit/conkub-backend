package models

import "time"

type Event struct {
	ID          uint      `gorm:"primaryKey"`
	Name        string    `gorm:"type:varchar(255);not null"`
	Description string    `gorm:"type:text"`
	Venue       string    `gorm:"type:varchar(255);not null"`
	ShowTime    time.Time `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}