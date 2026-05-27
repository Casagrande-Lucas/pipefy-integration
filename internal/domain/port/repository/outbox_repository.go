package repository

import "github.com/Casagrande-Lucas/pipefy-integration/internal/domain/entity"

// OutboxRepository defines persistence operations for the outbox pattern.
type OutboxRepository interface {
	// Save persists a new outbox event within the current transaction.
	Save(event *entity.OutboxEvent) error

	// FindPending returns up to limit events with status pending, ordered by created_at asc.
	FindPending(limit int) ([]*entity.OutboxEvent, error)

	// UpdateStatus persists the current status, attempts, last_error and processed_at of the event.
	UpdateStatus(event *entity.OutboxEvent) error
}
