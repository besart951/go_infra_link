package handlerutil

import (
	"net/http"
	"sort"

	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/common"
	"github.com/besart951/go_infra_link/backend/internal/requestutil"
	"github.com/gin-gonic/gin"
)

func RespondError(c *gin.Context, status int, code, message string) {
	respondErrorResponse(c, status, dto.ErrorResponse{
		Error:   code,
		Code:    code,
		Message: message,
	})
}

func RespondErrorWithLocalizedKey(c *gin.Context, status int, code, message, localizedKey string) {
	respondErrorResponse(c, status, dto.ErrorResponse{
		Error:        code,
		Code:         code,
		Message:      message,
		LocalizedKey: localizedKey,
	})
}

func RespondErrorWithDetails(c *gin.Context, status int, code, message, localizedKey string, details any) {
	respondErrorResponse(c, status, dto.ErrorResponse{
		Error:        code,
		Code:         code,
		Message:      message,
		LocalizedKey: localizedKey,
		Details:      details,
	})
}

func RespondValidationError(c *gin.Context, fields map[string]string) {
	RespondValidationErrorWithStatus(c, http.StatusBadRequest, fields)
}

func RespondValidationErrorWithStatus(c *gin.Context, status int, fields map[string]string) {
	RespondValidationErrorWithFieldErrors(c, status, fields, fieldErrorsFromMap(fields))
}

func RespondValidationErrorWithFieldErrors(c *gin.Context, status int, fields map[string]string, fieldErrors []dto.FieldErrorResponse) {
	respondErrorResponse(c, status, dto.ErrorResponse{
		Error:        "validation_error",
		Code:         "validation_error",
		Message:      "validation failed",
		LocalizedKey: "errors.validation_error",
		Fields:       cloneFields(fields),
		FieldErrors:  fieldErrors,
	})
}

func respondErrorResponse(c *gin.Context, status int, response dto.ErrorResponse) {
	if c == nil {
		return
	}
	if c.Request != nil && requestutil.IsRequestCanceled(c.Request.Context()) {
		c.Abort()
		return
	}
	if c.Writer != nil && c.Writer.Written() {
		return
	}

	if response.Code == "" {
		response.Code = response.Error
	}
	if response.Error == "" {
		response.Error = response.Code
	}
	response.RequestID = requestID(c)
	c.AbortWithStatusJSON(status, response)
}

func fieldErrorsFromMap(fields map[string]string) []dto.FieldErrorResponse {
	if len(fields) == 0 {
		return nil
	}
	paths := make([]string, 0, len(fields))
	for path := range fields {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	fieldErrors := make([]dto.FieldErrorResponse, 0, len(paths))
	for _, path := range paths {
		message := fields[path]
		fieldErrors = append(fieldErrors, dto.FieldErrorResponse{
			Path:    path,
			Code:    "validation_error",
			Message: message,
		})
	}
	return fieldErrors
}

func cloneFields(fields map[string]string) map[string]string {
	if len(fields) == 0 {
		return nil
	}
	clone := make(map[string]string, len(fields))
	for key, value := range fields {
		clone[key] = value
	}
	return clone
}

func requestID(c *gin.Context) string {
	if c == nil {
		return ""
	}
	for _, key := range []string{"request_id", "requestID", "RequestID"} {
		if value, ok := c.Get(key); ok {
			if text, ok := value.(string); ok && text != "" {
				return text
			}
		}
	}
	if c.Request == nil {
		return ""
	}
	for _, header := range []string{"X-Request-ID", "X-Request-Id", "X-Correlation-ID"} {
		if value := c.Request.Header.Get(header); value != "" {
			return value
		}
	}
	return ""
}
