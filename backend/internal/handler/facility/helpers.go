package facility

import (
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func respondError(c *gin.Context, status int, code, message string) {
	c.JSON(status, dto.ErrorResponse{
		Error:   code,
		Message: message,
	})
}

func respondValidationError(c *gin.Context, fields map[string]string) {
	c.JSON(http.StatusBadRequest, dto.ErrorResponse{
		Error:  "validation_error",
		Fields: fields,
	})
}

func respondNotFound(c *gin.Context, message string) {
	respondError(c, http.StatusNotFound, "not_found", message)
}

func bindJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		respondError(c, http.StatusBadRequest, "validation_error", err.Error())
		return false
	}
	return true
}

func bindQuery(c *gin.Context, dst any) bool {
	if err := c.ShouldBindQuery(dst); err != nil {
		respondError(c, http.StatusBadRequest, "validation_error", err.Error())
		return false
	}
	return true
}

func parseUUIDParam(c *gin.Context, name string) (uuid.UUID, bool) {
	idStr := c.Param(name)
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid_id", "Invalid UUID format")
		return uuid.Nil, false
	}
	return id, true
}
