package valueobject

import (
	"testing"
)

func TestNewEmail(t *testing.T) {
	t.Parallel()

	valid := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple valid email",
			input: "user@example.com",
			want:  "user@example.com",
		},
		{
			name:  "uppercase is normalized to lowercase",
			input: "User@Example.COM",
			want:  "user@example.com",
		},
		{
			name:  "leading and trailing spaces are trimmed",
			input: "  user@example.com  ",
			want:  "user@example.com",
		},
		{
			name:  "mixed case with spaces",
			input: "  LUCAS@Pipefy.IO  ",
			want:  "lucas@pipefy.io",
		},
		{
			name:  "email with plus sign",
			input: "user+tag@example.com",
			want:  "user+tag@example.com",
		},
		{
			name:  "email with dots in local part",
			input: "first.last@example.com",
			want:  "first.last@example.com",
		},
		{
			name:  "email with subdomain",
			input: "user@mail.example.com",
			want:  "user@mail.example.com",
		},
		{
			name:  "email with hyphen in domain",
			input: "user@my-company.com",
			want:  "user@my-company.com",
		},
	}

	for _, tc := range valid {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewEmail(tc.input)
			if err != nil {
				t.Fatalf("NewEmail(%q) returned unexpected error: %v", tc.input, err)
			}
			if string(got) != tc.want {
				t.Errorf("NewEmail(%q) = %q; want %q", tc.input, got, tc.want)
			}
		})
	}

	invalid := []struct {
		name  string
		input string
	}{
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "only spaces",
			input: "   ",
		},
		{
			name:  "missing at sign",
			input: "userexample.com",
		},
		{
			name:  "missing domain",
			input: "user@",
		},
		{
			name:  "missing local part",
			input: "@example.com",
		},
		{
			name:  "missing TLD",
			input: "user@example",
		},
		{
			name:  "double at sign",
			input: "user@@example.com",
		},
		{
			name:  "spaces inside email",
			input: "user @example.com",
		},
	}

	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewEmail(tc.input)
			if err == nil {
				t.Errorf("NewEmail(%q) expected error; got %q", tc.input, got)
			}
		})
	}
}

func TestEmailString(t *testing.T) {
	t.Parallel()
	e := Email("user@example.com")
	if e.String() != "user@example.com" {
		t.Errorf("Email.String() = %q; want %q", e.String(), "user@example.com")
	}
}

func TestEmailEquality(t *testing.T) {
	t.Parallel()
	a, _ := NewEmail("user@example.com")
	b, _ := NewEmail("USER@EXAMPLE.COM")
	if a != b {
		t.Errorf("expected %q == %q after normalization", a, b)
	}
}
