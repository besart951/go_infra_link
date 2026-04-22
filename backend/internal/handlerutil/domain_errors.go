package handlerutil

import (
	"errors"

	"github.com/besart951/go_infra_link/backend/internal/domain"
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
