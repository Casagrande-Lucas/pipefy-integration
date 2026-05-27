package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/Casagrande-Lucas/pipefy-integration/internal/adapter/outbound/persistence/mapper"
	"github.com/Casagrande-Lucas/pipefy-integration/internal/adapter/outbound/persistence/model"
	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/entity"
	"github.com/Casagrande-Lucas/pipefy-integration/pkg/apperror"
)

// processedEventRepository implements port/repository.ProcessedEventRepository using GORM.
type processedEventRepository struct {
	db *gorm.DB
}

// NewProcessedEventRepository returns a ProcessedEventRepository backed by PostgreSQL.
func NewProcessedEventRepository(db *gorm.DB) *processedEventRepository {
	return &processedEventRepository{db: db}
}

func (r *processedEventRepository) Exists(eventID string) (bool, error) {
	var m model.ProcessedEvent
	err := r.db.Where("event_id = ?", eventID).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, apperror.NewInternal(err)
	}
	return true, nil
}

func (r *processedEventRepository) Save(event *entity.ProcessedEvent) error {
	m := mapper.ProcessedEventToModel(event)
	if err := r.db.Create(m).Error; err != nil {
		return apperror.NewInternal(err)
	}
	return nil
}
