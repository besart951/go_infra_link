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
		for _, key := range []string{keyOrMessage, code, "errors." + code} {
			if key == "" {
				continue
			}
			translated := translator.Get(locale, key)
			if translated != key {
				message = translated
				break
			}
		}
	}

	RespondError(c, status, code, message)
}
