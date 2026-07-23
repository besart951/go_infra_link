package facility

import (
	"errors"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

func TestFieldDeviceMoveToChangesPlacementAsOneDomainTransition(t *testing.T) {
	fieldDevice := &FieldDevice{
		SPSControllerSystemTypeID: uuid.New(),
		SystemPartID:              uuid.New(),
		ApparatID:                 uuid.New(),
		ApparatNr:                 1,
	}
	want, err := NewFieldDevicePlacement(
		uuid.New(),
		uuid.New(),
		uuid.New(),
		99,
	)
	if err != nil {
		t.Fatalf("create placement: %v", err)
	}

	if err := fieldDevice.MoveTo(want); err != nil {
		t.Fatalf("move FieldDevice: %v", err)
	}
	if got := fieldDevice.Placement(); got != want {
		t.Fatalf("placement: got %+v, want %+v", got, want)
	}
}

func TestNewFieldDevicePlacementRejectsInvalidLocalState(t *testing.T) {
	validID := uuid.New()
	tests := []struct {
		name       string
		systemType uuid.UUID
		systemPart uuid.UUID
		apparat    uuid.UUID
		number     int
	}{
		{name: "missing system type", systemPart: validID, apparat: validID, number: 1},
		{name: "missing system part", systemType: validID, apparat: validID, number: 1},
		{name: "missing apparat", systemType: validID, systemPart: validID, number: 1},
		{name: "number below range", systemType: validID, systemPart: validID, apparat: validID},
		{name: "number above range", systemType: validID, systemPart: validID, apparat: validID, number: 100},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewFieldDevicePlacement(
				test.systemType,
				test.systemPart,
				test.apparat,
				test.number,
			)
			if !errors.Is(err, domain.ErrInvalidArgument) {
				t.Fatalf("error: got %v, want invalid argument", err)
			}
		})
	}
}
