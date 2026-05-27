package mapper

import (
	"github.com/Casagrande-Lucas/pipefy-integration/internal/adapter/outbound/persistence/model"
	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/entity"
)

// ProcessedEventToModel converts a domain ProcessedEvent entity to its GORM persistence model.
func ProcessedEventToModel(e *entity.ProcessedEvent) *model.ProcessedEvent {
	return &model.ProcessedEvent{
		EventID:     e.EventID,
		CardID:      e.CardID,
		ClientEmail: e.ClientEmail,
		ProcessedAt: e.ProcessedAt,
	}
}
