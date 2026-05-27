package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/Casagrande-Lucas/pipefy-integration/internal/adapter/outbound/persistence/mapper"
	"github.com/Casagrande-Lucas/pipefy-integration/internal/adapter/outbound/persistence/model"
	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/entity"
	"github.com/Casagrande-Lucas/pipefy-integration/pkg/apperror"
)

// clientRepository implements port/repository.ClientRepository using GORM.
type clientRepository struct {
	db *gorm.DB
}

// NewClientRepository returns a ClientRepository backed by PostgreSQL.
func NewClientRepository(db *gorm.DB) *clientRepository {
	return &clientRepository{db: db}
}

func (r *clientRepository) Save(client *entity.Client) error {
	m := mapper.ClientToModel(client)
	if err := r.db.Create(m).Error; err != nil {
		return apperror.NewInternal(err)
	}
	return nil
}

func (r *clientRepository) FindByEmail(email string) (*entity.Client, error) {
	var m model.Client
	err := r.db.Where("email = ?", email).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFound("client", err)
		}
		return nil, apperror.NewInternal(err)
	}

	client, err := mapper.ClientToEntity(&m)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	return client, nil
}

func (r *clientRepository) FindByPipefyCardID(cardID string) (*entity.Client, error) {
	var m model.Client
	err := r.db.Where("pipefy_card_id = ?", cardID).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFound("client", err)
		}
		return nil, apperror.NewInternal(err)
	}

	client, err := mapper.ClientToEntity(&m)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	return client, nil
}

// UpdatePipefyCardID updates only the pipefy_card_id field after the outbox worker
// successfully delivers the createCard mutation to Pipefy.
func (r *clientRepository) UpdatePipefyCardID(client *entity.Client) error {
	err := r.db.Model(&model.Client{}).
		Where("id = ?", client.ID.String()).
		Updates(map[string]interface{}{
			"pipefy_card_id": client.PipefyCardID,
			"updated_at":     client.UpdatedAt,
		}).Error
	if err != nil {
		return apperror.NewInternal(err)
	}
	return nil
}

func (r *clientRepository) UpdateStatusAndPriority(client *entity.Client) error {
	err := r.db.Model(&model.Client{}).
		Where("id = ?", client.ID.String()).
		Updates(map[string]interface{}{
			"status":     client.Status.String(),
			"priority":   client.Priority.String(),
			"updated_at": client.UpdatedAt,
		}).Error
	if err != nil {
		return apperror.NewInternal(err)
	}
	return nil
}
