package client

import (
	"encoding/json"
	"fmt"

	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/entity"
	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/port/repository"
	"github.com/Casagrande-Lucas/pipefy-integration/pkg/apperror"
)

// CreateClientInput holds the raw data required to create a client.
type CreateClientInput struct {
	Name           string
	Email          string
	RequestType    string
	PatrimonyValue float64
}

// CreateClientRepos declares exactly which repositories the CreateClient use
// case needs inside the transaction. No assumptions are made beyond what this
// use case actually requires.
type CreateClientRepos struct {
	Client repository.ClientRepository
	Outbox repository.OutboxRepository
}

// createCardPayload is the self-contained event body stored in the outbox.
// It carries enough data for the worker to call Pipefy without an extra DB
// read — making the event replayable and audit-friendly.
type createCardPayload struct {
	ClientID       string  `json:"client_id"`
	Name           string  `json:"name"`
	Email          string  `json:"email"`
	RequestType    string  `json:"request_type"`
	PatrimonyValue float64 `json:"patrimony_value"`
	Priority       string  `json:"priority"`
	Status         string  `json:"status"`
}

// CreateClient registers a new client and atomically enqueues a Pipefy
// createCard event in the outbox. The distributed write problem is solved by
// writing the client row and the outbox event inside a single transaction —
// either both commit or neither does. The Pipefy API call happens
// asynchronously via the outbox worker (PIPEFY-29).
type CreateClient struct {
	transactor repository.Transactor[CreateClientRepos]
}

// NewCreateClient returns a CreateClient use case wired with the given
// transactor. Wire it in main using database.NewGORMTransactor.
func NewCreateClient(transactor repository.Transactor[CreateClientRepos]) *CreateClient {
	return &CreateClient{transactor: transactor}
}

// Execute runs the create-client flow:
//  1. Validates required fields.
//  2. Builds the Client entity (email validation via value object).
//  3. Serializes a self-contained createCard payload for the outbox.
//  4. Atomically persists the client and the outbox event in one transaction.
//
// PipefyCardID will be empty on return — the outbox worker sets it after the
// Pipefy mutation succeeds.
func (uc *CreateClient) Execute(input CreateClientInput) (*entity.Client, error) {
	if err := validateInput(input); err != nil {
		return nil, err
	}

	client, err := entity.NewClient(input.Name, input.Email, input.RequestType, input.PatrimonyValue)
	if err != nil {
		return nil, apperror.NewValidation(err.Error(), err)
	}

	payload, err := buildCreateCardPayload(client)
	if err != nil {
		return nil, apperror.NewInternal(fmt.Errorf("serializing outbox payload: %w", err))
	}

	outboxEvent := entity.NewOutboxEvent(entity.EventTypeCreateCard, payload)

	if err := uc.transactor.Execute(func(repos CreateClientRepos) error {
		if err := repos.Client.Save(client); err != nil {
			return err
		}
		return repos.Outbox.Save(outboxEvent)
	}); err != nil {
		return nil, err
	}

	return client, nil
}

// buildCreateCardPayload serializes the client into a self-contained JSON
// payload so the outbox worker can call Pipefy without loading from the DB.
func buildCreateCardPayload(c *entity.Client) (string, error) {
	p := createCardPayload{
		ClientID:       c.ID.String(),
		Name:           c.Name,
		Email:          c.Email.String(),
		RequestType:    c.RequestType,
		PatrimonyValue: c.PatrimonyValue,
		Priority:       c.Priority.String(),
		Status:         c.Status.String(),
	}
	b, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// validateInput checks that all required fields are present and valid.
func validateInput(input CreateClientInput) error {
	switch {
	case input.Name == "":
		return apperror.NewValidation("cliente_nome is required", nil)
	case input.Email == "":
		return apperror.NewValidation("cliente_email is required", nil)
	case input.RequestType == "":
		return apperror.NewValidation("tipo_solicitacao is required", nil)
	case input.PatrimonyValue < 0:
		return apperror.NewValidation("valor_patrimonio must be zero or greater", nil)
	}
	return nil
}
