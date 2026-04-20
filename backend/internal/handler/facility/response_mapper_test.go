package facility

import (
	"testing"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func TestToSPSControllerSystemTypeResponseIncludesFieldDevicesCount(t *testing.T) {
	item := domainFacility.SPSControllerSystemType{
		SPSControllerID:   uuid.New(),
		SystemTypeID:      uuid.New(),
		SPSController:     domainFacility.SPSController{DeviceName: "SPS-A"},
		SystemType:        domainFacility.SystemType{Name: "Cooling"},
		FieldDevicesCount: 7,
	}

	response := toSPSControllerSystemTypeResponse(item)

	if response.FieldDevicesCount != 7 {
		t.Fatalf("expected field device count 7, got %d", response.FieldDevicesCount)
	}
}
