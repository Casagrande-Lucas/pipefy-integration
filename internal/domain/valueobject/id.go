package valueobject

import (
	"fmt"

	"github.com/google/uuid"
)

// ID is a value object that guarantees its value is a valid UUID v4.
// It is immutable once created via NewID or ParseID.
type ID string

// NewID generates a new random UUID and returns it as an ID.
func NewID() ID {
	return ID(uuid.NewString())
}

// ParseID validates that the given string is a well-formed UUID and returns an ID.
// Use this when rehydrating an entity from persistence or an external input.
func ParseID(value string) (ID, error) {
	if _, err := uuid.Parse(value); err != nil {
		return "", fmt.Errorf("id %q is not a valid UUID: %w", value, err)
	}
	return ID(value), nil
}

// String returns the ID as a plain string.
func (id ID) String() string {
	return string(id)
}

// IsZero reports whether the ID is the zero value (empty string).
func (id ID) IsZero() bool {
	return id == ""
}
