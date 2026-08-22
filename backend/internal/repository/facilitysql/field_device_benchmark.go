package facilitysql

import (
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

// EncodeFieldDeviceBenchmarkCursor creates a production cursor from an anchor
// selected by the benchmark setup. Application handlers must never call it.
func EncodeFieldDeviceBenchmarkCursor(query domainFacility.FieldDeviceCursorQuery, direction string, id uuid.UUID, value, second *string) (string, error) {
	normalized, _ := normalizeFieldDeviceCursorQuery(query)
	fingerprint, err := fieldDeviceQueryFingerprint(normalized)
	if err != nil {
		return "", err
	}
	return encodeFieldDeviceCursor(fieldDeviceCursorRow{ID: id, Value: value, SecondValue: second}, fingerprint, direction)
}
