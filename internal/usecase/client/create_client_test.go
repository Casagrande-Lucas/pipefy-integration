package client

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/entity"
)

// --- stubs ---

type stubTransactor[R any] struct {
	repos R
	err   error
}

func (s *stubTransactor[R]) Execute(fn func(R) error) error {
	if s.err != nil {
		return s.err
	}
	return fn(s.repos)
}

type stubClientRepo struct {
	saved   *entity.Client
	errSave error
}

func (s *stubClientRepo) Save(c *entity.Client) error                    { s.saved = c; return s.errSave }
func (s *stubClientRepo) FindByEmail(string) (*entity.Client, error)     { return nil, nil }
func (s *stubClientRepo) FindByPipefyCardID(string) (*entity.Client, error) { return nil, nil }
func (s *stubClientRepo) UpdateStatusAndPriority(*entity.Client) error   { return nil }
func (s *stubClientRepo) UpdatePipefyCardID(*entity.Client) error        { return nil }

type stubOutboxRepo struct {
	saved   *entity.OutboxEvent
	errSave error
}

func (s *stubOutboxRepo) Save(e *entity.OutboxEvent) error               { s.saved = e; return s.errSave }
func (s *stubOutboxRepo) FindPending(int) ([]*entity.OutboxEvent, error) { return nil, nil }
func (s *stubOutboxRepo) UpdateStatus(*entity.OutboxEvent) error         { return nil }

// --- helpers ---

func newUC(clientErr, outboxErr error) (*CreateClient, *stubClientRepo, *stubOutboxRepo) {
	cr := &stubClientRepo{errSave: clientErr}
	or := &stubOutboxRepo{errSave: outboxErr}
	tx := &stubTransactor[CreateClientRepos]{repos: CreateClientRepos{Client: cr, Outbox: or}}
	return NewCreateClient(tx), cr, or
}

func validInput() CreateClientInput {
	return CreateClientInput{
		Name:           "Lucas",
		Email:          "lucas@example.com",
		RequestType:    "consultoria",
		PatrimonyValue: 150_000,
	}
}

// --- tests ---

func TestCreateClient_Success(t *testing.T) {
	t.Parallel()

	uc, clientRepo, outboxRepo := newUC(nil, nil)
	got, err := uc.Execute(validInput())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || clientRepo.saved == nil {
		t.Fatal("client not persisted")
	}
	if outboxRepo.saved == nil {
		t.Fatal("outbox event not persisted")
	}
	if outboxRepo.saved.EventType != entity.EventTypeCreateCard {
		t.Errorf("event type = %q; want %q", outboxRepo.saved.EventType, entity.EventTypeCreateCard)
	}

	// Payload must be valid JSON and carry the right email.
	var p createCardPayload
	if err := json.Unmarshal([]byte(outboxRepo.saved.Payload), &p); err != nil {
		t.Fatalf("invalid payload JSON: %v", err)
	}
	if p.Email != "lucas@example.com" {
		t.Errorf("payload.email = %q; want %q", p.Email, "lucas@example.com")
	}
	if p.ClientID == "" {
		t.Error("payload.client_id must not be empty")
	}
}

func TestCreateClient_ValidationErrors(t *testing.T) {
	t.Parallel()

	base := validInput()
	cases := []struct {
		name  string
		input CreateClientInput
	}{
		{"empty name", func() CreateClientInput { i := base; i.Name = ""; return i }()},
		{"empty email", func() CreateClientInput { i := base; i.Email = ""; return i }()},
		{"invalid email", func() CreateClientInput { i := base; i.Email = "not-an-email"; return i }()},
		{"empty request_type", func() CreateClientInput { i := base; i.RequestType = ""; return i }()},
		{"negative patrimony", func() CreateClientInput { i := base; i.PatrimonyValue = -1; return i }()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			uc, _, _ := newUC(nil, nil)
			_, err := uc.Execute(tc.input)
			if err == nil {
				t.Error("expected error; got nil")
			}
		})
	}
}

func TestCreateClient_ClientSaveError(t *testing.T) {
	t.Parallel()

	uc, _, _ := newUC(errors.New("db error"), nil)
	_, err := uc.Execute(validInput())
	if err == nil {
		t.Fatal("expected error; got nil")
	}
}

func TestCreateClient_OutboxSaveError(t *testing.T) {
	t.Parallel()

	uc, _, _ := newUC(nil, errors.New("db error"))
	_, err := uc.Execute(validInput())
	if err == nil {
		t.Fatal("expected error; got nil")
	}
}
