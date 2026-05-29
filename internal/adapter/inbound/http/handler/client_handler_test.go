package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Casagrande-Lucas/pipefy-integration/internal/adapter/inbound/http/dto"
	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/entity"
	"github.com/Casagrande-Lucas/pipefy-integration/internal/usecase/client"
	"github.com/Casagrande-Lucas/pipefy-integration/pkg/apperror"
)

func init() { gin.SetMode(gin.TestMode) }

// --- stub ---

type stubCreateClientUC struct {
	result *entity.Client
	err    error
}

func (s *stubCreateClientUC) Execute(client.CreateClientInput) (*entity.Client, error) {
	return s.result, s.err
}

// --- helpers ---

func newClientRouter(uc createClientUseCase) *gin.Engine {
	r := gin.New()
	r.POST("/clients", NewClientHandler(uc, zap.NewNop()).Create)
	return r
}

func buildClient(t *testing.T) *entity.Client {
	t.Helper()
	c, err := entity.NewClient("Lucas", "lucas@example.com", "consultoria", 150_000)
	if err != nil {
		t.Fatalf("buildClient: %v", err)
	}
	return c
}

func postJSON(t *testing.T, r *gin.Engine, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

// --- tests ---

func TestClientHandler_Create_Success(t *testing.T) {
	t.Parallel()

	r := newClientRouter(&stubCreateClientUC{result: buildClient(t)})
	body := `{"cliente_nome":"Lucas","cliente_email":"lucas@example.com","tipo_solicitacao":"consultoria","valor_patrimonio":150000}`
	w := postJSON(t, r, "/clients", body)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d; want 201", w.Code)
	}

	var resp dto.ClientResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Email != "lucas@example.com" {
		t.Errorf("email = %q; want %q", resp.Email, "lucas@example.com")
	}
	if resp.ID == "" {
		t.Error("id must not be empty")
	}
}

func TestClientHandler_Create_InvalidJSON(t *testing.T) {
	t.Parallel()

	w := postJSON(t, newClientRouter(&stubCreateClientUC{}), "/clients", `{invalid}`)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want 400", w.Code)
	}
}

func TestClientHandler_Create_MissingRequiredField(t *testing.T) {
	t.Parallel()

	// cliente_nome is missing — binding:"required" should reject it.
	body := `{"cliente_email":"a@b.com","tipo_solicitacao":"consulting","valor_patrimonio":0}`
	w := postJSON(t, newClientRouter(&stubCreateClientUC{}), "/clients", body)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want 400", w.Code)
	}
}

func TestClientHandler_Create_UseCaseErrors(t *testing.T) {
	t.Parallel()

	validBody := `{"cliente_nome":"X","cliente_email":"x@x.com","tipo_solicitacao":"c","valor_patrimonio":0}`

	cases := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{
			name:       "validation error → 400",
			err:        apperror.NewValidation("email is invalid", nil),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "not found → 404",
			err:        apperror.NewNotFound("client", nil),
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "internal error → 500",
			err:        apperror.NewInternal(nil),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := newClientRouter(&stubCreateClientUC{err: tc.err})
			w := postJSON(t, r, "/clients", validBody)

			if w.Code != tc.wantStatus {
				t.Errorf("status = %d; want %d", w.Code, tc.wantStatus)
			}

			var resp dto.ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("decode error response: %v", err)
			}
			if resp.Code != tc.wantStatus {
				t.Errorf("body.code = %d; want %d", resp.Code, tc.wantStatus)
			}
		})
	}
}
