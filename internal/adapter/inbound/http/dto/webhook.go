package dto

// WebhookRequest represents the Pipefy webhook payload for card field updates.
// Pipefy delivers this body on POST /webhook; the event UUID for idempotency
// is read from the X-Pipefy-Webhook-UUID request header by the handler.
//
// Pipefy webhook reference:
// https://developers.pipefy.com/reference/webhooks
type WebhookRequest struct {
	Action string      `json:"action"`
	Data   WebhookData `json:"data"`
}

// WebhookData holds the event-specific fields from the Pipefy webhook body.
type WebhookData struct {
	Card  WebhookCard  `json:"card"`
	Field WebhookField `json:"field"`
}

// WebhookCard carries the card identifiers included in every Pipefy event.
type WebhookCard struct {
	ID     string `json:"id"`
	PipeID string `json:"pipe_id"`
}

// WebhookField describes a single field change event.
// field_id maps to the Pipefy pipe field identifier (e.g. "status").
type WebhookField struct {
	FieldID  string `json:"field_id"`
	NewValue string `json:"new_value"`
}
