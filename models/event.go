// models/event.go
package models

import "time"

type Event struct {
	ID          uint      `gorm:"primaryKey"`
	Name        string    `gorm:"type:varchar(255);not null"`
	Artist      string    `gorm:"type:varchar(255);not null"` //  เพิ่มฟิลด์ Artist
	Description string    `gorm:"type:text"`
	Venue       string    `gorm:"type:varchar(255);not null"`
	Category    string    `gorm:"type:varchar(100)"`          //  เพิ่มฟิลด์ Category
	ImageURL    string    `gorm:"type:text"`                  //  เพิ่มฟิลด์ ImageURL
	ShowTime    time.Time `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}