package handlerutil

import (
	"net/http"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// RespondLocalizedError responds with a localized error message.
// The translator and locale are automatically retrieved from the request context.
func RespondLocalizedError(c *gin.Context, status int, code string, keyOrMessage string) {
	translator := middleware.GetTranslator(c)
	locale := middleware.GetLocale(c)

	message := keyOrMessage
	if translator != nil {
		// Try to translate the error code first
		translated := translator.Get(locale, code)
		if translated != code { // If translation exists (not the key itself)
			message = translated
		} else {
			// Try the full message key path (e.g., "errors.not_found")
			translated = translator.Get(locale, "errors."+code)
			if translated != "errors."+code {
				message = translated
			}
		}
	}

	RespondError(c, status, code, message)
}

// RespondLocalizedValidationError responds with localized validation errors.
// It translates field names and validation messages.
func RespondLocalizedValidationError(c *gin.Context, fields map[string]string) {
	translator := middleware.GetTranslator(c)
	locale := middleware.GetLocale(c)

	translatedFields := make(map[string]string)
	for field, message := range fields {
		// Try to translate the validation message
		translatedMsg := message
		if translator != nil {
			// Try to find a localized message based on the validation rule
			translated := translator.Get(locale, message)
			if translated != message {
				translatedMsg = translated
			}
		}
		translatedFields[field] = translatedMsg
	}

	c.JSON(http.StatusBadRequest, dto.ErrorResponse{
		Error:  "validation_error",
		Fields: translatedFields,
	})
}

// RespondLocalizedNotFound responds with a localized "not found" error.
func RespondLocalizedNotFound(c *gin.Context) {
	translator := middleware.GetTranslator(c)
	locale := middleware.GetLocale(c)

	message := "not found"
	if translator != nil {
		message = translator.Get(locale, "errors.not_found")
	}

	RespondLocalizedError(c, http.StatusNotFound, "not_found", message)
}

// RespondLocalizedValidationFieldError responds with a localized single field validation error.
// It translates the field name and the validation message.
// For example: RespondLocalizedValidationFieldError(c, "email", "errors.email_invalid")
func RespondLocalizedValidationFieldError(c *gin.Context, field, messageKey string) {
	translator := middleware.GetTranslator(c)
	locale := middleware.GetLocale(c)

	message := messageKey
	if translator != nil {
		message = translator.Get(locale, messageKey)
	}

	fields := map[string]string{field: message}
	RespondLocalizedValidationError(c, fields)
}

// LocalizeValidationErrors translates validator.FieldError to localized messages.
// It uses the validator tag and field structure to generate appropriate keys.
func LocalizeValidationErrors(c *gin.Context, dst any, verr validator.ValidationErrors) map[string]string {
	translator := middleware.GetTranslator(c)
	locale := middleware.GetLocale(c)

	fields := make(map[string]string, len(verr))
	fieldNames := structJSONFieldMap(dst)

	for _, fe := range verr {
		fieldName := fe.StructField()
		if mapped, ok := fieldNames[fieldName]; ok && mapped != "" {
			fieldName = mapped
		}

		// Generate localization key based on validation tag
		// Example: "validation.min_length" for min tag
		message := generateValidationMessageKey(fe)

		if translator != nil {
			translated := translator.Get(locale, message)
			if translated != message {
				// If we have a parameterized message, substitute parameters
				if strings.Contains(translated, "{field}") {
					translated = strings.ReplaceAll(translated, "{field}", fieldName)
				}
				if strings.Contains(translated, "{min}") {
					translated = strings.ReplaceAll(translated, "{min}", fe.Param())
				}
				if strings.Contains(translated, "{max}") {
					translated = strings.ReplaceAll(translated, "{max}", fe.Param())
				}
				fields[fieldName] = translated
			} else {
				// Fallback to non-localized message
				fields[fieldName] = validationMessage(fe)
			}
		} else {
			fields[fieldName] = validationMessage(fe)
		}
	}

	return fields
}

// generateValidationMessageKey generates a localization key for a validation error.
func generateValidationMessageKey(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "validation.required"
	case "email":
		return "validation.email_invalid"
	case "min":
		return "validation.min_length"
	case "max":
		return "validation.max_length"
	case "len":
		return "validation.exact_length"
	case "oneof":
		return "validation.oneof"
	case "numeric":
		return "validation.numeric"
	case "alphanum":
		return "validation.alphanumeric"
	case "unique":
		return "validation.unique"
	case "datetime":
		return "validation.date_invalid"
	default:
		return "validation." + fe.Tag()
	}
}
