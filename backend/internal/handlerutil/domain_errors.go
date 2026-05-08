package handlerutil

import (
	"errors"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/gin-gonic/gin"
)

type ErrorSpec struct {
	Status    int
	Code      string
	Message   string
	Localized bool
}

type ErrorMapping struct {
	Target error
	Spec   ErrorSpec
}

func PlainError(status int, code, message string) ErrorSpec {
	return ErrorSpec{Status: status, Code: code, Message: message}
}

func LocalizedError(status int, code, key string) ErrorSpec {
	return ErrorSpec{Status: status, Code: code, Message: key, Localized: true}
}

func MapError(target error, spec ErrorSpec) ErrorMapping {
	return ErrorMapping{Target: target, Spec: spec}
}

func RespondMappedDomainError(c *gin.Context, err error, mappings ...ErrorMapping) bool {
	if err == nil {
		return false
	}

	if validationErr, ok := domain.AsValidationError(err); ok {
		RespondValidationError(c, validationErr.Fields)
		return true
	}

	for _, mapping := range mappings {
		if errors.Is(err, mapping.Target) {
			respondWithSpec(c, mapping.Spec)
			return true
		}
	}

	if mapping, ok := defaultDomainErrorMapping(err); ok {
		respondWithSpec(c, mapping.Spec)
		return true
	}

	return false
}

func RespondDomainError(c *gin.Context, err error, fallback ErrorSpec, mappings ...ErrorMapping) bool {
	if err == nil {
		return false
	}

	if RespondMappedDomainError(c, err, mappings...) {
		return true
	}

	respondWithSpec(c, fallback)
	return true
}

func respondWithSpec(c *gin.Context, spec ErrorSpec) {
	if spec.Localized {
		RespondLocalizedError(c, spec.Status, spec.Code, spec.Message)
		return
	}

	RespondError(c, spec.Status, spec.Code, spec.Message)
}

func defaultDomainErrorMapping(err error) (ErrorMapping, bool) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return MapError(domain.ErrNotFound, LocalizedError(http.StatusNotFound, "not_found", "errors.not_found")), true
	case errors.Is(err, domain.ErrConflict):
		return MapError(domain.ErrConflict, LocalizedError(http.StatusConflict, "conflict", "errors.conflict")), true
	case errors.Is(err, domain.ErrInvalidArgument):
		return MapError(domain.ErrInvalidArgument, LocalizedError(http.StatusBadRequest, "validation_error", "errors.validation_error")), true
	case errors.Is(err, domainUser.ErrForbiddenUserDirectory):
		return MapError(domainUser.ErrForbiddenUserDirectory, LocalizedError(http.StatusForbidden, "forbidden", "errors.forbidden")), true
	default:
		return ErrorMapping{}, false
	}
}
