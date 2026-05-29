package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Casagrande-Lucas/pipefy-integration/internal/adapter/inbound/http/handler"
	"github.com/Casagrande-Lucas/pipefy-integration/internal/adapter/inbound/http/middleware"
)

// NewRouter builds and returns a configured Gin engine.
// Call gin.SetMode before NewRouter if release mode is required.
func NewRouter(
	log *zap.Logger,
	clientHandler *handler.ClientHandler,
	webhookHandler *handler.WebhookHandler,
) *gin.Engine {
	r := gin.New() // no default middleware — we register our own below

	r.Use(middleware.Recovery(log))
	r.Use(middleware.Logger(log))

	// Health check — used by load balancers and container orchestrators.
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	{
		api.POST("/clients", clientHandler.Create)
		api.POST("/webhook", webhookHandler.Handle)
	}

	return r
}
