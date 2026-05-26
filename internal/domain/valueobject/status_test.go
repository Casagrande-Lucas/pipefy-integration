package valueobject

import (
	"testing"
)

func TestInitialStatus(t *testing.T) {
	t.Parallel()
	got := InitialStatus()
	if got != StatusPending {
		t.Errorf("InitialStatus() = %q; want %q", got, StatusPending)
	}
}

func TestNewStatus(t *testing.T) {
	t.Parallel()

	valid := []struct {
		name  string
		input string
		want  Status
	}{
		{
			name:  "pending status",
			input: "aguardando_analise",
			want:  StatusPending,
		},
		{
			name:  "processed status",
			input: "processado",
			want:  StatusProcessed,
		},
	}

	for _, tc := range valid {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewStatus(tc.input)
			if err != nil {
				t.Fatalf("NewStatus(%q) returned unexpected error: %v", tc.input, err)
			}
			if got != tc.want {
				t.Errorf("NewStatus(%q) = %q; want %q", tc.input, got, tc.want)
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
			name:  "unknown value",
			input: "em_processamento",
		},
		{
			name:  "uppercase variant",
			input: "PROCESSADO",
		},
		{
			name:  "partial match",
			input: "processad",
		},
		{
			name:  "random string",
			input: "invalid_status",
		},
	}

	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewStatus(tc.input)
			if err == nil {
				t.Errorf("NewStatus(%q) expected error; got %q", tc.input, got)
			}
		})
	}
}

func TestStatusIsValid(t *testing.T) {
	t.Parallel()

	if !StatusPending.IsValid() {
		t.Errorf("StatusPending.IsValid() = false; want true")
	}
	if !StatusProcessed.IsValid() {
		t.Errorf("StatusProcessed.IsValid() = false; want true")
	}
	if Status("outro").IsValid() {
		t.Errorf(`Status("outro").IsValid() = true; want false`)
	}
}

func TestStatusString(t *testing.T) {
	t.Parallel()

	if StatusPending.String() != "aguardando_analise" {
		t.Errorf("StatusPending.String() = %q; want %q", StatusPending.String(), "aguardando_analise")
	}
	if StatusProcessed.String() != "processado" {
		t.Errorf("StatusProcessed.String() = %q; want %q", StatusProcessed.String(), "processado")
	}
}
