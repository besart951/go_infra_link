package facility_test

import (
	"context"
	"errors"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	service "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/google/uuid"
)

func TestSPSControllerServiceDeleteByIDDeletesOwnedSystemTypes(t *testing.T) {
	controllerID := uuid.New()
	otherControllerID := uuid.New()
	ownedSystemTypeID := uuid.New()
	otherSystemTypeID := uuid.New()

	controllers := &fakeHierarchySPSControllerRepo{items: map[uuid.UUID]*facility.SPSController{
		controllerID:      {Base: domain.Base{ID: controllerID}},
		otherControllerID: {Base: domain.Base{ID: otherControllerID}},
	}}
	systemTypes := &fakeSpsControllerSystemTypeRepo{items: map[uuid.UUID]*facility.SPSControllerSystemType{
		ownedSystemTypeID: {Base: domain.Base{ID: ownedSystemTypeID}, SPSControllerID: controllerID},
		otherSystemTypeID: {Base: domain.Base{ID: otherSystemTypeID}, SPSControllerID: otherControllerID},
	}}
	fieldDevices := &fakeFieldDeviceStore{items: map[uuid.UUID]*facility.FieldDevice{}}

	service := service.NewSPSControllerService(
		controllers,
		nil,
		nil,
		nil,
		systemTypes,
		fieldDevices,
		nil,
	)

	if err := service.DeleteByID(context.Background(), controllerID); err != nil {
		t.Fatalf("DeleteByID() error = %v", err)
	}
	if _, exists := controllers.items[controllerID]; exists {
		t.Fatal("deleted controller still exists")
	}
	if _, exists := systemTypes.items[ownedSystemTypeID]; exists {
		t.Fatal("owned system type still exists")
	}
	if _, exists := systemTypes.items[otherSystemTypeID]; !exists {
		t.Fatal("unrelated system type was deleted")
	}
}

func TestSPSControllerServiceDeleteByIDBlocksFieldDeviceReferences(t *testing.T) {
	controllerID := uuid.New()
	systemTypeID := uuid.New()
	fieldDeviceID := uuid.New()

	controllers := &fakeHierarchySPSControllerRepo{items: map[uuid.UUID]*facility.SPSController{
		controllerID: {Base: domain.Base{ID: controllerID}},
	}}
	systemTypes := &fakeSpsControllerSystemTypeRepo{items: map[uuid.UUID]*facility.SPSControllerSystemType{
		systemTypeID: {Base: domain.Base{ID: systemTypeID}, SPSControllerID: controllerID},
	}}
	fieldDevices := &fakeFieldDeviceStore{items: map[uuid.UUID]*facility.FieldDevice{
		fieldDeviceID: {Base: domain.Base{ID: fieldDeviceID}, SPSControllerSystemTypeID: systemTypeID},
	}}

	service := service.NewSPSControllerService(
		controllers,
		nil,
		nil,
		nil,
		systemTypes,
		fieldDevices,
		nil,
	)

	err := service.DeleteByID(context.Background(), controllerID)
	if !errors.Is(err, facility.ErrReferenceInUse) {
		t.Fatalf("DeleteByID() error = %v, want ErrReferenceInUse", err)
	}
	if _, exists := controllers.items[controllerID]; !exists {
		t.Fatal("controller was deleted despite field-device reference")
	}
	if _, exists := systemTypes.items[systemTypeID]; !exists {
		t.Fatal("system type was deleted despite field-device reference")
	}
}
