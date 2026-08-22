// Package postgresjson maps optional JSON documents to PostgreSQL jsonb.
package postgresjson

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// Document is an optional validated JSON value. Its zero value persists as
// SQL NULL, never as PostgreSQL's invalid empty JSON string.
type Document []byte

func (d Document) Value() (driver.Value, error) {
	if len(d) == 0 {
		return nil, nil
	}
	if !json.Valid(d) {
		return nil, fmt.Errorf("invalid JSON document")
	}
	return string(d), nil
}

func (d *Document) Scan(value any) error {
	if value == nil {
		*d = nil
		return nil
	}
	bytes, ok := scannedBytes(value)
	if !ok {
		return fmt.Errorf("scan JSON document: unsupported type %T", value)
	}
	if !json.Valid(bytes) {
		return fmt.Errorf("scan JSON document: invalid JSON")
	}
	*d = append((*d)[:0], bytes...)
	return nil
}

func scannedBytes(value any) ([]byte, bool) {
	switch typed := value.(type) {
	case []byte:
		return typed, true
	case string:
		return []byte(typed), true
	default:
		return nil, false
	}
}

func (d Document) Bytes() []byte {
	return append([]byte(nil), d...)
}
