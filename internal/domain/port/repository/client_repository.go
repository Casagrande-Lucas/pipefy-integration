package repository

import "github.com/Casagrande-Lucas/pipefy-integration/internal/domain/entity"

// ClientRepository defines persistence operations for the Client aggregate.
type ClientRepository interface {
	// Save persists a new client and sets the generated ID on the entity.
	Save(client *entity.Client) error

	// FindByEmail returns the client with the given email, or apperror.NotFound if absent.
	FindByEmail(email string) (*entity.Client, error)

	// UpdateStatusAndPriority updates only the status, priority, and updated_at fields.
	UpdateStatusAndPriority(client *entity.Client) error

	// UpdatePipefyCardID updates only the pipefy_card_id field after the outbox worker
	// successfully delivers the createCard mutation to Pipefy.
	UpdatePipefyCardID(client *entity.Client) error

	// FindByPipefyCardID returns the client with the given pipefy_card_id,
	// or apperror.NotFound if absent.
	FindByPipefyCardID(cardID string) (*entity.Client, error)
}
