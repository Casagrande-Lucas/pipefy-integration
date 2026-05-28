package dto

import (
	"time"

	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/entity"
)

// CreateClientRequest is the JSON body for POST /clients.
// Field names follow the Pipefy pipe field IDs used in the technical spec.
type CreateClientRequest struct {
	Nome            string  `json:"cliente_nome"      binding:"required"`
	Email           string  `json:"cliente_email"     binding:"required"`
	TipoSolicitacao string  `json:"tipo_solicitacao"  binding:"required"`
	ValorPatrimonio float64 `json:"valor_patrimonio"`
}

// ClientResponse is the JSON body returned after a successful client operation.
type ClientResponse struct {
	ID              string  `json:"id"`
	Nome            string  `json:"nome"`
	Email           string  `json:"email"`
	TipoSolicitacao string  `json:"tipo_solicitacao"`
	ValorPatrimonio float64 `json:"valor_patrimonio"`
	Prioridade      string  `json:"prioridade"`
	Status          string  `json:"status"`
	PipefyCardID    string  `json:"pipefy_card_id"`
	CriadoEm        string  `json:"criado_em"`
	AtualizadoEm    string  `json:"atualizado_em"`
}

// NewClientResponse builds a ClientResponse from a domain entity.
func NewClientResponse(c *entity.Client) ClientResponse {
	return ClientResponse{
		ID:              c.ID.String(),
		Nome:            c.Name,
		Email:           c.Email.String(),
		TipoSolicitacao: c.RequestType,
		ValorPatrimonio: c.PatrimonyValue,
		Prioridade:      c.Priority.String(),
		Status:          c.Status.String(),
		PipefyCardID:    c.PipefyCardID,
		CriadoEm:        c.CreatedAt.Format(time.RFC3339),
		AtualizadoEm:    c.UpdatedAt.Format(time.RFC3339),
	}
}
