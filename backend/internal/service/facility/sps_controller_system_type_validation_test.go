package facility

import (
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func TestAssignSystemTypeNumbersReportsIndexedFieldErrors(t *testing.T) {
	typeID := uuid.New()
	first := 2
	duplicate := 2
	outOfRange := 9
	inputs := []domainFacility.SPSControllerSystemType{
		{SystemTypeID: typeID, Number: &first},
		{SystemTypeID: typeID, Number: &duplicate},
		{SystemTypeID: typeID, Number: &outOfRange},
	}

	err := assignSystemTypeNumbers(inputs, map[uuid.UUID]domainFacility.SystemType{
		typeID: {NumberMin: 1, NumberMax: 3},
	})
	validationErr, ok := domain.AsValidationError(err)
	if !ok {
		t.Fatalf("error = %T, want *domain.ValidationError", err)
	}
	if got := validationErr.Codes["spscontroller.system_types[1].number"]; got != "unique" {
		t.Errorf("duplicate code = %q, want unique", got)
	}
	if got := validationErr.Codes["spscontroller.system_types[2].number"]; got != "range" {
		t.Errorf("range code = %q, want range", got)
	}
}
