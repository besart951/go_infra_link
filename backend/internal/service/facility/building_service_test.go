package facility_test

import (
	"context"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/google/uuid"
)

func TestBuildingServiceUpdateRegeneratesDescendantSPSControllerNames(t *testing.T) {
	buildingID := uuid.New()
	cabinetID := uuid.New()
	controllerID := uuid.New()
	cabinetNumber := "AK01"
	gaDevice := "ABC"

	buildings := &fakeHierarchyBuildingRepo{items: map[uuid.UUID]*domainFacility.Building{
		buildingID: {
			Base:          domain.Base{ID: buildingID},
			IWSCode:       "E001",
			BuildingGroup: 1,
		},
	}}
	cabinets := &fakeHierarchyControlCabinetRepo{items: map[uuid.UUID]*domainFacility.ControlCabinet{
		cabinetID: {
			Base:             domain.Base{ID: cabinetID},
			BuildingID:       buildingID,
			ControlCabinetNr: &cabinetNumber,
		},
	}}
	controllers := &fakeHierarchySPSControllerRepo{items: map[uuid.UUID]*domainFacility.SPSController{
		controllerID: {
			Base:             domain.Base{ID: controllerID},
			ControlCabinetID: cabinetID,
			GADevice:         &gaDevice,
			DeviceName:       "E001_AK01_ABC",
		},
	}}

	synchronizer := facility.NewSPSControllerNameSynchronizer(buildings, cabinets, controllers)
	service := facility.NewBuildingService(buildings, synchronizer)
	updated := *buildings.items[buildingID]
	updated.IWSCode = "E002"

	if err := service.Update(context.Background(), &updated); err != nil {
		t.Fatalf("update building: %v", err)
	}

	if got := controllers.items[controllerID].DeviceName; got != "E002_AK01_ABC" {
		t.Fatalf("expected controller name to follow IWS-code change, got %q", got)
	}
}
