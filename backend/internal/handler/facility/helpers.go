package facility

import (
	"errors"
	"net/http"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func respondError(c *gin.Context, status int, code, message string) {
	handlerutil.RespondError(c, status, code, message)
}

func respondLocalizedError(c *gin.Context, status int, code, translationKey string) {
	handlerutil.RespondLocalizedError(c, status, code, translationKey)
}

func respondValidationError(c *gin.Context, fields map[string]string) {
	handlerutil.RespondValidationError(c, fields)
}

func respondNotFound(c *gin.Context, message string) {
	handlerutil.RespondNotFound(c, message)
}

func respondLocalizedNotFoundError(c *gin.Context, translationKey string) {
	handlerutil.RespondLocalizedError(c, http.StatusNotFound, "not_found", translationKey)
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

func parsePaginationQuery(c *gin.Context) (dto.PaginationQuery, bool) {
	var query dto.PaginationQuery
	if !bindQuery(c, &query) {
		return dto.PaginationQuery{}, false
	}
	return query, true
}

func parseUUIDQueryParam(c *gin.Context, name string) (*uuid.UUID, bool) {
	value := strings.TrimSpace(c.Query(name))
	if value == "" {
		return nil, true
	}
	parsed, err := uuid.Parse(value)
	if err != nil {
		respondInvalidArgument(c, name+" is invalid")
		return nil, false
	}
	return &parsed, true
}

func respondValidationOrError(c *gin.Context, err error, fallbackCode string) bool {
	if err == nil {
		return false
	}
	if ve, ok := domain.AsValidationError(err); ok {
		respondValidationError(c, ve.Fields)
		return true
	}
	respondError(c, http.StatusInternalServerError, fallbackCode, err.Error())
	return true
}

func respondLocalizedValidationOrError(c *gin.Context, err error, translationKey string) bool {
	if err == nil {
		return false
	}
	if ve, ok := domain.AsValidationError(err); ok {
		respondValidationError(c, ve.Fields)
		return true
	}
	respondLocalizedError(c, http.StatusInternalServerError, "error", translationKey)
	return true
}

func respondInvalidReference(c *gin.Context) {
	respondLocalizedError(c, http.StatusBadRequest, "invalid_reference", "facility.invalid_reference")
}

func respondLocalizedConflict(c *gin.Context, translationKey string) {
	respondLocalizedError(c, http.StatusConflict, "conflict", translationKey)
}

func respondInvalidArgument(c *gin.Context, message string) {
	respondError(c, http.StatusBadRequest, "validation_error", message)
}

func respondLocalizedInvalidArgument(c *gin.Context, translationKey string) {
	respondLocalizedError(c, http.StatusBadRequest, "validation_error", translationKey)
}

func respondLocalizedNotFoundIf(c *gin.Context, err error, translationKey string) bool {
	if errors.Is(err, domain.ErrNotFound) {
		respondLocalizedNotFoundError(c, translationKey)
		return true
	}
	return false
}

func parseUUIDString(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
