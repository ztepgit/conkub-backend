package models

import "time"

type StripeWebhookEvent struct {
	ID            uint      `gorm:"primaryKey"`
	StripeEventID string    `gorm:"uniqueIndex;not null"` // Used for Idempotency
	EventType     string    `gorm:"not null"`
	ProcessedAt   time.Time `gorm:"autoCreateTime"`
}
