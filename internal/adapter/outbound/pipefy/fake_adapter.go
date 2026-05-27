package pipefy

import (
	"go.uber.org/zap"

	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/entity"
	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/valueobject"
)

// FakeAdapter implements PipefyService for the dev environment.
// It never makes HTTP calls — it logs the GraphQL payload that would be sent
// and returns a generated UUID as the card ID.
type FakeAdapter struct {
	log *zap.Logger
}

// NewFakeAdapter returns a FakeAdapter using the provided logger.
func NewFakeAdapter(log *zap.Logger) *FakeAdapter {
	return &FakeAdapter{log: log}
}

// CreateCard logs the createCard mutation payload and returns a fake card ID.
func (a *FakeAdapter) CreateCard(client *entity.Client) (string, error) {
	fakeCardID := valueobject.NewID().String()

	a.log.Info("pipefy.fake: createCard",
		zap.String("mutation", "createCard"),
		zap.String("pipe_id", "[from config]"),
		zap.String("cliente_nome", client.Name),
		zap.String("cliente_email", client.Email.String()),
		zap.String("tipo_solicitacao", client.RequestType),
		zap.Float64("valor_patrimonio", client.PatrimonyValue),
		zap.String("fake_card_id", fakeCardID),
	)

	return fakeCardID, nil
}

// UpdateCard logs the updateCardField mutation payloads for status and priority.
func (a *FakeAdapter) UpdateCard(client *entity.Client) error {
	a.log.Info("pipefy.fake: updateCardField (status)",
		zap.String("mutation", "updateCardField"),
		zap.String("card_id", client.PipefyCardID),
		zap.String("field_id", "status"),
		zap.String("new_value", client.Status.String()),
	)

	a.log.Info("pipefy.fake: updateCardField (priority)",
		zap.String("mutation", "updateCardField"),
		zap.String("card_id", client.PipefyCardID),
		zap.String("field_id", "priority"),
		zap.String("new_value", client.Priority.String()),
	)

	return nil
}
