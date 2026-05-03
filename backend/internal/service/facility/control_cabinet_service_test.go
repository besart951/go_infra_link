package facility_test

import (
	"context"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/google/uuid"
)

func TestControlCabinetServiceUpdateRegeneratesSPSControllerDeviceNames(t *testing.T) {
	ctx := context.Background()
	buildingID := uuid.New()
	controlCabinetID := uuid.New()
	firstControllerID := uuid.New()
	secondControllerID := uuid.New()
	oldCabinetNr := "AK01"
	newCabinetNr := "AK02"
	firstGADevice := "A01"
	secondGADevice := "B02"

	buildings := &fakeHierarchyBuildingRepo{items: map[uuid.UUID]*domainFacility.Building{
		buildingID: {Base: domain.Base{ID: buildingID}, IWSCode: "IWS1"},
	}}
	controlCabinets := &fakeHierarchyControlCabinetRepo{items: map[uuid.UUID]*domainFacility.ControlCabinet{
		controlCabinetID: {
			Base:             domain.Base{ID: controlCabinetID},
			BuildingID:       buildingID,
			ControlCabinetNr: &oldCabinetNr,
		},
	}}
	spsControllers := &fakeHierarchySPSControllerRepo{items: map[uuid.UUID]*domainFacility.SPSController{
		firstControllerID: {
			Base:             domain.Base{ID: firstControllerID},
			ControlCabinetID: controlCabinetID,
			GADevice:         &firstGADevice,
			DeviceName:       "IWS1_AK01_A01",
		},
		secondControllerID: {
			Base:             domain.Base{ID: secondControllerID},
			ControlCabinetID: controlCabinetID,
			GADevice:         &secondGADevice,
			DeviceName:       "IWS1_AK01_B02",
		},
	}}

	svc := facility.NewControlCabinetService(
		controlCabinets,
		buildings,
		spsControllers,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	updated := *controlCabinets.items[controlCabinetID]
	updated.ControlCabinetNr = &newCabinetNr

	if err := svc.Update(ctx, &updated); err != nil {
		t.Fatalf("update control cabinet: %v", err)
	}

	if got := spsControllers.items[firstControllerID].DeviceName; got != "IWS1_AK02_A01" {
		t.Fatalf("expected first SPS device name to follow cabinet rename, got %q", got)
	}
	if got := spsControllers.items[secondControllerID].DeviceName; got != "IWS1_AK02_B02" {
		t.Fatalf("expected second SPS device name to follow cabinet rename, got %q", got)
	}
}
