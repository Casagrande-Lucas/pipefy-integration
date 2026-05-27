package valueobject

import "fmt"

// OutboxEventStatus represents the processing lifecycle of an outbox event.
type OutboxEventStatus string

const (
	// OutboxStatusPending is the initial status — event is waiting to be processed.
	OutboxStatusPending OutboxEventStatus = "pending"

	// OutboxStatusProcessing means a worker has picked up the event and is executing it.
	OutboxStatusProcessing OutboxEventStatus = "processing"

	// OutboxStatusDone means the event was successfully delivered to Pipefy.
	OutboxStatusDone OutboxEventStatus = "done"

	// OutboxStatusFailed means the event exceeded max retry attempts.
	OutboxStatusFailed OutboxEventStatus = "failed"
)

// NewOutboxEventStatus parses and validates an outbox event status string.
func NewOutboxEventStatus(value string) (OutboxEventStatus, error) {
	s := OutboxEventStatus(value)
	if !s.IsValid() {
		return "", fmt.Errorf("outbox event status %q is invalid", value)
	}
	return s, nil
}

// IsValid reports whether the status is a recognized value.
func (s OutboxEventStatus) IsValid() bool {
	switch s {
	case OutboxStatusPending, OutboxStatusProcessing, OutboxStatusDone, OutboxStatusFailed:
		return true
	}
	return false
}

// String returns the status as a plain string.
func (s OutboxEventStatus) String() string {
	return string(s)
}
