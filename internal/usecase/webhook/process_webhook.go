package webhook

import (
	"time"

	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/entity"
	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/port/repository"
	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/valueobject"
	"github.com/Casagrande-Lucas/pipefy-integration/pkg/apperror"
)

// ProcessWebhookInput carries the relevant fields from a Pipefy webhook event.
// Status mapping (Pipefy phase → domain Status) is the handler's responsibility.
type ProcessWebhookInput struct {
	// EventID is Pipefy's unique event identifier used for idempotency.
	EventID string

	// PipefyCardID is the card that triggered the webhook.
	PipefyCardID string

	// NewStatus is the validated domain status derived from the Pipefy phase.
	NewStatus string
}

// ProcessWebhookRepos declares the repositories required by this use case.
type ProcessWebhookRepos struct {
	Client         repository.ClientRepository
	ProcessedEvent repository.ProcessedEventRepository
}

// ProcessWebhook applies a Pipefy webhook event to the local client record.
//
// The flow is idempotent: duplicate event IDs are silently ignored.
// Atomicity is guaranteed by the Transactor — the status update and the
// idempotency record are either both committed or both rolled back.
type ProcessWebhook struct {
	transactor repository.Transactor[ProcessWebhookRepos]
}

// NewProcessWebhook returns a ProcessWebhook use case wired with the given
// transactor.
func NewProcessWebhook(transactor repository.Transactor[ProcessWebhookRepos]) *ProcessWebhook {
	return &ProcessWebhook{transactor: transactor}
}

// Execute processes a Pipefy webhook event:
//  1. Validates input.
//  2. Finds the client by PipefyCardID.
//  3. Inside a transaction:
//     a. Skips if the event was already processed (idempotency).
//     b. Updates the client status and priority.
//     c. Records the event in processed_events.
func (uc *ProcessWebhook) Execute(input ProcessWebhookInput) error {
	if err := validateWebhookInput(input); err != nil {
		return err
	}

	newStatus, err := valueobject.NewStatus(input.NewStatus)
	if err != nil {
		return apperror.NewValidation(err.Error(), err)
	}

	return uc.transactor.Execute(func(repos ProcessWebhookRepos) error {
		// Idempotency check inside the transaction: the unique constraint on
		// event_id is the authoritative guard; this is a cheap fast-path.
		exists, err := repos.ProcessedEvent.Exists(input.EventID)
		if err != nil {
			return err
		}
		if exists {
			return nil
		}

		client, err := repos.Client.FindByPipefyCardID(input.PipefyCardID)
		if err != nil {
			return err
		}

		client.Status = newStatus
		client.Priority = valueobject.NewPriority(client.PatrimonyValue)
		client.UpdatedAt = time.Now().UTC()

		if err := repos.Client.UpdateStatusAndPriority(client); err != nil {
			return err
		}

		processed := entity.NewProcessedEvent(
			input.EventID,
			input.PipefyCardID,
			client.Email.String(),
		)
		return repos.ProcessedEvent.Save(processed)
	})
}

func validateWebhookInput(input ProcessWebhookInput) error {
	switch {
	case input.EventID == "":
		return apperror.NewValidation("event_id is required", nil)
	case input.PipefyCardID == "":
		return apperror.NewValidation("pipefy_card_id is required", nil)
	case input.NewStatus == "":
		return apperror.NewValidation("new_status is required", nil)
	}
	return nil
}
