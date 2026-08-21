package cursor

import (
	"errors"
	"testing"
)

func TestCodecRoundTripAndKindIsolation(t *testing.T) {
	type payload struct {
		ID          string `json:"id"`
		Fingerprint string `json:"f"`
	}

	encoded, err := Encode("facility_jobs", payload{ID: "abc", Fingerprint: "scope"})
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	var decoded payload
	if err := Decode(encoded, "facility_jobs", &decoded); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if decoded.ID != "abc" || decoded.Fingerprint != "scope" {
		t.Fatalf("decoded = %#v", decoded)
	}
	if err := Decode(encoded, "field_devices", &decoded); !errors.Is(err, ErrInvalid) {
		t.Fatalf("wrong-kind error = %v, want %v", err, ErrInvalid)
	}
}
