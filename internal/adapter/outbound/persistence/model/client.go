package model

import "time"

// Client is the GORM persistence model for the Client aggregate.
// It is an infrastructure concern and must not be used outside the persistence adapter.
type Client struct {
	ID             string `gorm:"primaryKey;type:uuid"`
	Name           string `gorm:"not null"`
	Email          string `gorm:"uniqueIndex;not null"`
	RequestType    string `gorm:"not null"`
	PatrimonyValue float64 `gorm:"not null"`
	Priority       string `gorm:"not null"`
	Status         string `gorm:"not null"`
	PipefyCardID   string
	CreatedAt      time.Time `gorm:"not null"`
	UpdatedAt      time.Time `gorm:"not null"`
}
