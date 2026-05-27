package client

import (
	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/entity"
	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/port/repository"
	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/port/service"
	"github.com/Casagrande-Lucas/pipefy-integration/pkg/apperror"
)

// CreateClientInput holds the raw data required to create a client.
type CreateClientInput struct {
	Name           string
	Email          string
	RequestType    string
	PatrimonyValue float64
}

// CreateClient is the use case responsible for registering a new client
// and mapping it to a Pipefy card.
type CreateClient struct {
	clientRepo  repository.ClientRepository
	pipefySvc   service.PipefyService
}

// NewCreateClient returns a CreateClient use case with its dependencies injected.
func NewCreateClient(clientRepo repository.ClientRepository, pipefySvc service.PipefyService) *CreateClient {
	return &CreateClient{
		clientRepo: clientRepo,
		pipefySvc:  pipefySvc,
	}
}

// Execute runs the create client flow:
//  1. Validates required fields.
//  2. Builds the Client entity (email validation via value object).
//  3. Creates the Pipefy card and stores the returned card ID on the entity.
//  4. Persists the client with PipefyCardID already set (single DB write).
func (uc *CreateClient) Execute(input CreateClientInput) (*entity.Client, error) {
	if err := validateInput(input); err != nil {
		return nil, err
	}

	client, err := entity.NewClient(input.Name, input.Email, input.RequestType, input.PatrimonyValue)
	if err != nil {
		return nil, apperror.NewValidation(err.Error(), err)
	}

	cardID, err := uc.pipefySvc.CreateCard(client)
	if err != nil {
		return nil, err
	}
	client.PipefyCardID = cardID

	if err := uc.clientRepo.Save(client); err != nil {
		return nil, err
	}

	return client, nil
}

// validateInput checks that all required string fields are present.
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
