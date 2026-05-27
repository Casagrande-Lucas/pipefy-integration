package outbox

import (
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/entity"
	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/port/repository"
	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/port/service"
	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/valueobject"
	"github.com/Casagrande-Lucas/pipefy-integration/pkg/apperror"
)

const defaultBatchSize = 10

// createCardPayload mirrors the payload written by the CreateClient use case.
// It is self-contained so the worker can call Pipefy without a DB read.
type createCardPayload struct {
	ClientID       string  `json:"client_id"`
	Name           string  `json:"name"`
	Email          string  `json:"email"`
	RequestType    string  `json:"request_type"`
	PatrimonyValue float64 `json:"patrimony_value"`
}

// updateCardPayload mirrors the payload written by the ProcessWebhook use case.
// Priority and Status are stored as their string representations.
type updateCardPayload struct {
	ClientID     string `json:"client_id"`
	PipefyCardID string `json:"pipefy_card_id"`
	Priority     string `json:"priority"`
	Status       string `json:"status"`
}

// ProcessOutboxResult summarises a single worker run.
type ProcessOutboxResult struct {
	Processed int
	Failed    int
}

// ProcessOutbox is the background worker use case that polls the outbox table,
// dispatches Pipefy mutations, and updates event statuses.
//
// For each pending event the worker:
//  1. Marks the event processing and persists the incremented attempt counter.
//  2. Dispatches the appropriate Pipefy mutation (createCard or updateCardField).
//  3. On success: marks done; for create_card also stores the returned card ID.
//  4. On failure: marks failed — back to pending for retry until MaxAttempts is
//     exhausted, then permanently failed.
type ProcessOutbox struct {
	outboxRepo repository.OutboxRepository
	clientRepo repository.ClientRepository
	pipefySvc  service.PipefyService
	log        *zap.Logger
	batchSize  int
}

// NewProcessOutbox returns a ProcessOutbox wired with the given dependencies.
// Pass batchSize <= 0 to use the default (10).
func NewProcessOutbox(
	outboxRepo repository.OutboxRepository,
	clientRepo repository.ClientRepository,
	pipefySvc service.PipefyService,
	log *zap.Logger,
	batchSize int,
) *ProcessOutbox {
	if batchSize <= 0 {
		batchSize = defaultBatchSize
	}
	return &ProcessOutbox{
		outboxRepo: outboxRepo,
		clientRepo: clientRepo,
		pipefySvc:  pipefySvc,
		log:        log,
		batchSize:  batchSize,
	}
}

// Execute polls up to batchSize pending events and processes each one.
// Individual event failures are logged and counted; they do not abort the batch.
func (w *ProcessOutbox) Execute() (ProcessOutboxResult, error) {
	events, err := w.outboxRepo.FindPending(w.batchSize)
	if err != nil {
		return ProcessOutboxResult{}, fmt.Errorf("fetching pending outbox events: %w", err)
	}

	var result ProcessOutboxResult
	for _, event := range events {
		if err := w.process(event); err != nil {
			result.Failed++
			w.log.Warn("outbox: event failed",
				zap.String("event_id", event.ID.String()),
				zap.String("event_type", string(event.EventType)),
				zap.Int("attempts", event.Attempts),
				zap.Bool("exhausted", !event.CanRetry()),
				zap.Error(err),
			)
			continue
		}
		result.Processed++
		w.log.Info("outbox: event processed",
			zap.String("event_id", event.ID.String()),
			zap.String("event_type", string(event.EventType)),
		)
	}

	return result, nil
}

// process handles a single outbox event end-to-end.
func (w *ProcessOutbox) process(event *entity.OutboxEvent) error {
	// Persist the attempt increment before the API call so a mid-flight crash
	// does not silently lose the attempt.
	event.MarkProcessing()
	if err := w.outboxRepo.UpdateStatus(event); err != nil {
		return fmt.Errorf("marking event processing: %w", err)
	}

	dispatchErr := w.dispatch(event)

	if dispatchErr != nil {
		event.MarkFailed(dispatchErr)
		if err := w.outboxRepo.UpdateStatus(event); err != nil {
			return fmt.Errorf("marking event failed: %w", err)
		}
		return dispatchErr
	}

	event.MarkDone()
	if err := w.outboxRepo.UpdateStatus(event); err != nil {
		return fmt.Errorf("marking event done: %w", err)
	}

	return nil
}

// dispatch routes the event to the appropriate handler by event type.
func (w *ProcessOutbox) dispatch(event *entity.OutboxEvent) error {
	switch event.EventType {
	case entity.EventTypeCreateCard:
		return w.handleCreateCard(event)
	case entity.EventTypeUpdateCard:
		return w.handleUpdateCard(event)
	default:
		return apperror.NewInternal(
			fmt.Errorf("unknown outbox event type %q", event.EventType),
		)
	}
}

// handleCreateCard calls the Pipefy createCard mutation and persists the
// returned card ID on the client record.
func (w *ProcessOutbox) handleCreateCard(event *entity.OutboxEvent) error {
	var p createCardPayload
	if err := json.Unmarshal([]byte(event.Payload), &p); err != nil {
		return apperror.NewInternal(fmt.Errorf("unmarshalling create_card payload: %w", err))
	}

	client, err := buildClientForCreate(p)
	if err != nil {
		return err
	}

	cardID, err := w.pipefySvc.CreateCard(client)
	if err != nil {
		return err
	}

	client.PipefyCardID = cardID
	client.UpdatedAt = time.Now().UTC()
	if err := w.clientRepo.UpdatePipefyCardID(client); err != nil {
		return fmt.Errorf("persisting pipefy card id for client %s: %w", client.ID, err)
	}

	return nil
}

// handleUpdateCard calls the Pipefy updateCardField mutations for priority and
// status.
func (w *ProcessOutbox) handleUpdateCard(event *entity.OutboxEvent) error {
	var p updateCardPayload
	if err := json.Unmarshal([]byte(event.Payload), &p); err != nil {
		return apperror.NewInternal(fmt.Errorf("unmarshalling update_card payload: %w", err))
	}

	client, err := buildClientForUpdate(p)
	if err != nil {
		return err
	}

	return w.pipefySvc.UpdateCard(client)
}

// buildClientForCreate reconstructs a minimal entity.Client from the
// create_card payload. Priority is recalculated from patrimony_value to honour
// the domain business rule rather than trusting a stored string.
func buildClientForCreate(p createCardPayload) (*entity.Client, error) {
	id, err := valueobject.ParseID(p.ClientID)
	if err != nil {
		return nil, apperror.NewInternal(fmt.Errorf("invalid client_id in create payload: %w", err))
	}

	email, err := valueobject.NewEmail(p.Email)
	if err != nil {
		return nil, apperror.NewInternal(fmt.Errorf("invalid email in create payload: %w", err))
	}

	return &entity.Client{
		ID:             id,
		Name:           p.Name,
		Email:          email,
		RequestType:    p.RequestType,
		PatrimonyValue: p.PatrimonyValue,
		Priority:       valueobject.NewPriority(p.PatrimonyValue),
	}, nil
}

// buildClientForUpdate reconstructs a minimal entity.Client for the
// updateCardField mutation. Only PipefyCardID, Priority and Status are required
// by the Pipefy adapter; Priority and Status are cast directly from the stored
// string values (already validated at write time).
func buildClientForUpdate(p updateCardPayload) (*entity.Client, error) {
	id, err := valueobject.ParseID(p.ClientID)
	if err != nil {
		return nil, apperror.NewInternal(fmt.Errorf("invalid client_id in update payload: %w", err))
	}

	status, err := valueobject.NewStatus(p.Status)
	if err != nil {
		return nil, apperror.NewInternal(fmt.Errorf("invalid status in update payload: %w", err))
	}

	return &entity.Client{
		ID:           id,
		PipefyCardID: p.PipefyCardID,
		Priority:     valueobject.Priority(p.Priority),
		Status:       status,
	}, nil
}
