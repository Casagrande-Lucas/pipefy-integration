package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Casagrande-Lucas/pipefy-integration/internal/adapter/inbound/http/dto"
)

// Recovery returns a Gin middleware that catches panics, logs them with a
// stack trace via Zap, and responds with 500 Internal Server Error.
// It replaces gin.Recovery() to keep all logging through the Zap instance.
func Recovery(log *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, recovered any) {
		log.Error("panic recovered",
			zap.Any("panic", recovered),
			zap.String("path", c.FullPath()),
			zap.String("method", c.Request.Method),
		)
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			dto.NewErrorResponse(http.StatusInternalServerError, "internal server error"),
		)
	})
}
