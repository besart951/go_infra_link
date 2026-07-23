package facility_test

import (
	"context"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/google/uuid"
)

func TestSPSControllerServiceNextAvailableGADeviceDoesNotExcludeControllerFromAnotherCabinet(t *testing.T) {
	targetCabinetID := uuid.New()
	sourceCabinetID := uuid.New()
	targetControllerID := uuid.New()
	sourceControllerID := uuid.New()
	gaDevice := "AAA"

	repo := &fakeHierarchySPSControllerRepo{items: map[uuid.UUID]*domainFacility.SPSController{
		targetControllerID: {
			Base:             domain.Base{ID: targetControllerID},
			ControlCabinetID: targetCabinetID,
			GADevice:         &gaDevice,
		},
		sourceControllerID: {
			Base:             domain.Base{ID: sourceControllerID},
			ControlCabinetID: sourceCabinetID,
			GADevice:         &gaDevice,
		},
	}}
	service := facility.NewSPSControllerService(
		repo,
		&fakeHierarchyControlCabinetRepo{items: map[uuid.UUID]*domainFacility.ControlCabinet{
			targetCabinetID: {Base: domain.Base{ID: targetCabinetID}},
		}},
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	next, err := service.NextAvailableGADevice(
		context.Background(),
		targetCabinetID,
		&sourceControllerID,
	)
	if err != nil {
		t.Fatalf("next GA device: %v", err)
	}
	if next != "BAA" {
		t.Fatalf("expected target cabinet's AAA to remain occupied, got %q", next)
	}
}

func TestSPSControllerServiceNextAvailableGADeviceCanReuseExcludedControllerInTargetCabinet(t *testing.T) {
	targetCabinetID := uuid.New()
	targetControllerID := uuid.New()
	gaDevice := "AAA"

	repo := &fakeHierarchySPSControllerRepo{items: map[uuid.UUID]*domainFacility.SPSController{
		targetControllerID: {
			Base:             domain.Base{ID: targetControllerID},
			ControlCabinetID: targetCabinetID,
			GADevice:         &gaDevice,
		},
	}}
	service := facility.NewSPSControllerService(
		repo,
		&fakeHierarchyControlCabinetRepo{items: map[uuid.UUID]*domainFacility.ControlCabinet{
			targetCabinetID: {Base: domain.Base{ID: targetCabinetID}},
		}},
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	next, err := service.NextAvailableGADevice(
		context.Background(),
		targetCabinetID,
		&targetControllerID,
	)
	if err != nil {
		t.Fatalf("next GA device: %v", err)
	}
	if next != gaDevice {
		t.Fatalf("expected excluded controller's %s to be reusable, got %q", gaDevice, next)
	}
}

func TestSPSControllerServiceUpdateMoveRegeneratesDestinationDeviceName(t *testing.T) {
	oldBuildingID := uuid.New()
	newBuildingID := uuid.New()
	oldCabinetID := uuid.New()
	newCabinetID := uuid.New()
	controllerID := uuid.New()
	oldCabinetNumber := "CC01"
	newCabinetNumber := "CC02"
	gaDevice := "AAA"

	repo := &fakeHierarchySPSControllerRepo{items: map[uuid.UUID]*domainFacility.SPSController{
		controllerID: {
			Base:             domain.Base{ID: controllerID},
			ControlCabinetID: oldCabinetID,
			GADevice:         &gaDevice,
			DeviceName:       "OLD_CC01_AAA",
		},
	}}
	service := facility.NewSPSControllerService(
		repo,
		&fakeHierarchyControlCabinetRepo{items: map[uuid.UUID]*domainFacility.ControlCabinet{
			oldCabinetID: {
				Base:             domain.Base{ID: oldCabinetID},
				BuildingID:       oldBuildingID,
				ControlCabinetNr: &oldCabinetNumber,
			},
			newCabinetID: {
				Base:             domain.Base{ID: newCabinetID},
				BuildingID:       newBuildingID,
				ControlCabinetNr: &newCabinetNumber,
			},
		}},
		&fakeHierarchyBuildingRepo{items: map[uuid.UUID]*domainFacility.Building{
			oldBuildingID: {Base: domain.Base{ID: oldBuildingID}, IWSCode: "OLD"},
			newBuildingID: {Base: domain.Base{ID: newBuildingID}, IWSCode: "NEW"},
		}},
		nil,
		nil,
		nil,
		nil,
	)
	updated := *repo.items[controllerID]
	updated.ControlCabinetID = newCabinetID

	if err := service.Update(context.Background(), &updated); err != nil {
		t.Fatalf("move SPS controller: %v", err)
	}
	if got := repo.items[controllerID].ControlCabinetID; got != newCabinetID {
		t.Fatalf("control cabinet: got %s, want %s", got, newCabinetID)
	}
	if got := repo.items[controllerID].DeviceName; got != "NEW_CC02_AAA" {
		t.Fatalf("generated device name: got %q, want NEW_CC02_AAA", got)
	}
}

func TestSPSControllerServiceUpdateMoveRejectsDestinationGAConflict(t *testing.T) {
	buildingID := uuid.New()
	oldCabinetID := uuid.New()
	newCabinetID := uuid.New()
	movingControllerID := uuid.New()
	occupyingControllerID := uuid.New()
	oldCabinetNumber := "CC01"
	newCabinetNumber := "CC02"
	gaDevice := "AAA"

	repo := &fakeHierarchySPSControllerRepo{items: map[uuid.UUID]*domainFacility.SPSController{
		movingControllerID: {
			Base:             domain.Base{ID: movingControllerID},
			ControlCabinetID: oldCabinetID,
			GADevice:         &gaDevice,
			DeviceName:       "BLD_CC01_AAA",
		},
		occupyingControllerID: {
			Base:             domain.Base{ID: occupyingControllerID},
			ControlCabinetID: newCabinetID,
			GADevice:         &gaDevice,
			DeviceName:       "BLD_CC02_AAA",
		},
	}}
	service := facility.NewSPSControllerService(
		repo,
		&fakeHierarchyControlCabinetRepo{items: map[uuid.UUID]*domainFacility.ControlCabinet{
			oldCabinetID: {
				Base:             domain.Base{ID: oldCabinetID},
				BuildingID:       buildingID,
				ControlCabinetNr: &oldCabinetNumber,
			},
			newCabinetID: {
				Base:             domain.Base{ID: newCabinetID},
				BuildingID:       buildingID,
				ControlCabinetNr: &newCabinetNumber,
			},
		}},
		&fakeHierarchyBuildingRepo{items: map[uuid.UUID]*domainFacility.Building{
			buildingID: {Base: domain.Base{ID: buildingID}, IWSCode: "BLD"},
		}},
		nil,
		nil,
		nil,
		nil,
	)
	updated := *repo.items[movingControllerID]
	updated.ControlCabinetID = newCabinetID

	err := service.Update(context.Background(), &updated)

	validationErr, ok := domain.AsValidationError(err)
	if !ok || validationErr.Fields["spscontroller.ga_device"] == "" {
		t.Fatalf("expected destination GA validation error, got %v", err)
	}
	if got := repo.items[movingControllerID].ControlCabinetID; got != oldCabinetID {
		t.Fatalf("failed move changed persisted cabinet to %s", got)
	}
}
