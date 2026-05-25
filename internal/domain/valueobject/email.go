package valueobject

import (
	"fmt"
	"regexp"
	"strings"
)

// Email represents a validated, normalized email address.
// It is immutable once created via NewEmail.
type Email string

// emailRegex validates the structure of an email address.
// Covers standard formats; intentionally excludes quoted strings and IP literals.
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// NewEmail validates and normalizes the raw email string.
// It trims whitespace and lowercases the value before validation.
func NewEmail(raw string) (Email, error) {
	value := strings.ToLower(strings.TrimSpace(raw))

	if value == "" {
		return "", fmt.Errorf("email is required")
	}

	if !emailRegex.MatchString(value) {
		return "", fmt.Errorf("email %q is invalid", value)
	}

	return Email(value), nil
}

// String returns the email as a plain string.
func (e Email) String() string {
	return string(e)
}
