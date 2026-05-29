package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Casagrande-Lucas/pipefy-integration/internal/usecase/webhook"
	"github.com/Casagrande-Lucas/pipefy-integration/pkg/apperror"
)

// --- stub ---

type stubProcessWebhookUC struct {
	err           error
	capturedInput webhook.ProcessWebhookInput
}

func (s *stubProcessWebhookUC) Execute(input webhook.ProcessWebhookInput) error {
	s.capturedInput = input
	return s.err
}

// --- helpers ---

func newWebhookRouter(uc processWebhookUseCase) (*gin.Engine, *stubProcessWebhookUC) {
	stub, _ := uc.(*stubProcessWebhookUC)
	r := gin.New()
	r.POST("/webhook", NewWebhookHandler(uc, zap.NewNop()).Handle)
	return r, stub
}

const validWebhookBody = `{
	"action": "card.field.update",
	"data": {
		"card": {"id": "card-42", "pipe_id": "pipe-1"},
		"field": {"field_id": "status", "new_value": "processado"}
	}
}`

func postWebhook(t *testing.T, r *gin.Engine, body string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	r.ServeHTTP(w, req)
	return w
}

// --- tests ---

func TestWebhookHandler_Handle_Success_WithHeader(t *testing.T) {
	t.Parallel()

	stub := &stubProcessWebhookUC{}
	r, _ := newWebhookRouter(stub)
	w := postWebhook(t, r, validWebhookBody, map[string]string{
		pipefyWebhookUUIDHeader: "evt-uuid-001",
	})

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d; want 204", w.Code)
	}
	if stub.capturedInput.EventID != "evt-uuid-001" {
		t.Errorf("event_id = %q; want %q", stub.capturedInput.EventID, "evt-uuid-001")
	}
	if stub.capturedInput.PipefyCardID != "card-42" {
		t.Errorf("card_id = %q; want %q", stub.capturedInput.PipefyCardID, "card-42")
	}
}

func TestWebhookHandler_Handle_Success_DerivedEventID(t *testing.T) {
	t.Parallel()

	stub := &stubProcessWebhookUC{}
	r, _ := newWebhookRouter(stub)
	w := postWebhook(t, r, validWebhookBody, nil)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d; want 204", w.Code)
	}
	// Header absent → deterministic key: action:card_id:field_id
	wantEventID := "card.field.update:card-42:status"
	if stub.capturedInput.EventID != wantEventID {
		t.Errorf("event_id = %q; want %q", stub.capturedInput.EventID, wantEventID)
	}
}

func TestWebhookHandler_Handle_InvalidJSON(t *testing.T) {
	t.Parallel()

	r, _ := newWebhookRouter(&stubProcessWebhookUC{})
	w := postWebhook(t, r, `{bad json}`, nil)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want 400", w.Code)
	}
}

func TestWebhookHandler_Handle_UseCaseErrors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{
			name:       "validation → 400",
			err:        apperror.NewValidation("invalid status", nil),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "not found → 404",
			err:        apperror.NewNotFound("client", nil),
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "internal → 500",
			err:        apperror.NewInternal(nil),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r, _ := newWebhookRouter(&stubProcessWebhookUC{err: tc.err})
			w := postWebhook(t, r, validWebhookBody, nil)

			if w.Code != tc.wantStatus {
				t.Errorf("status = %d; want %d", w.Code, tc.wantStatus)
			}
		})
	}
}
