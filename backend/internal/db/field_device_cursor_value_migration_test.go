package db

import (
	"strings"
	"testing"
)

func TestFieldDeviceCursorProjectionPreservesNumericSortTypes(t *testing.T) {
	for _, definition := range []string{
		"additional_info_size integer",
		"electrical_connection_ph integer",
		"electrical_connection_amperage double precision",
		"electrical_connection_power double precision",
		"electrical_connection_rotation integer",
	} {
		if !strings.Contains(createFieldDeviceCursorValueTable, definition) {
			t.Errorf("cursor projection is missing %q", definition)
		}
	}
}
