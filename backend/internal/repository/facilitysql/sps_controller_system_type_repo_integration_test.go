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

	seedFacilityRecord(t, db, &fieldDeviceRecord{SPSControllerSystemTypeID: targetSystemType.ID, ApparatNr: 1, SystemPartID: uuid.New(), ApparatID: uuid.New()})
	seedFacilityRecord(t, db, &fieldDeviceRecord{SPSControllerSystemTypeID: targetSystemType.ID, ApparatNr: 2, SystemPartID: uuid.New(), ApparatID: uuid.New()})

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
