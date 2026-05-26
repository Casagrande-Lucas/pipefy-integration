package valueobject

import (
	"strings"
	"testing"
)

func TestNewID(t *testing.T) {
	t.Parallel()

	a := NewID()
	b := NewID()

	if a.IsZero() {
		t.Error("NewID() returned zero value")
	}
	if a == b {
		t.Error("NewID() returned same value twice — not unique")
	}
	// UUID v4 has 36 chars (8-4-4-4-12 + dashes)
	if len(a.String()) != 36 {
		t.Errorf("NewID() length = %d; want 36", len(a.String()))
	}
}

func TestParseID(t *testing.T) {
	t.Parallel()

	valid := []string{
		"550e8400-e29b-41d4-a716-446655440000",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
	}

	for _, v := range valid {
		t.Run(v, func(t *testing.T) {
			t.Parallel()
			id, err := ParseID(v)
			if err != nil {
				t.Fatalf("ParseID(%q) unexpected error: %v", v, err)
			}
			if id.String() != strings.ToLower(v) {
				t.Errorf("ParseID(%q) = %q; want %q", v, id.String(), strings.ToLower(v))
			}
		})
	}

	invalid := []string{
		"",
		"not-a-uuid",
		"550e8400-e29b-41d4-a716",
		"550e8400-e29b-41d4-a716-44665544000Z",
	}

	for _, v := range invalid {
		t.Run("invalid_"+v, func(t *testing.T) {
			t.Parallel()
			_, err := ParseID(v)
			if err == nil {
				t.Errorf("ParseID(%q) expected error; got nil", v)
			}
		})
	}
}

func TestIDIsZero(t *testing.T) {
	t.Parallel()

	if !ID("").IsZero() {
		t.Error("zero ID.IsZero() = false; want true")
	}
	if NewID().IsZero() {
		t.Error("NewID().IsZero() = true; want false")
	}
}
