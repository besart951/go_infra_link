package dto

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"
)

const swissLocationName = "Europe/Zurich"

// SwissDateTime parses RFC3339 timestamps or date-only strings in Swiss time.
type SwissDateTime struct {
	time.Time
}

func (t *SwissDateTime) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		return nil
	}

	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	raw = strings.TrimSpace(raw)
	if raw == "" {
		return &time.ParseError{Layout: time.RFC3339, Value: raw, LayoutElem: "T", ValueElem: ""}
	}

	if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
		t.Time = parsed
		return nil
	}

	loc, err := time.LoadLocation(swissLocationName)
	if err != nil {
		loc = time.Local
	}

	parsed, err := time.ParseInLocation("2006-01-02", raw, loc)
	if err != nil {
		return err
	}

	t.Time = parsed
	return nil
}
