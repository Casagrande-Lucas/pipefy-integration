package pipefy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/entity"
	"github.com/Casagrande-Lucas/pipefy-integration/pkg/apperror"
)

const (
	pipefyGraphQLEndpoint = "https://api.pipefy.com/graphql"
	httpTimeout           = 10 * time.Second
)

// graphQLRequest is the JSON body sent to the Pipefy GraphQL endpoint.
type graphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

// graphQLResponse is the JSON body returned by the Pipefy GraphQL endpoint.
type graphQLResponse struct {
	Data   json.RawMessage  `json:"data"`
	Errors []graphQLError   `json:"errors"`
}

type graphQLError struct {
	Message string `json:"message"`
}

// createCardData maps the data field of a createCard response.
type createCardData struct {
	CreateCard struct {
		Card struct {
			ID string `json:"id"`
		} `json:"card"`
	} `json:"createCard"`
}

// RealAdapter implements PipefyService for staging and prod environments.
// When simulate=true (staging), it logs the mutation payload without sending HTTP requests.
// When simulate=false (prod), it sends the exact GraphQL mutations to the Pipefy API.
type RealAdapter struct {
	token      string
	pipeID     string
	simulate   bool
	log        *zap.Logger
	httpClient *http.Client
}

// NewRealAdapter returns a RealAdapter configured for staging or prod.
func NewRealAdapter(token, pipeID string, simulate bool, log *zap.Logger) *RealAdapter {
	return &RealAdapter{
		token:    token,
		pipeID:   pipeID,
		simulate: simulate,
		log:      log,
		httpClient: &http.Client{
			Timeout: httpTimeout,
		},
	}
}

// CreateCard sends (or simulates) the Pipefy createCard GraphQL mutation.
// Mutation reference: https://developers.pipefy.com/reference/create-a-card-with-the-required-fields-fulfilled
func (a *RealAdapter) CreateCard(client *entity.Client) (string, error) {
	const mutation = `
		mutation CreateCard($input: CreateCardInput!) {
			createCard(input: $input) {
				card {
					id
				}
			}
		}`

	variables := map[string]any{
		"input": map[string]any{
			"pipe_id": a.pipeID,
			"title":   client.Name,
			"fields_attributes": []map[string]any{
				{"field_id": "cliente_nome", "field_value": client.Name},
				{"field_id": "cliente_email", "field_value": client.Email.String()},
				{"field_id": "tipo_solicitacao", "field_value": client.RequestType},
				{"field_id": "valor_patrimonio", "field_value": fmt.Sprintf("%.2f", client.PatrimonyValue)},
				{"field_id": "prioridade", "field_value": pipefyPriorityLabel(client.Priority.String())},
			},
		},
	}

	if a.simulate {
		a.log.Info("pipefy.real: createCard (simulate)",
			zap.String("mutation", "createCard"),
			zap.Any("variables", variables),
		)
		return "simulated-card-id", nil
	}

	var data createCardData
	if err := a.do(mutation, variables, &data); err != nil {
		return "", err
	}

	cardID := data.CreateCard.Card.ID
	if cardID == "" {
		return "", apperror.NewInternal(fmt.Errorf("pipefy returned empty card ID"))
	}

	a.log.Info("pipefy.real: createCard success", zap.String("card_id", cardID))
	return cardID, nil
}

// UpdateCard sends (or simulates) two Pipefy updateCardField GraphQL mutations:
// one for the status field and one for the priority field.
// Mutation reference: https://developers.pipefy.com/reference/cards#card-mutations
func (a *RealAdapter) UpdateCard(client *entity.Client) error {
	const mutation = `
		mutation UpdateCardField($input: UpdateCardFieldInput!) {
			updateCardField(input: $input) {
				card {
					id
				}
				success
			}
		}`

	updates := []struct {
		fieldID string
		value   string
	}{
		{"status", client.Status.String()},
		{"priority", client.Priority.String()},
	}

	for _, u := range updates {
		variables := map[string]any{
			"input": map[string]any{
				"card_id":   client.PipefyCardID,
				"field_id":  u.fieldID,
				"new_value": u.value,
			},
		}

		if a.simulate {
			a.log.Info("pipefy.real: updateCardField (simulate)",
				zap.String("mutation", "updateCardField"),
				zap.Any("variables", variables),
			)
			continue
		}

		if err := a.do(mutation, variables, nil); err != nil {
			return err
		}

		a.log.Info("pipefy.real: updateCardField success",
			zap.String("card_id", client.PipefyCardID),
			zap.String("field_id", u.fieldID),
		)
	}

	return nil
}

// do executes a GraphQL request against the Pipefy API and unmarshals the data field into dest.
// Pass nil for dest when the response data is not needed.
func (a *RealAdapter) do(query string, variables map[string]any, dest any) error {
	body, err := json.Marshal(graphQLRequest{Query: query, Variables: variables})
	if err != nil {
		return apperror.NewInternal(fmt.Errorf("marshalling graphql request: %w", err))
	}

	req, err := http.NewRequest(http.MethodPost, pipefyGraphQLEndpoint, bytes.NewReader(body))
	if err != nil {
		return apperror.NewInternal(fmt.Errorf("creating http request: %w", err))
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.token)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return apperror.NewInternal(fmt.Errorf("sending request to pipefy: %w", err))
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return apperror.NewInternal(fmt.Errorf("reading pipefy response: %w", err))
	}

	if resp.StatusCode != http.StatusOK {
		return apperror.NewInternal(fmt.Errorf("pipefy returned status %d: %s", resp.StatusCode, raw))
	}

	var gqlResp graphQLResponse
	if err := json.Unmarshal(raw, &gqlResp); err != nil {
		return apperror.NewInternal(fmt.Errorf("unmarshalling pipefy response: %w", err))
	}

	if len(gqlResp.Errors) > 0 {
		return apperror.NewInternal(fmt.Errorf("pipefy graphql error: %s", gqlResp.Errors[0].Message))
	}

	if dest != nil && gqlResp.Data != nil {
		if err := json.Unmarshal(gqlResp.Data, dest); err != nil {
			return apperror.NewInternal(fmt.Errorf("unmarshalling pipefy data: %w", err))
		}
	}

	return nil
}

// pipefyPriorityLabel maps a domain priority string to the Pipefy radio
// option label defined in the pipe's "Prioridade" field.
//
// Domain → Pipefy option label
//
//	prioridade_alta   → Alta
//	prioridade_normal → Normal
func pipefyPriorityLabel(priority string) string {
	if priority == "prioridade_alta" {
		return "Alta"
	}
	return "Normal"
}
