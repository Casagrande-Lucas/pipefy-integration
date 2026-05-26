package service

import "github.com/Casagrande-Lucas/pipefy-integration/internal/domain/entity"

// PipefyService defines the outbound port for Pipefy card operations.
// Implementations must structure the exact GraphQL mutations as specified
// in the Pipefy API documentation (https://developers.pipefy.com/reference/cards).
type PipefyService interface {
	// CreateCard sends (or simulates) the createCard GraphQL mutation.
	// Returns the card ID assigned by Pipefy, which must be stored on the
	// client entity (PipefyCardID) for future UpdateCard calls.
	CreateCard(client *entity.Client) (string, error)

	// UpdateCard sends (or simulates) two updateCardField GraphQL mutations:
	// one for the status field and one for the priority field.
	UpdateCard(client *entity.Client) error
}
