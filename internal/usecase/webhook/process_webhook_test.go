package webhook

import (
	"testing"
	"time"

	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/entity"
	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/valueobject"
	"github.com/Casagrande-Lucas/pipefy-integration/pkg/apperror"
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
	client    *entity.Client
	errFind   error
	updated   *entity.Client
	errUpdate error
}

func (s *stubClientRepo) Save(*entity.Client) error                   { return nil }
func (s *stubClientRepo) FindByEmail(string) (*entity.Client, error)  { return nil, nil }
func (s *stubClientRepo) UpdatePipefyCardID(*entity.Client) error     { return nil }
func (s *stubClientRepo) FindByPipefyCardID(string) (*entity.Client, error) {
	return s.client, s.errFind
}
func (s *stubClientRepo) UpdateStatusAndPriority(c *entity.Client) error {
	s.updated = c
	return s.errUpdate
}

type stubProcessedEventRepo struct {
	exists   bool
	errExist error
	saved    *entity.ProcessedEvent
	errSave  error
}

func (s *stubProcessedEventRepo) Exists(string) (bool, error) { return s.exists, s.errExist }
func (s *stubProcessedEventRepo) Save(e *entity.ProcessedEvent) error {
	s.saved = e
	return s.errSave
}

// --- helpers ---

func buildClient() *entity.Client {
	email, _ := valueobject.NewEmail("client@example.com")
	return &entity.Client{
		ID:             valueobject.NewID(),
		Name:           "Test Client",
		Email:          email,
		RequestType:    "consulting",
		PatrimonyValue: 100_000,
		Priority:       valueobject.NewPriority(100_000),
		Status:         valueobject.InitialStatus(),
		PipefyCardID:   "card-42",
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
}

func newUC(cr *stubClientRepo, pe *stubProcessedEventRepo) *ProcessWebhook {
	tx := &stubTransactor[ProcessWebhookRepos]{
		repos: ProcessWebhookRepos{Client: cr, ProcessedEvent: pe},
	}
	return NewProcessWebhook(tx)
}

func validInput() ProcessWebhookInput {
	return ProcessWebhookInput{
		EventID:      "evt-001",
		PipefyCardID: "card-42",
		NewStatus:    "processado",
	}
}

// --- tests ---

func TestProcessWebhook_Success(t *testing.T) {
	t.Parallel()

	cr := &stubClientRepo{client: buildClient()}
	pe := &stubProcessedEventRepo{}
	err := newUC(cr, pe).Execute(validInput())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cr.updated == nil {
		t.Fatal("expected UpdateStatusAndPriority to be called")
	}
	if cr.updated.Status != valueobject.StatusProcessed {
		t.Errorf("status = %q; want %q", cr.updated.Status, valueobject.StatusProcessed)
	}
	if pe.saved == nil {
		t.Fatal("expected ProcessedEvent to be saved")
	}
	if pe.saved.EventID != "evt-001" {
		t.Errorf("saved event_id = %q; want %q", pe.saved.EventID, "evt-001")
	}
}

func TestProcessWebhook_DuplicateEvent_Skipped(t *testing.T) {
	t.Parallel()

	cr := &stubClientRepo{client: buildClient()}
	pe := &stubProcessedEventRepo{exists: true}
	err := newUC(cr, pe).Execute(validInput())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cr.updated != nil {
		t.Error("UpdateStatusAndPriority must not be called for duplicate events")
	}
	if pe.saved != nil {
		t.Error("ProcessedEvent must not be saved for duplicate events")
	}
}

func TestProcessWebhook_ValidationErrors(t *testing.T) {
	t.Parallel()

	base := validInput()
	cases := []struct {
		name  string
		input ProcessWebhookInput
	}{
		{"empty event_id", func() ProcessWebhookInput { i := base; i.EventID = ""; return i }()},
		{"empty card_id", func() ProcessWebhookInput { i := base; i.PipefyCardID = ""; return i }()},
		{"empty status", func() ProcessWebhookInput { i := base; i.NewStatus = ""; return i }()},
		{"invalid status", func() ProcessWebhookInput { i := base; i.NewStatus = "unknown"; return i }()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := newUC(&stubClientRepo{}, &stubProcessedEventRepo{}).Execute(tc.input)
			if err == nil {
				t.Error("expected error; got nil")
			}
		})
	}
}

func TestProcessWebhook_ClientNotFound(t *testing.T) {
	t.Parallel()

	cr := &stubClientRepo{errFind: apperror.NewNotFound("client", nil)}
	pe := &stubProcessedEventRepo{}
	err := newUC(cr, pe).Execute(validInput())

	if err == nil {
		t.Fatal("expected error; got nil")
	}
}
