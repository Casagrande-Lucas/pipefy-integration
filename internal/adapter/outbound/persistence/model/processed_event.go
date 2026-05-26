package model

import "time"

// ProcessedEvent is the GORM persistence model for the ProcessedEvent entity.
// EventID is the natural primary key — it is the webhook event_id from the payload.
// It is an infrastructure concern and must not be used outside the persistence adapter.
type ProcessedEvent struct {
	EventID     string    `gorm:"primaryKey"`
	CardID      string    `gorm:"not null"`
	ClientEmail string    `gorm:"index;not null"`
	ProcessedAt time.Time `gorm:"not null"`
}
