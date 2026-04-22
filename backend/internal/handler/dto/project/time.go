package project

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

// OptionalSwissDateTime tracks whether a timestamp field was present in a
// PATCH-like request. A present null clears the target value.
type OptionalSwissDateTime struct {
	Set   bool
	Value *SwissDateTime
}

func (t *OptionalSwissDateTime) UnmarshalJSON(data []byte) error {
	t.Set = true
	if bytes.Equal(data, []byte("null")) {
		t.Value = nil
		return nil
	}

	var parsed SwissDateTime
	if err := parsed.UnmarshalJSON(data); err != nil {
		return err
	}
	t.Value = &parsed
	return nil
}
