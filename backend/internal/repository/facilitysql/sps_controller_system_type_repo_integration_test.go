package facilitysql

import (
	"context"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	projectsql "github.com/besart951/go_infra_link/backend/internal/repository/projectsql"
	"github.com/google/uuid"
)

func TestSPSControllerSystemTypeRepo_ProjectListUsesProjectSPSLinksAndCountsFieldDevices(t *testing.T) {
	ctx := context.Background()
	db := newFieldDeviceRepoTestDB(t)
	repo := NewSPSControllerSystemTypeRepository(db)
	projectSPSLinks := projectsql.NewProjectSPSControllerRepository(db)

	systemType := seedFacilityRecord(t, db, &domainFacility.SystemType{Name: "HVAC", NumberMin: 1, NumberMax: 99})
	otherSystemType := seedFacilityRecord(t, db, &domainFacility.SystemType{Name: "Lighting", NumberMin: 1, NumberMax: 99})
	controller := seedFacilityRecord(t, db, &domainFacility.SPSController{ControlCabinetID: uuid.New(), DeviceName: "SPS-A"})
	otherController := seedFacilityRecord(t, db, &domainFacility.SPSController{ControlCabinetID: uuid.New(), DeviceName: "SPS-B"})

	targetNumber := 7
	targetDocument := "DOC-A"
	targetSystemType := seedFacilityRecord(t, db, &domainFacility.SPSControllerSystemType{
		Number:          &targetNumber,
		DocumentName:    &targetDocument,
		SPSControllerID: controller.ID,
		SystemTypeID:    systemType.ID,
	})
	otherNumber := 8
	seedFacilityRecord(t, db, &domainFacility.SPSControllerSystemType{
		Number:          &otherNumber,
		SPSControllerID: otherController.ID,
		SystemTypeID:    otherSystemType.ID,
	})

	seedFacilityRecord(t, db, &FieldDeviceRecord{SPSControllerSystemTypeID: targetSystemType.ID, ApparatNr: 1, SystemPartID: uuid.New(), ApparatID: uuid.New()})
	seedFacilityRecord(t, db, &FieldDeviceRecord{SPSControllerSystemTypeID: targetSystemType.ID, ApparatNr: 2, SystemPartID: uuid.New(), ApparatID: uuid.New()})

	projectID := uuid.New()
	if err := projectSPSLinks.Create(ctx, &domainProject.ProjectSPSController{ProjectID: projectID, SPSControllerID: controller.ID}); err != nil {
		t.Fatalf("expected project sps link create to succeed, got %v", err)
	}

	list, err := repo.GetPaginatedListByProjectID(ctx, projectID, domain.PaginationParams{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("expected project-filtered system type list to succeed, got %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("expected one project-linked system type, got %+v", list.Items)
	}

	item := list.Items[0]
	if item.ID != targetSystemType.ID {
		t.Fatalf("expected target system type %s, got %s", targetSystemType.ID, item.ID)
	}
	if item.FieldDevicesCount != 2 {
		t.Fatalf("expected two field devices in aggregate count, got %d", item.FieldDevicesCount)
	}
	if item.SPSController.ID != controller.ID || item.SystemType.ID != systemType.ID {
		t.Fatalf("expected preloaded controller and system type, got controller=%+v systemType=%+v", item.SPSController, item.SystemType)
	}
}

func TestSPSControllerSystemTypeRepo_ListBySPSControllerIDsFiltersAndCountsFieldDevices(t *testing.T) {
	ctx := context.Background()
	db := newFieldDeviceRepoTestDB(t)
	repo := NewSPSControllerSystemTypeRepository(db)

	systemType := seedFacilityRecord(t, db, &domainFacility.SystemType{Name: "HVAC", NumberMin: 1, NumberMax: 99})
	otherSystemType := seedFacilityRecord(t, db, &domainFacility.SystemType{Name: "Lighting", NumberMin: 1, NumberMax: 99})
	controllerA := seedFacilityRecord(t, db, &domainFacility.SPSController{ControlCabinetID: uuid.New(), DeviceName: "SPS-A"})
	controllerB := seedFacilityRecord(t, db, &domainFacility.SPSController{ControlCabinetID: uuid.New(), DeviceName: "SPS-B"})
	controllerOutsideFilter := seedFacilityRecord(t, db, &domainFacility.SPSController{ControlCabinetID: uuid.New(), DeviceName: "SPS-C"})

	firstNumber := 1
	firstSystemType := seedFacilityRecord(t, db, &domainFacility.SPSControllerSystemType{
		Number:          &firstNumber,
		SPSControllerID: controllerA.ID,
		SystemTypeID:    systemType.ID,
	})
	secondNumber := 2
	secondSystemType := seedFacilityRecord(t, db, &domainFacility.SPSControllerSystemType{
		Number:          &secondNumber,
		SPSControllerID: controllerB.ID,
		SystemTypeID:    otherSystemType.ID,
	})
	outsideNumber := 3
	seedFacilityRecord(t, db, &domainFacility.SPSControllerSystemType{
		Number:          &outsideNumber,
		SPSControllerID: controllerOutsideFilter.ID,
		SystemTypeID:    systemType.ID,
	})

	seedFacilityRecord(t, db, &FieldDeviceRecord{SPSControllerSystemTypeID: firstSystemType.ID, ApparatNr: 1, SystemPartID: uuid.New(), ApparatID: uuid.New()})
	seedFacilityRecord(t, db, &FieldDeviceRecord{SPSControllerSystemTypeID: secondSystemType.ID, ApparatNr: 2, SystemPartID: uuid.New(), ApparatID: uuid.New()})
	seedFacilityRecord(t, db, &FieldDeviceRecord{SPSControllerSystemTypeID: secondSystemType.ID, ApparatNr: 3, SystemPartID: uuid.New(), ApparatID: uuid.New()})

	list, err := repo.GetPaginatedListBySPSControllerIDs(ctx, []uuid.UUID{controllerA.ID, controllerB.ID}, domain.PaginationParams{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("expected controller-filtered system type list to succeed, got %v", err)
	}
	if len(list.Items) != 2 {
		t.Fatalf("expected two matching system types, got %+v", list.Items)
	}

	itemsByID := make(map[uuid.UUID]domainFacility.SPSControllerSystemType, len(list.Items))
	for _, item := range list.Items {
		itemsByID[item.ID] = item
	}

	firstItem, ok := itemsByID[firstSystemType.ID]
	if !ok {
		t.Fatalf("expected first controller system type %s in result", firstSystemType.ID)
	}
	if firstItem.FieldDevicesCount != 1 {
		t.Fatalf("expected first system type to count one field device, got %d", firstItem.FieldDevicesCount)
	}
	if firstItem.SPSController.ID != controllerA.ID || firstItem.SystemType.ID != systemType.ID {
		t.Fatalf("expected preloaded first controller and system type, got controller=%+v systemType=%+v", firstItem.SPSController, firstItem.SystemType)
	}

	secondItem, ok := itemsByID[secondSystemType.ID]
	if !ok {
		t.Fatalf("expected second controller system type %s in result", secondSystemType.ID)
	}
	if secondItem.FieldDevicesCount != 2 {
		t.Fatalf("expected second system type to count two field devices, got %d", secondItem.FieldDevicesCount)
	}
	if secondItem.SPSController.ID != controllerB.ID || secondItem.SystemType.ID != otherSystemType.ID {
		t.Fatalf("expected preloaded second controller and system type, got controller=%+v systemType=%+v", secondItem.SPSController, secondItem.SystemType)
	}
}
