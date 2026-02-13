// This is an EXAMPLE file showing how to integrate localized messages into handlers.
// Copy patterns from this file into your actual handlers.

package handler

import (
	"errors"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// EXAMPLE_LocalizationPatterns demonstrates common localization patterns
type EXAMPLE_LocalizationPatterns struct{}

// Pattern 1: Simple Not Found Error with Localization
// This shows how to respond with a localized "not found" message
func (h *EXAMPLE_LocalizationPatterns) Pattern1_NotFound(c *gin.Context) {
	// Get id from request...
	// Try to find resource...

	// Respond with localized error message
	// This automatically translates "errors.not_found" from the translation file
	handlerutil.RespondLocalizedNotFound(c)
}

// Pattern 2: Domain Error Mapping with Localization
// This shows how to map domain errors to localized messages
func (h *EXAMPLE_LocalizationPatterns) Pattern2_DomainErrors(c *gin.Context) {
	// var req = ... read from request
	// service.DoSomething() -> may return domain.ErrNotFound, domain.ErrConflict, etc.
	// err := h.service.DoSomething(req)

	var exampleErr error // This would come from your service

	if exampleErr != nil {
		if errors.Is(exampleErr, domain.ErrNotFound) {
			handlerutil.RespondLocalizedNotFound(c)
			return
		}

		if errors.Is(exampleErr, domain.ErrConflict) {
			handlerutil.RespondLocalizedError(
				c,
				http.StatusConflict,
				"conflict",
				"A resource with this value already exists",
			)
			return
		}

		// Generic internal error
		handlerutil.RespondLocalizedError(
			c,
			http.StatusInternalServerError,
			"internal_error",
			"errors.internal_server_error",
		)
		return
	}
}

// Pattern 3: Validation Error with Automatic Translation
// This shows how to automatically translate validation errors
func (h *EXAMPLE_LocalizationPatterns) Pattern3_AutomaticValidation(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
		Name     string `json:"name" binding:"required,max=100"`
	}

	// BindJSON automatically translates validation errors
	if !handlerutil.BindJSON(c, &req) {
		return
	}

	// If validation passed, continue with business logic...
}

// Pattern 4: Manual Field Validation with Localization
// This shows how to validate and localize custom validation logic
func (h *EXAMPLE_LocalizationPatterns) Pattern4_ManualValidation(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErr, ok := err.(validator.ValidationErrors); ok {
			// Automatically translates validation errors
			fields := handlerutil.LocalizeValidationErrors(c, &req, validationErr)
			handlerutil.RespondLocalizedValidationError(c, fields)
			return
		}

		handlerutil.RespondLocalizedError(
			c,
			http.StatusBadRequest,
			"invalid_request",
			"errors.bad_request",
		)
		return
	}

	// Custom validation (not from decoder)
	translator := middleware.GetTranslator(c)
	locale := middleware.GetLocale(c)

	if len(req.Password) < 8 {
		message := translator.Get(locale, "validation.password_too_short")
		fields := map[string]string{"password": message}
		handlerutil.RespondLocalizedValidationError(c, fields)
		return
	}

	// If validation passed...
}

// Pattern 5: Message with Parameters
// This shows how to use parameter substitution in messages
func (h *EXAMPLE_LocalizationPatterns) Pattern5_ParameterSubstitution(c *gin.Context) {
	translator := middleware.GetTranslator(c)
	locale := middleware.GetLocale(c)

	// Example 1: Single field validation with parameter
	message := translator.GetWithParams(
		locale,
		"validation.min_length",
		map[string]string{
			"field": "Passwort",
			"min":   "8",
		},
	)
	// Result: "Passwort muss mindestens 8 Zeichen lang sein."

	fields := map[string]string{"password": message}
	handlerutil.RespondLocalizedValidationError(c, fields)
}

// Pattern 6: Business Logic Message with Localization
// This shows how to return success or business logic messages
func (h *EXAMPLE_LocalizationPatterns) Pattern6_BusinessLogicMessage(c *gin.Context) {
	translator := middleware.GetTranslator(c)
	locale := middleware.GetLocale(c)

	// Service successfully deleted a user
	message := translator.Get(locale, "user.user_deleted")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
	})
}

// Pattern 7: Complex Error with Multiple Fields
// This shows how to handle multiple field validation errors
func (h *EXAMPLE_LocalizationPatterns) Pattern7_MultipleFieldErrors(c *gin.Context) {
	var req struct {
		Email     string `json:"email" binding:"required,email"`
		Password  string `json:"password" binding:"required,min=8"`
		FirstName string `json:"first_name" binding:"required"`
		LastName  string `json:"last_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErr, ok := err.(validator.ValidationErrors); ok {
			fields := handlerutil.LocalizeValidationErrors(c, &req, validationErr)
			handlerutil.RespondLocalizedValidationError(c, fields)
			return
		}
	}
}

// Pattern 8: Conditional Error Based on Business Logic
// This shows how to determine error codes based on service response
func (h *EXAMPLE_LocalizationPatterns) Pattern8_ConditionalError(c *gin.Context) {
	translator := middleware.GetTranslator(c)
	locale := middleware.GetLocale(c)

	// var req = ...
	// err := h.service.CreateUser(req)

	var exampleErr error // From service

	if exampleErr != nil {
		message := translator.Get(locale, "errors.internal_server_error")

		// Map specific service errors to localized messages
		if exampleErr.Error() == "email_already_registered" {
			message = translator.Get(locale, "auth.email_already_registered")
		} else if exampleErr.Error() == "weak_password" {
			message = translator.Get(locale, "auth.weak_password")
		}

		handlerutil.RespondError(
			c,
			http.StatusBadRequest,
			"validation_error",
			message,
		)
		return
	}
}

// Pattern 9: Direct Translator Access
// For advanced use cases where you need direct translator access
func (h *EXAMPLE_LocalizationPatterns) Pattern9_DirectTranslator(c *gin.Context) {
	translator := middleware.GetTranslator(c)
	if translator == nil {
		handlerutil.RespondLocalizedError(
			c,
			http.StatusInternalServerError,
			"internal_error",
			"Translator not available",
		)
		return
	}

	locale := middleware.GetLocale(c)
	// locale = "de_CH" or what the user requested

	// Get a simple translation
	message := translator.Get(locale, "common.success")

	// Get a plural form
	count := 5
	itemMessage := translator.GetPlural(locale, "items.count", count)

	// Get with parameters
	paramMessage := translator.GetWithParams(
		locale,
		"validation.required",
		map[string]string{"field": "Email"},
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"items":   itemMessage,
		"param":   paramMessage,
	})
}

// Pattern 10: Creating a Helper Function for Your Feature
// This shows how to create domain-specific helper functions
func (h *EXAMPLE_LocalizationPatterns) respondUserNotFound(c *gin.Context) {
	translator := middleware.GetTranslator(c)
	locale := middleware.GetLocale(c)

	message := translator.Get(locale, "user.user_not_found")
	if message == "user.user_not_found" {
		// Fallback if translation not found
		message = translator.Get(locale, "errors.not_found")
	}

	handlerutil.RespondError(c, http.StatusNotFound, "user_not_found", message)
}

func (h *EXAMPLE_LocalizationPatterns) respondTeamNameRequired(c *gin.Context) {
	translator := middleware.GetTranslator(c)
	locale := middleware.GetLocale(c)

	message := translator.Get(locale, "validation.required")
	if message == "validation.required" {
		message = "Team name is required"
	}

	fields := map[string]string{"name": message}
	handlerutil.RespondLocalizedValidationError(c, fields)
}

// Pattern 11: List Response with Localization
// This shows how to handle empty/paginated results with localization
func (h *EXAMPLE_LocalizationPatterns) Pattern11_ListResponse(c *gin.Context) {
	translator := middleware.GetTranslator(c)
	locale := middleware.GetLocale(c)

	// Simulate getting items from service
	var items []dto.ProjectResponse
	items = []dto.ProjectResponse{} // Empty result

	if len(items) == 0 {
		// Return with localized message
		message := translator.Get(locale, "pagination.no_results")
		c.JSON(http.StatusOK, gin.H{
			"items":   items,
			"total":   0,
			"message": message,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"total": len(items),
	})
}

// Pattern 12: Update/Delete Success Response
// Shows how to respond with localized success messages
func (h *EXAMPLE_LocalizationPatterns) Pattern12_SuccessResponse(c *gin.Context) {
	translator := middleware.GetTranslator(c)
	locale := middleware.GetLocale(c)

	// Determine which resource was deleted/updated
	resourceType := "user" // From route or request

	successKey := resourceType + ".user_updated"
	message := translator.Get(locale, successKey)

	if message == successKey {
		// Fallback to generic message
		message = translator.Get(locale, "common.success")
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
	})
}
