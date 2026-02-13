package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Loader loads translation files from the filesystem.
type Loader struct {
	basePath string
}

// NewLoader creates a new Loader with the given base path for translation files.
// basePath should point to the directory containing locale files.
// Example: "./internal/locales"
func NewLoader(basePath string) *Loader {
	return &Loader{
		basePath: basePath,
	}
}

// Load loads a translation file for a specific language.
// filename should be the name of the file (e.g., "de_ch.json")
// Returns the parsed translation data as a map.
func (l *Loader) Load(filename string) (map[string]interface{}, error) {
	path := filepath.Join(l.basePath, filename)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read translation file %s: %w", filename, err)
	}

	var translations map[string]interface{}
	if err := json.Unmarshal(data, &translations); err != nil {
		return nil, fmt.Errorf("failed to parse translation file %s: %w", filename, err)
	}

	return translations, nil
}

// LoadAll loads all translation files from the base directory.
// Only .json files are loaded.
// Returns a map of language -> translations.
func (l *Loader) LoadAll() (map[string]map[string]interface{}, error) {
	result := make(map[string]map[string]interface{})

	entries, err := os.ReadDir(l.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read locales directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		translations, err := l.Load(entry.Name())
		if err != nil {
			return nil, err
		}

		// Extract language code from filename (e.g., "de_ch.json" -> "de_CH")
		lang := filenameTolanguageCode(entry.Name())
		result[lang] = translations
	}

	return result, nil
}

// filenameTolanguageCode converts a filename to a language code.
// Example: "de_ch.json" -> "de_CH"
func filenameTolanguageCode(filename string) string {
	// Remove .json extension
	lang := filename[:len(filename)-5]
	return lang
}
