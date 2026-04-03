package app

import (
	"github.com/besart951/go_infra_link/backend/pkg/i18n"
	applogger "github.com/besart951/go_infra_link/backend/pkg/logger"
)

const defaultLocale = "de_CH"

func initializeTranslator(log applogger.Logger) (*i18n.Translator, *i18n.Loader) {
	translator := i18n.NewTranslator(defaultLocale)
	loader := i18n.NewLoader("./translations")

	allTranslations, err := loader.LoadAll()
	if err != nil {
		log.Error("Failed to load translations", "err", err)
		return translator, loader
	}

	for lang, translations := range allTranslations {
		if err := translator.LoadLanguage(lang, translations); err != nil {
			log.Error("Failed to load language translations", "lang", lang, "err", err)
		}
	}

	if _, ok := allTranslations[defaultLocale]; !ok {
		log.Info("Default language translations may not be loaded", "locale", defaultLocale)
	}

	return translator, loader
}
