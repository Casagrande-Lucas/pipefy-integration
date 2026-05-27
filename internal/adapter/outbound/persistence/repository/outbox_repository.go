package repository

import (
	"gorm.io/gorm"

	"github.com/Casagrande-Lucas/pipefy-integration/internal/adapter/outbound/persistence/mapper"
	"github.com/Casagrande-Lucas/pipefy-integration/internal/adapter/outbound/persistence/model"
	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/entity"
	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/valueobject"
	"github.com/Casagrande-Lucas/pipefy-integration/pkg/apperror"
)

// outboxRepository implements port/repository.OutboxRepository using GORM.
type outboxRepository struct {
	db *gorm.DB
}

// NewOutboxRepository returns an OutboxRepository backed by PostgreSQL.
func NewOutboxRepository(db *gorm.DB) *outboxRepository {
	return &outboxRepository{db: db}
}

// Save persists a new outbox event within the current transaction.
func (r *outboxRepository) Save(event *entity.OutboxEvent) error {
	m := mapper.OutboxEventToModel(event)
	if err := r.db.Create(m).Error; err != nil {
		return apperror.NewInternal(err)
	}
	return nil
}

// FindPending returns up to limit events with status pending, ordered by created_at asc.
func (r *outboxRepository) FindPending(limit int) ([]*entity.OutboxEvent, error) {
	var models []model.OutboxEvent
	err := r.db.
		Where("status = ?", valueobject.OutboxStatusPending.String()).
		Order("created_at asc").
		Limit(limit).
		Find(&models).Error
	if err != nil {
		return nil, apperror.NewInternal(err)
	}

	events := make([]*entity.OutboxEvent, 0, len(models))
	for i := range models {
		e, err := mapper.OutboxEventToEntity(&models[i])
		if err != nil {
			return nil, apperror.NewInternal(err)
		}
		events = append(events, e)
	}
	return events, nil
}

// UpdateStatus persists the current status, attempts, last_error and processed_at of the event.
func (r *outboxRepository) UpdateStatus(event *entity.OutboxEvent) error {
	err := r.db.Model(&model.OutboxEvent{}).
		Where("id = ?", event.ID.String()).
		Updates(map[string]interface{}{
			"status":       event.Status.String(),
			"attempts":     event.Attempts,
			"last_error":   event.LastError,
			"processed_at": event.ProcessedAt,
		}).Error
	if err != nil {
		return apperror.NewInternal(err)
	}
	return nil
}
