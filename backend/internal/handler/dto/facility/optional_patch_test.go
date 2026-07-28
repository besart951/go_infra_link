package facility

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestOptionalUUIDDistinguishesOmittedNullAndValue(t *testing.T) {
	type payload struct {
		FieldDeviceID OptionalUUID `json:"field_device_id"`
	}

	var omitted payload
	if err := json.Unmarshal([]byte(`{}`), &omitted); err != nil {
		t.Fatalf("decode omitted: %v", err)
	}
	if omitted.FieldDeviceID.Set {
		t.Fatal("omitted UUID was marked as set")
	}

	var detached payload
	if err := json.Unmarshal([]byte(`{"field_device_id":null}`), &detached); err != nil {
		t.Fatalf("decode null: %v", err)
	}
	if !detached.FieldDeviceID.Set || detached.FieldDeviceID.Value != nil {
		t.Fatalf("null UUID: %+v", detached.FieldDeviceID)
	}

	want := uuid.MustParse("00000000-0000-0000-0000-000000000101")
	var assigned payload
	if err := json.Unmarshal(
		[]byte(`{"field_device_id":"`+want.String()+`"}`),
		&assigned,
	); err != nil {
		t.Fatalf("decode UUID: %v", err)
	}
	if !assigned.FieldDeviceID.Set ||
		assigned.FieldDeviceID.Value == nil ||
		*assigned.FieldDeviceID.Value != want {
		t.Fatalf("assigned UUID: %+v", assigned.FieldDeviceID)
	}
}
