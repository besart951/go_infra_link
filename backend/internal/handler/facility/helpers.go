package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func respondError(c *gin.Context, status int, code, message string) {
	handlerutil.RespondError(c, status, code, message)
}

func respondValidationError(c *gin.Context, fields map[string]string) {
	handlerutil.RespondValidationError(c, fields)
}

func respondNotFound(c *gin.Context, message string) {
	handlerutil.RespondNotFound(c, message)
}

func bindJSON(c *gin.Context, dst any) bool {
	return handlerutil.BindJSON(c, dst)
}

func bindQuery(c *gin.Context, dst any) bool {
	return handlerutil.BindQuery(c, dst)
}

func parseUUIDParam(c *gin.Context, name string) (uuid.UUID, bool) {
	return handlerutil.ParseUUIDParam(c, name)
}
