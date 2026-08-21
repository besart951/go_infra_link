// Package cursor encodes versioned, query-specific keyset cursors.
//
// A cursor is intentionally not an authorization token. Callers must apply
// the authenticated scope again when executing the next page.
package cursor

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
)

var ErrInvalid = errors.New("invalid cursor")

type envelope struct {
	Version int             `json:"v"`
	Kind    string          `json:"k"`
	Payload json.RawMessage `json:"p"`
}

// Encode serializes a query-specific payload as an opaque URL-safe cursor.
func Encode(kind string, payload any) (string, error) {
	if kind == "" {
		return "", fmt.Errorf("%w: missing kind", ErrInvalid)
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encode cursor payload: %w", err)
	}
	encoded, err := json.Marshal(envelope{Version: 1, Kind: kind, Payload: raw})
	if err != nil {
		return "", fmt.Errorf("encode cursor envelope: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(encoded), nil
}

// Decode validates the cursor version and query kind before decoding it.
func Decode(value, kind string, target any) error {
	if value == "" || kind == "" || target == nil {
		return ErrInvalid
	}
	raw, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return ErrInvalid
	}
	var decoded envelope
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return ErrInvalid
	}
	if decoded.Version != 1 || decoded.Kind != kind || len(decoded.Payload) == 0 {
		return ErrInvalid
	}
	if err := json.Unmarshal(decoded.Payload, target); err != nil {
		return ErrInvalid
	}
	return nil
}
