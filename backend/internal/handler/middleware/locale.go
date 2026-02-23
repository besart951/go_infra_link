package middleware

import (
	"strings"

	"github.com/besart951/go_infra_link/backend/pkg/i18n"
	"github.com/gin-gonic/gin"
)

const (
	// ContextLocaleKey is the key used to store the locale in the request context
	ContextLocaleKey = "locale"
	// ContextTranslatorKey is the key used to store the translator in the request context
	ContextTranslatorKey = "translator"
)

// LocaleMiddleware creates middleware that detects and sets the locale from the request.
// It checks the Accept-Language header and falls back to the default locale if not found.
func LocaleMiddleware(translator *i18n.Translator, defaultLocale string) gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := extractLocale(c.GetHeader("Accept-Language"), defaultLocale)

		c.Set(ContextLocaleKey, locale)
		c.Set(ContextTranslatorKey, translator)

		c.Next()
	}
}

// extractLocale parses the Accept-Language header and returns the best matching language.
// Falls back to defaultLocale if no match is found.
func extractLocale(acceptLanguage, defaultLocale string) string {
	if acceptLanguage == "" {
		return defaultLocale
	}

	// Parse Accept-Language header (e.g., "de-CH,de;q=0.9,en;q=0.8")
	// format: language[-region][;q=quality]
	parts := strings.Split(acceptLanguage, ",")
	part := strings.TrimSpace(parts[0])

	if part == "" {
		return defaultLocale
	}

	// Extract language-region part (before quality parameter)
	langPart := part
	if idx := strings.Index(part, ";"); idx != -1 {
		langPart = part[:idx]
	}

	langPart = strings.TrimSpace(langPart)

	// Convert "de-CH" to "de_CH", "de-ch" to "de_ch"
	locale := normalizeLanguageTag(langPart)

	// For now, we support all locales and let translator handle fallback
	if locale == "" {
		return defaultLocale
	}
	return locale
}

// normalizeLanguageTag converts language tags from RFC format to our internal format.
// Example: "de-CH" -> "de_CH", "de-ch" -> "de_CH"
func normalizeLanguageTag(tag string) string {
	parts := strings.Split(tag, "-")
	if len(parts) < 1 {
		return ""
	}

	lang := strings.ToLower(parts[0])
	if len(parts) > 1 {
		region := strings.ToUpper(parts[1])
		return lang + "_" + region
	}

	return lang
}

// GetLocale retrieves the locale from the request context.
// Falls back to "de_CH" if not found.
func GetLocale(c *gin.Context) string {
	if locale, ok := c.Get(ContextLocaleKey); ok {
		if localeStr, ok := locale.(string); ok {
			return localeStr
		}
	}
	return "de_CH"
}

// GetTranslator retrieves the translator from the request context.
func GetTranslator(c *gin.Context) *i18n.Translator {
	if tr, ok := c.Get(ContextTranslatorKey); ok {
		if translator, ok := tr.(*i18n.Translator); ok {
			return translator
		}
	}
	// This should not happen if middleware is properly configured
	return nil
}
