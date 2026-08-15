package main

import (
	"slices"
	"testing"
)

func TestSnapshotColumnsUsesOnlyCurrentSnapshotColumns(t *testing.T) {
	rows := []map[string]any{
		{"id": "one", "name": "first", "legacy": true},
		{"id": "two", "name": "second"},
	}
	databaseColumns := map[string]struct{}{
		"id":      {},
		"name":    {},
		"version": {},
	}

	got := snapshotColumns(rows, databaseColumns)
	want := []string{"id", "name"}
	if !slices.Equal(got, want) {
		t.Fatalf("snapshotColumns() = %v, want %v", got, want)
	}
}
