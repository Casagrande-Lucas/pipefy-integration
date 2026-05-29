package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Casagrande-Lucas/pipefy-integration/internal/adapter/inbound/http/dto"
	"github.com/Casagrande-Lucas/pipefy-integration/internal/domain/entity"
	"github.com/Casagrande-Lucas/pipefy-integration/internal/usecase/client"
	"github.com/Casagrande-Lucas/pipefy-integration/pkg/apperror"
)

// createClientUseCase is the port this handler depends on.
// Accepting an interface rather than the concrete type keeps the handler
// decoupled and straightforwardly testable.
type createClientUseCase interface {
	Execute(input client.CreateClientInput) (*entity.Client, error)
}

// ClientHandler handles HTTP requests for the client resource.
type ClientHandler struct {
	createClient createClientUseCase
	log          *zap.Logger
}

// NewClientHandler returns a ClientHandler with the given dependencies.
func NewClientHandler(uc createClientUseCase, log *zap.Logger) *ClientHandler {
	return &ClientHandler{createClient: uc, log: log}
}

// Create handles POST /clients.
//
//	@Summary      Create a client
//	@Description  Registers a new client and enqueues a Pipefy createCard event.
//	@Tags         clients
//	@Accept       json
//	@Produce      json
//	@Param        body  body      dto.CreateClientRequest  true  "Client data"
//	@Success      201   {object}  dto.ClientResponse
//	@Failure      400   {object}  dto.ErrorResponse
//	@Failure      500   {object}  dto.ErrorResponse
//	@Router       /clients [post]
func (h *ClientHandler) Create(c *gin.Context) {
	var req dto.CreateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	input := client.CreateClientInput{
		Name:           req.Nome,
		Email:          req.Email,
		RequestType:    req.TipoSolicitacao,
		PatrimonyValue: req.ValorPatrimonio,
	}

	result, err := h.createClient.Execute(input)
	if err != nil {
		h.log.Warn("create client failed", zap.Error(err))
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.NewClientResponse(result))
}

// respondError maps an AppError to the appropriate HTTP status and writes
// an ErrorResponse. Unknown errors fall back to 500.
func respondError(c *gin.Context, err error) {
	if appErr, ok := errors.AsType[*apperror.AppError](err); ok {
		c.JSON(appErr.HTTPStatus, dto.NewErrorResponse(appErr.HTTPStatus, appErr.Message))
		return
	}
	c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
		http.StatusInternalServerError, "internal server error",
	))
}
