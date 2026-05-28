package outbox

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/entity"
	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/valueobject"
)

// --- stubs ---

type stubOutboxRepo struct {
	pending    []*entity.OutboxEvent
	updateLog  []*entity.OutboxEvent
	errFind    error
	errUpdate  error
}

func (s *stubOutboxRepo) Save(*entity.OutboxEvent) error { return nil }
func (s *stubOutboxRepo) FindPending(int) ([]*entity.OutboxEvent, error) {
	return s.pending, s.errFind
}
func (s *stubOutboxRepo) UpdateStatus(e *entity.OutboxEvent) error {
	clone := *e
	s.updateLog = append(s.updateLog, &clone)
	return s.errUpdate
}

type stubClientRepo struct {
	cardIDSet  string
	errUpdate  error
}

func (s *stubClientRepo) Save(*entity.Client) error                      { return nil }
func (s *stubClientRepo) FindByEmail(string) (*entity.Client, error)     { return nil, nil }
func (s *stubClientRepo) FindByPipefyCardID(string) (*entity.Client, error) { return nil, nil }
func (s *stubClientRepo) UpdateStatusAndPriority(*entity.Client) error   { return nil }
func (s *stubClientRepo) UpdatePipefyCardID(c *entity.Client) error {
	s.cardIDSet = c.PipefyCardID
	return s.errUpdate
}

type stubPipefySvc struct {
	returnCardID  string
	errCreate     error
	errUpdate     error
	createCalled  bool
	updateCalled  bool
}

func (s *stubPipefySvc) CreateCard(*entity.Client) (string, error) {
	s.createCalled = true
	return s.returnCardID, s.errCreate
}
func (s *stubPipefySvc) UpdateCard(*entity.Client) error {
	s.updateCalled = true
	return s.errUpdate
}

// --- helpers ---

func newWorker(outbox *stubOutboxRepo, client *stubClientRepo, svc *stubPipefySvc) *ProcessOutbox {
	return NewProcessOutbox(outbox, client, svc, zap.NewNop(), 10)
}

func makeCreateCardEvent(attempts, max int) *entity.OutboxEvent {
	p, _ := json.Marshal(createCardPayload{
		ClientID:       "550e8400-e29b-41d4-a716-446655440000",
		Name:           "Test",
		Email:          "test@example.com",
		RequestType:    "consulting",
		PatrimonyValue: 100_000,
	})
	return &entity.OutboxEvent{
		ID:          valueobject.NewID(),
		EventType:   entity.EventTypeCreateCard,
		Payload:     string(p),
		Status:      valueobject.OutboxStatusPending,
		Attempts:    attempts,
		MaxAttempts: max,
		CreatedAt:   time.Now().UTC(),
	}
}

func makeUpdateCardEvent() *entity.OutboxEvent {
	p, _ := json.Marshal(updateCardPayload{
		ClientID:     "550e8400-e29b-41d4-a716-446655440000",
		PipefyCardID: "card-42",
		Priority:     "prioridade_normal",
		Status:       "processado",
	})
	return &entity.OutboxEvent{
		ID:          valueobject.NewID(),
		EventType:   entity.EventTypeUpdateCard,
		Payload:     string(p),
		Status:      valueobject.OutboxStatusPending,
		Attempts:    0,
		MaxAttempts: 3,
		CreatedAt:   time.Now().UTC(),
	}
}

// --- tests ---

func TestProcessOutbox_CreateCard_Success(t *testing.T) {
	t.Parallel()

	outboxRepo := &stubOutboxRepo{pending: []*entity.OutboxEvent{makeCreateCardEvent(0, 3)}}
	clientRepo := &stubClientRepo{}
	svc := &stubPipefySvc{returnCardID: "pipefy-card-99"}

	result, err := newWorker(outboxRepo, clientRepo, svc).Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Processed != 1 || result.Failed != 0 {
		t.Errorf("result = %+v; want Processed=1 Failed=0", result)
	}
	if !svc.createCalled {
		t.Error("expected CreateCard to be called")
	}
	if clientRepo.cardIDSet != "pipefy-card-99" {
		t.Errorf("cardIDSet = %q; want %q", clientRepo.cardIDSet, "pipefy-card-99")
	}
	// UpdateStatus called twice: processing + done.
	if len(outboxRepo.updateLog) != 2 {
		t.Errorf("UpdateStatus calls = %d; want 2", len(outboxRepo.updateLog))
	}
	if outboxRepo.updateLog[0].Status != valueobject.OutboxStatusProcessing {
		t.Errorf("first status = %q; want processing", outboxRepo.updateLog[0].Status)
	}
	if outboxRepo.updateLog[1].Status != valueobject.OutboxStatusDone {
		t.Errorf("second status = %q; want done", outboxRepo.updateLog[1].Status)
	}
}

func TestProcessOutbox_UpdateCard_Success(t *testing.T) {
	t.Parallel()

	outboxRepo := &stubOutboxRepo{pending: []*entity.OutboxEvent{makeUpdateCardEvent()}}
	svc := &stubPipefySvc{}

	result, err := newWorker(outboxRepo, &stubClientRepo{}, svc).Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Processed != 1 {
		t.Errorf("processed = %d; want 1", result.Processed)
	}
	if !svc.updateCalled {
		t.Error("expected UpdateCard to be called")
	}
}

func TestProcessOutbox_APIFailure_EventRetried(t *testing.T) {
	t.Parallel()

	event := makeCreateCardEvent(0, 3)
	outboxRepo := &stubOutboxRepo{pending: []*entity.OutboxEvent{event}}
	svc := &stubPipefySvc{errCreate: errors.New("pipefy timeout")}

	result, err := newWorker(outboxRepo, &stubClientRepo{}, svc).Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Failed != 1 || result.Processed != 0 {
		t.Errorf("result = %+v; want Failed=1", result)
	}
	last := outboxRepo.updateLog[len(outboxRepo.updateLog)-1]
	if last.Status != valueobject.OutboxStatusPending {
		t.Errorf("status after retry = %q; want pending", last.Status)
	}
}

func TestProcessOutbox_APIFailure_EventExhausted(t *testing.T) {
	t.Parallel()

	// Attempts=2, MaxAttempts=3 → after MarkProcessing attempts=3 → CanRetry=false.
	event := makeCreateCardEvent(2, 3)
	outboxRepo := &stubOutboxRepo{pending: []*entity.OutboxEvent{event}}
	svc := &stubPipefySvc{errCreate: errors.New("permanent error")}

	newWorker(outboxRepo, &stubClientRepo{}, svc).Execute() //nolint:errcheck

	last := outboxRepo.updateLog[len(outboxRepo.updateLog)-1]
	if last.Status != valueobject.OutboxStatusFailed {
		t.Errorf("exhausted status = %q; want failed", last.Status)
	}
}

func TestProcessOutbox_EmptyBatch(t *testing.T) {
	t.Parallel()

	result, err := newWorker(&stubOutboxRepo{}, &stubClientRepo{}, &stubPipefySvc{}).Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Processed != 0 || result.Failed != 0 {
		t.Errorf("result = %+v; want 0/0", result)
	}
}

func TestProcessOutbox_BatchContinuesOnSingleFailure(t *testing.T) {
	t.Parallel()

	events := []*entity.OutboxEvent{
		makeCreateCardEvent(0, 3), // will fail
		makeUpdateCardEvent(),     // must succeed
	}
	outboxRepo := &stubOutboxRepo{pending: events}
	svc := &stubPipefySvc{errCreate: errors.New("api down")}

	result, err := newWorker(outboxRepo, &stubClientRepo{}, svc).Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Processed != 1 || result.Failed != 1 {
		t.Errorf("result = %+v; want Processed=1 Failed=1", result)
	}
}
