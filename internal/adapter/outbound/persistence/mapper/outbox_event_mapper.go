package mapper

import (
	"fmt"

	"github.com/Casagrande-Lucas/pipefy-integration/internal/adapter/outbound/persistence/model"
	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/entity"
	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/valueobject"
)

// OutboxEventToModel converts a domain OutboxEvent into its GORM model.
func OutboxEventToModel(e *entity.OutboxEvent) *model.OutboxEvent {
	return &model.OutboxEvent{
		ID:          e.ID.String(),
		EventType:   string(e.EventType),
		Payload:     e.Payload,
		Status:      e.Status.String(),
		Attempts:    e.Attempts,
		MaxAttempts: e.MaxAttempts,
		LastError:   e.LastError,
		CreatedAt:   e.CreatedAt,
		ProcessedAt: e.ProcessedAt,
	}
}

// OutboxEventToEntity converts a GORM OutboxEvent model into the domain entity.
func OutboxEventToEntity(m *model.OutboxEvent) (*entity.OutboxEvent, error) {
	id, err := valueobject.ParseID(m.ID)
	if err != nil {
		return nil, fmt.Errorf("outbox_event mapper: invalid id %q: %w", m.ID, err)
	}

	status, err := valueobject.NewOutboxEventStatus(m.Status)
	if err != nil {
		return nil, fmt.Errorf("outbox_event mapper: invalid status %q: %w", m.Status, err)
	}

	return &entity.OutboxEvent{
		ID:          id,
		EventType:   entity.OutboxEventType(m.EventType),
		Payload:     m.Payload,
		Status:      status,
		Attempts:    m.Attempts,
		MaxAttempts: m.MaxAttempts,
		LastError:   m.LastError,
		CreatedAt:   m.CreatedAt,
		ProcessedAt: m.ProcessedAt,
	}, nil
}
