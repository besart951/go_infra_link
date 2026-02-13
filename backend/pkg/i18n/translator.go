package i18n

import (
	"fmt"
	"strings"
)

// Translator provides localization services.
// It loads and manages translations for different languages and provides
// convenient methods to retrieve translated strings.
type Translator struct {
	translations map[string]map[string]string // language -> key -> message
	defaultLang  string
}

// NewTranslator creates a new Translator instance.
// defaultLang is the fallback language if a translation is not found.
func NewTranslator(defaultLang string) *Translator {
	return &Translator{
		translations: make(map[string]map[string]string),
		defaultLang:  defaultLang,
	}
}

// LoadLanguage adds translations for a specific language.
// translations should be a nested map where first level keys are namespaces
// (e.g., "auth", "validation") and second level keys are message keys.
func (t *Translator) LoadLanguage(lang string, translations map[string]interface{}) error {
	if t.translations[lang] == nil {
		t.translations[lang] = make(map[string]string)
	}

	// Flatten nested maps into dot-separated keys
	flatten(translations, "", t.translations[lang])
	return nil
}

// Get retrieves a translated string for the given key and language.
// If the translation is not found, it returns the key itself as fallback.
// Falls back to defaultLang if language-specific translation is not found.
func (t *Translator) Get(lang, key string) string {
	// Try the requested language
	if msgs, ok := t.translations[lang]; ok {
		if msg, ok := msgs[key]; ok {
			return msg
		}
	}

	// Fallback to default language
	if lang != t.defaultLang {
		if msgs, ok := t.translations[t.defaultLang]; ok {
			if msg, ok := msgs[key]; ok {
				return msg
			}
		}
	}

	// Return key as fallback
	return key
}

// GetWithParams retrieves a translated string and substitutes parameters.
// Parameters are specified as {paramName} in the translation string.
// Example: translator.GetWithParams("de_CH", "errors.welcome", map[string]string{"name": "Max"})
func (t *Translator) GetWithParams(lang, key string, params map[string]string) string {
	msg := t.Get(lang, key)
	for name, value := range params {
		msg = strings.ReplaceAll(msg, "{"+name+"}", value)
	}
	return msg
}

// GetPlural retrieves a pluralized translation based on count.
// The translation key should follow pattern: "key.zero", "key.one", "key.other"
func (t *Translator) GetPlural(lang, key string, count int) string {
	pluralKey := getPluralKey(key, count)
	return t.Get(lang, pluralKey)
}

// flatten converts nested maps into dot-separated keys
func flatten(data map[string]interface{}, prefix string, result map[string]string) {
	for k, v := range data {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}

		switch val := v.(type) {
		case map[string]interface{}:
			flatten(val, key, result)
		case string:
			result[key] = val
		default:
			result[key] = fmt.Sprintf("%v", val)
		}
	}
}

// getPluralKey returns the appropriate plural form key
func getPluralKey(key string, count int) string {
	if count == 0 {
		return key + ".zero"
	}
	if count == 1 {
		return key + ".one"
	}
	return key + ".other"
}
