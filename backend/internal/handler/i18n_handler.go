package handler

import (
	"net/http"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/besart951/go_infra_link/backend/pkg/i18n"
	"github.com/gin-gonic/gin"
)

type I18nHandler struct {
	loader *i18n.Loader
}

func NewI18nHandler(loader *i18n.Loader) *I18nHandler {
	return &I18nHandler{
		loader: loader,
	}
}

// GetTranslations godoc
// @Summary Get translations for a specific locale
// @Tags i18n
// @Produce json
// @Param locale path string true "Locale code (e.g., de_CH, en_US)"
// @Success 200 {object} map[string]interface{} "Translation data"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/i18n/{locale} [get]
func (h *I18nHandler) GetTranslations(c *gin.Context) {
	locale := c.Param("locale")

	// Validate locale format
	if locale == "" || len(locale) > 10 {
		handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "invalid_locale", "errors.invalid_locale")
		return
	}

	// Sanitize locale (allow only alphanumeric and underscore)
	if !isValidLocale(locale) {
		handlerutil.RespondLocalizedError(c, http.StatusBadRequest, "invalid_locale", "errors.invalid_locale")
		return
	}

	// Load translation file
	filename := locale + ".json"
	translations, err := h.loader.Load(filename)
	if err != nil {
		// Check if file not found
		if strings.Contains(err.Error(), "no such file") {
			handlerutil.RespondLocalizedError(c, http.StatusNotFound, "locale_not_found", "errors.locale_not_found")
			return
		}

		handlerutil.RespondLocalizedError(c, http.StatusInternalServerError, "failed_to_load_translations", "errors.internal_server_error")
		return
	}

	c.JSON(http.StatusOK, translations)
}

// isValidLocale checks if the locale string is valid (alphanumeric + underscore only)
func isValidLocale(locale string) bool {
	for _, r := range locale {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}
	return true
}
