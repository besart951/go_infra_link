package facilitysql

import (
	"strings"
	"testing"
)

func TestNullCursorAtLastRankSkipsEarlierRanks(t *testing.T) {
	predicate, args := fieldDeviceKeysetPredicate(fieldDeviceKeyset{
		columns: []fieldDeviceCursorColumn{{expression: "fd_cursor.value", nullable: true}},
		values:  []*string{nil},
		order:   fieldDeviceCursorOrder{order: "asc", direction: "next"},
	})
	if strings.Contains(predicate, " > ? OR") {
		t.Fatalf("boundary predicate scans earlier ranks: %s", predicate)
	}
	if !strings.Contains(predicate, "fd_cursor.value IS NULL") {
		t.Fatalf("predicate does not constrain the nullable index column: %s", predicate)
	}
	if len(args) != 2 || args[0] != 1 {
		t.Fatalf("unexpected boundary arguments: %#v", args)
	}
}

func TestNonNullPreviousCursorSkipsLaterNullRank(t *testing.T) {
	value := "AC"
	predicate, _ := fieldDeviceKeysetPredicate(fieldDeviceKeyset{
		columns: []fieldDeviceCursorColumn{{expression: "fd_cursor.value", nullable: true}},
		values:  []*string{&value},
		order:   fieldDeviceCursorOrder{order: "asc", direction: "previous"},
	})
	rankRange := "CASE WHEN fd_cursor.value IS NULL THEN 1 ELSE 0 END < ? OR"
	if strings.Contains(predicate, rankRange) {
		t.Fatalf("boundary predicate includes an impossible rank: %s", predicate)
	}
}

func TestCursorRankSplitCoversBothCrossRankDirections(t *testing.T) {
	value := "AC"
	column := []fieldDeviceCursorColumn{{expression: "fd_cursor.value", nullable: true}}
	tests := []fieldDeviceKeyset{
		{columns: column, values: []*string{&value}, order: fieldDeviceCursorOrder{direction: "next"}},
		{columns: column, values: []*string{nil}, order: fieldDeviceCursorOrder{direction: "previous"}},
	}
	for _, keyset := range tests {
		if !needsCursorRankSplit(keyset) {
			t.Fatalf("expected rank split for %#v", keyset)
		}
	}
}

func TestSingleColumnRankSplitUsesIndexableTuple(t *testing.T) {
	value := "field device 42"
	predicate, args := fieldDeviceOuterSameRankPredicate(fieldDeviceKeyset{
		columns:      []fieldDeviceCursorColumn{{expression: "field_devices.description", nullable: true}},
		values:       []*string{&value},
		idExpression: "field_devices.id",
		order:        fieldDeviceCursorOrder{order: "desc", direction: "next"},
	})
	if !strings.Contains(predicate, "(field_devices.description,field_devices.id) < (?,?)") {
		t.Fatalf("predicate does not use the covering-index tuple: %s", predicate)
	}
	if len(args) != 3 || args[0] != 0 {
		t.Fatalf("unexpected tuple arguments: %#v", args)
	}
}

func TestSingleColumnPreviousCursorUsesIndexableTuple(t *testing.T) {
	value := "supplier 42"
	predicate, _ := fieldDeviceKeysetPredicate(fieldDeviceKeyset{
		columns: []fieldDeviceCursorColumn{{expression: "fd_cursor.supplier", nullable: true}},
		values:  []*string{&value},
		order:   fieldDeviceCursorOrder{order: "asc", direction: "previous"},
	})
	if !strings.Contains(predicate, "(fd_cursor.supplier,field_devices.id) < (?,?)") {
		t.Fatalf("previous predicate does not use tuple comparison: %s", predicate)
	}
}

func TestMultiColumnCursorUsesCoveringIndexTuple(t *testing.T) {
	first, second := "42", "DOC-42"
	predicate, args := fieldDeviceKeysetPredicate(fieldDeviceKeyset{
		columns: []fieldDeviceCursorColumn{
			{expression: "fd_cursor.sps_number", nullable: true},
			{expression: "fd_cursor.sps_document_name", nullable: true},
		},
		values:       []*string{&first, &second},
		idExpression: "fd_cursor.field_device_id",
		order:        fieldDeviceCursorOrder{order: "asc", direction: "next"},
	})
	want := "(CASE WHEN fd_cursor.sps_number IS NULL THEN 1 ELSE 0 END,fd_cursor.sps_number,CASE WHEN fd_cursor.sps_document_name IS NULL THEN 1 ELSE 0 END,fd_cursor.sps_document_name,fd_cursor.field_device_id) > (?,?,?,?,?)"
	if predicate != want {
		t.Fatalf("predicate = %s, want %s", predicate, want)
	}
	if len(args) != 5 {
		t.Fatalf("unexpected tuple arguments: %#v", args)
	}
}

func TestSecondNullNextCursorUsesSplitPlan(t *testing.T) {
	first := "990"
	keyset := fieldDeviceKeyset{
		columns: []fieldDeviceCursorColumn{
			{expression: "fd_cursor.sps_number", nullable: true},
			{expression: "fd_cursor.sps_document_name", nullable: true},
		},
		values: []*string{&first, nil},
		order:  fieldDeviceCursorOrder{order: "asc", direction: "next"},
	}
	if !needsSecondNullNextSplit(keyset) {
		t.Fatal("expected a split at the nested NULL boundary")
	}
	predicate, _ := secondNullSameValuePredicate(keyset)
	if strings.Contains(predicate, " OR ") {
		t.Fatalf("primary branch is not indexable: %s", predicate)
	}
}
