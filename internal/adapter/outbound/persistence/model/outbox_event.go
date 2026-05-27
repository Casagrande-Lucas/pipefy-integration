package model

import "time"

// OutboxEvent is the GORM model for the outbox_events table.
type OutboxEvent struct {
	ID          string     `gorm:"primaryKey;type:uuid"`
	EventType   string     `gorm:"not null"`
	Payload     string     `gorm:"not null;type:text"`
	Status      string     `gorm:"not null;index"`
	Attempts    int        `gorm:"not null;default:0"`
	MaxAttempts int        `gorm:"not null;default:3"`
	LastError   string     `gorm:"type:text"`
	CreatedAt   time.Time  `gorm:"not null;index"`
	ProcessedAt *time.Time `gorm:"index"`
}
