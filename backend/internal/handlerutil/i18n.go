package handlerutil

import (
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/gin-gonic/gin"
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
