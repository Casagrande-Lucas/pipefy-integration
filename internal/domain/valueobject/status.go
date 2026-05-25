package valueobject

import "fmt"

// Status represents the processing status of a client record.
type Status string

const (
	// StatusPending is the initial status assigned when a client is created.
	StatusPending Status = "aguardando_analise"

	// StatusProcessed is assigned after a webhook event is successfully processed.
	StatusProcessed Status = "processado"
)

// InitialStatus returns the status that every new client starts with.
func InitialStatus() Status {
	return StatusPending
}

// NewStatus parses and validates a status string.
// Returns an error if the value is not a recognized status.
func NewStatus(value string) (Status, error) {
	s := Status(value)
	if !s.IsValid() {
		return "", fmt.Errorf("status %q is invalid: accepted values are %q, %q",
			value, StatusPending, StatusProcessed)
	}
	return s, nil
}

// IsValid reports whether the status is a recognized value.
func (s Status) IsValid() bool {
	return s == StatusPending || s == StatusProcessed
}

// String returns the status as a plain string.
func (s Status) String() string {
	return string(s)
}
