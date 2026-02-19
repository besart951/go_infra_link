package facility

import "strings"

func normalizeBacnetTextFix(value string) string {
	replacer := strings.NewReplacer(
		"ä", "ae",
		"ö", "oe",
		"ü", "ue",
		"Ä", "Ae",
		"Ö", "Oe",
		"Ü", "Ue",
	)
	return replacer.Replace(value)
}
