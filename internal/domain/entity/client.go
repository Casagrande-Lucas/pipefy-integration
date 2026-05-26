package entity

import (
	"time"

	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/valueobject"
)

// Client represents a managed client in the system.
// It is the aggregate root of the client domain.
type Client struct {
	ID             valueobject.ID
	Name           string
	Email          valueobject.Email
	RequestType    string
	PatrimonyValue float64
	Priority       valueobject.Priority
	Status         valueobject.Status
	PipefyCardID   string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// NewClient constructs a Client with validated value objects and initial state.
// A UUID is assigned at construction time so the entity always has a valid identity.
// Persistence is the repository's responsibility.
func NewClient(name, email, requestType string, patrimonyValue float64) (*Client, error) {
	e, err := valueobject.NewEmail(email)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	return &Client{
		ID:             valueobject.NewID(),
		Name:           name,
		Email:          e,
		RequestType:    requestType,
		PatrimonyValue: patrimonyValue,
		Priority:       valueobject.NewPriority(patrimonyValue),
		Status:         valueobject.InitialStatus(),
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}
