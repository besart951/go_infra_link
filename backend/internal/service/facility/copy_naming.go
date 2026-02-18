package facility

import (
	"strconv"
	"strings"
)

func nextIncrementedValue(base string, increment int, maxLen int) string {
	trimmedBase := strings.TrimSpace(base)
	if increment < 1 {
		increment = 1
	}

	prefix, currentNumber := splitTrailingNumber(trimmedBase)
	nextNumber := currentNumber + increment
	suffix := strconv.Itoa(nextNumber)

	if prefix == "" {
		if maxLen > 0 && len(suffix) > maxLen {
			return suffix[len(suffix)-maxLen:]
		}
		return suffix
	}

	if maxLen > 0 {
		allowedPrefixLen := maxLen - len(suffix)
		if allowedPrefixLen <= 0 {
			if len(suffix) > maxLen {
				return suffix[len(suffix)-maxLen:]
			}
			return suffix
		}
		if len(prefix) > allowedPrefixLen {
			prefix = prefix[:allowedPrefixLen]
		}
	}

	return prefix + suffix
}

func splitTrailingNumber(value string) (prefix string, number int) {
	if value == "" {
		return "", 0
	}

	end := len(value)
	start := end
	for start > 0 {
		ch := value[start-1]
		if ch < '0' || ch > '9' {
			break
		}
		start--
	}

	if start == end {
		return value, 0
	}

	parsed, err := strconv.Atoi(value[start:end])
	if err != nil {
		return value, 0
	}

	return value[:start], parsed
}
