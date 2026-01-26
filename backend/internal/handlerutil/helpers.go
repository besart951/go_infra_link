package handlerutil

import (
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RespondError(c *gin.Context, status int, code, message string) {
	c.JSON(status, dto.ErrorResponse{
		Error:   code,
		Message: message,
	})
}

func RespondValidationError(c *gin.Context, fields map[string]string) {
	c.JSON(http.StatusBadRequest, dto.ErrorResponse{
		Error:  "validation_error",
		Fields: fields,
	})
}

func RespondNotFound(c *gin.Context, message string) {
	RespondError(c, http.StatusNotFound, "not_found", message)
}

func BindJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
		return false
	}
	return true
}

func BindQuery(c *gin.Context, dst any) bool {
	if err := c.ShouldBindQuery(dst); err != nil {
		RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
		return false
	}
	return true
}

func ParseUUIDParam(c *gin.Context, name string) (uuid.UUID, bool) {
	idStr := c.Param(name)
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "invalid_id", "Invalid UUID format")
		return uuid.Nil, false
	}
	return id, true
}

func ParseUUIDParamWithCode(c *gin.Context, name, code string) (uuid.UUID, bool) {
	idStr := c.Param(name)
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(c, http.StatusBadRequest, code, "Invalid UUID format")
		return uuid.Nil, false
	}
	return id, true
}
