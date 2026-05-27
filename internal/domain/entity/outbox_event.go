package entity

import (
	"time"

	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/valueobject"
)

const defaultMaxAttempts = 3

// OutboxEventType identifies which Pipefy operation the event represents.
type OutboxEventType string

const (
	EventTypeCreateCard OutboxEventType = "create_card"
	EventTypeUpdateCard OutboxEventType = "update_card"
)

// OutboxEvent records a pending Pipefy operation to be delivered by the background worker.
// It is an Entity: identified by ID, with a mutable lifecycle (status and retry tracking).
type OutboxEvent struct {
	ID          valueobject.ID
	EventType   OutboxEventType
	Payload     string
	Status      valueobject.OutboxEventStatus
	Attempts    int
	MaxAttempts int
	LastError   string
	CreatedAt   time.Time
	ProcessedAt *time.Time
}

// NewOutboxEvent constructs a pending OutboxEvent ready to be persisted.
func NewOutboxEvent(eventType OutboxEventType, payload string) *OutboxEvent {
	return &OutboxEvent{
		ID:          valueobject.NewID(),
		EventType:   eventType,
		Payload:     payload,
		Status:      valueobject.OutboxStatusPending,
		Attempts:    0,
		MaxAttempts: defaultMaxAttempts,
		CreatedAt:   time.Now().UTC(),
	}
}

// CanRetry reports whether the event has remaining attempts.
func (e *OutboxEvent) CanRetry() bool {
	return e.Attempts < e.MaxAttempts
}

// MarkProcessing transitions the event to the processing state.
func (e *OutboxEvent) MarkProcessing() {
	e.Status = valueobject.OutboxStatusProcessing
	e.Attempts++
}

// MarkDone transitions the event to done and stamps the processed time.
func (e *OutboxEvent) MarkDone() {
	now := time.Now().UTC()
	e.Status = valueobject.OutboxStatusDone
	e.ProcessedAt = &now
	e.LastError = ""
}

// MarkFailed records the error and transitions to pending (retry) or failed (exhausted).
func (e *OutboxEvent) MarkFailed(err error) {
	e.LastError = err.Error()
	if e.CanRetry() {
		e.Status = valueobject.OutboxStatusPending
	} else {
		e.Status = valueobject.OutboxStatusFailed
	}
}
