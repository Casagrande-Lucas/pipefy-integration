package entity

import "time"

// ProcessedEvent records a webhook event that has already been handled.
// It exists to enforce the idempotency business rule: the same event_id
// must never trigger processing more than once.
//
// ProcessedEvent is an Entity: it is identified by EventID and persists
// in the system for as long as idempotency must be guaranteed.
type ProcessedEvent struct {
	EventID     string
	CardID      string
	ClientEmail string
	ProcessedAt time.Time
}

// NewProcessedEvent constructs a ProcessedEvent stamped with the current UTC time.
func NewProcessedEvent(eventID, cardID, clientEmail string) *ProcessedEvent {
	return &ProcessedEvent{
		EventID:     eventID,
		CardID:      cardID,
		ClientEmail: clientEmail,
		ProcessedAt: time.Now().UTC(),
	}
}
