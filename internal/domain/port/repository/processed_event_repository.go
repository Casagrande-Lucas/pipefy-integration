package repository

import "github.com/Casagrande-Lucas/pipefy-integration/internal/domain/entity"

// ProcessedEventRepository defines persistence operations for webhook idempotency.
type ProcessedEventRepository interface {
	// Exists reports whether the given event_id has already been processed.
	Exists(eventID string) (bool, error)

	// Save records the event so future duplicates are rejected.
	Save(event *entity.ProcessedEvent) error
}
