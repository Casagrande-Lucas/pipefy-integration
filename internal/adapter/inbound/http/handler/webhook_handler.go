package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Casagrande-Lucas/pipefy-integration/internal/adapter/inbound/http/dto"
	"github.com/Casagrande-Lucas/pipefy-integration/internal/usecase/webhook"
)

// pipefyWebhookUUIDHeader is the Pipefy header carrying a unique event identifier.
// Used as the idempotency key to prevent duplicate processing.
const pipefyWebhookUUIDHeader = "X-Pipefy-Webhook-UUID"

// processWebhookUseCase is the port this handler depends on.
type processWebhookUseCase interface {
	Execute(input webhook.ProcessWebhookInput) error
}

// WebhookHandler handles inbound Pipefy webhook events.
type WebhookHandler struct {
	processWebhook processWebhookUseCase
	log            *zap.Logger
}

// NewWebhookHandler returns a WebhookHandler with the given dependencies.
func NewWebhookHandler(uc processWebhookUseCase, log *zap.Logger) *WebhookHandler {
	return &WebhookHandler{processWebhook: uc, log: log}
}

// Handle handles POST /webhook.
//
// The idempotency key is taken from the X-Pipefy-Webhook-UUID header when
// present. If absent, a deterministic key is derived from the event fields
// so duplicate deliveries are still deduplicated.
//
//	@Summary      Receive Pipefy webhook
//	@Description  Processes a Pipefy card.field.update event and updates the client status.
//	@Tags         webhook
//	@Accept       json
//	@Produce      json
//	@Param        X-Pipefy-Webhook-UUID  header    string               false  "Pipefy event UUID"
//	@Param        body                   body      dto.WebhookRequest   true   "Webhook payload"
//	@Success      204
//	@Failure      400  {object}  dto.ErrorResponse
//	@Failure      500  {object}  dto.ErrorResponse
//	@Router       /webhook [post]
func (h *WebhookHandler) Handle(c *gin.Context) {
	var req dto.WebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	eventID := c.GetHeader(pipefyWebhookUUIDHeader)
	if eventID == "" {
		// Derive a deterministic key from the event fields so deduplication
		// still works even when the header is missing.
		eventID = fmt.Sprintf("%s:%s:%s", req.Action, req.Data.Card.ID, req.Data.Field.FieldID)
	}

	input := webhook.ProcessWebhookInput{
		EventID:      eventID,
		PipefyCardID: req.Data.Card.ID,
		NewStatus:    req.Data.Field.NewValue,
	}

	if err := h.processWebhook.Execute(input); err != nil {
		h.log.Warn("process webhook failed",
			zap.String("event_id", eventID),
			zap.String("card_id", req.Data.Card.ID),
			zap.Error(err),
		)
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
