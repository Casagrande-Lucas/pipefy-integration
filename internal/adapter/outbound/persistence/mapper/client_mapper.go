package mapper

import (
	"github.com/Casagrande-Lucas/pipefy-integration/internal/adapter/outbound/persistence/model"
	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/entity"
	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/valueobject"
)

// ClientToModel converts a domain Client entity to its GORM persistence model.
func ClientToModel(c *entity.Client) *model.Client {
	return &model.Client{
		ID:             c.ID.String(),
		Name:           c.Name,
		Email:          c.Email.String(),
		RequestType:    c.RequestType,
		PatrimonyValue: c.PatrimonyValue,
		Priority:       c.Priority.String(),
		Status:         c.Status.String(),
		PipefyCardID:   c.PipefyCardID,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}

// ClientToEntity converts a GORM Client model to a domain entity.
// Returns an error if any persisted value fails value object validation.
func ClientToEntity(m *model.Client) (*entity.Client, error) {
	id, err := valueobject.ParseID(m.ID)
	if err != nil {
		return nil, err
	}

	email, err := valueobject.NewEmail(m.Email)
	if err != nil {
		return nil, err
	}

	status, err := valueobject.NewStatus(m.Status)
	if err != nil {
		return nil, err
	}

	return &entity.Client{
		ID:             id,
		Name:           m.Name,
		Email:          email,
		RequestType:    m.RequestType,
		PatrimonyValue: m.PatrimonyValue,
		Priority:       valueobject.Priority(m.Priority),
		Status:         status,
		PipefyCardID:   m.PipefyCardID,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}, nil
}
