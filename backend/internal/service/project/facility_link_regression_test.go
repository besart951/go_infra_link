package project

import (
	"context"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/google/uuid"
)

func TestProjectService_CopySPSController_CharacterizesDeepCopyAndProjectLinks(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	buildingID := uuid.New()
	controlCabinetID := uuid.New()
	originalSPSID := uuid.New()
	systemTypeDefinitionID := uuid.New()
	originalSystemTypeID := uuid.New()
	originalFieldDeviceID := uuid.New()
	apparatID := uuid.New()
	cabinetNr := "AK1"
	gaDevice := "A01"
	originalNumber := 1

	spsLinks := newProjectSPSControllerRepo()
	fieldDeviceLinks := newProjectFieldDeviceRepo()
	controlCabinetRepo := newProjectControlCabinetStore()
	buildingRepo := newProjectBuildingRepo()
	spsRepo := newProjectSPSRepo()
	systemTypeRepo := newProjectSystemTypeRepo()
	spsSystemRepo := newProjectSPSSystemTypeRepo()
	fieldDeviceRepo := newProjectFieldDeviceStore()
	specRepo := newProjectSpecificationRepo()
	bacnetRepo := newProjectBacnetObjectRepo()

	buildingRepo.items[buildingID] = &domainFacility.Building{
		Base:          domain.Base{ID: buildingID},
		IWSCode:       "B01",
		BuildingGroup: 1,
	}
	controlCabinetRepo.items[controlCabinetID] = &domainFacility.ControlCabinet{
		Base:             domain.Base{ID: controlCabinetID},
		BuildingID:       buildingID,
		ControlCabinetNr: &cabinetNr,
	}
	spsRepo.items[originalSPSID] = &domainFacility.SPSController{
		Base:             domain.Base{ID: originalSPSID},
		ControlCabinetID: controlCabinetID,
		GADevice:         &gaDevice,
		DeviceName:       "B01_AK1_A01",
	}
	systemTypeRepo.items[systemTypeDefinitionID] = &domainFacility.SystemType{
		Base:      domain.Base{ID: systemTypeDefinitionID},
		NumberMin: 1,
		NumberMax: 99,
	}
	spsSystemRepo.items[originalSystemTypeID] = &domainFacility.SPSControllerSystemType{
		Base:            domain.Base{ID: originalSystemTypeID},
		Number:          &originalNumber,
		SPSControllerID: originalSPSID,
		SystemTypeID:    systemTypeDefinitionID,
	}
	fieldDeviceRepo.items[originalFieldDeviceID] = &domainFacility.FieldDevice{
		Base:                      domain.Base{ID: originalFieldDeviceID},
		SPSControllerSystemTypeID: originalSystemTypeID,
		ApparatID:                 apparatID,
		ApparatNr:                 3,
	}

	hierarchyCopier := facilityservice.NewHierarchyCopier(
		controlCabinetRepo,
		buildingRepo,
		spsRepo,
		systemTypeRepo,
		spsSystemRepo,
		fieldDeviceRepo,
		specRepo,
		bacnetRepo,
	)
	svc := newProjectCharacterizationServices(
		newProjectRepo(),
		newProjectControlCabinetRepo(),
		spsLinks,
		fieldDeviceLinks,
		newProjectObjectDataRepo(),
		bacnetRepo,
		specRepo,
		controlCabinetRepo,
		spsRepo,
		spsSystemRepo,
		fieldDeviceRepo,
		hierarchyCopier,
	).FacilityLink

	copiedSPS, err := svc.CopySPSController(ctx, projectID, originalSPSID)
	if err != nil {
		t.Fatalf("expected sps controller copy to succeed, got %v", err)
	}

	if copiedSPS.ID == uuid.Nil || copiedSPS.ID == originalSPSID {
		t.Fatalf("expected copied sps controller to get a new id, got %s", copiedSPS.ID)
	}
	if copiedSPS.GADevice == nil || *copiedSPS.GADevice == gaDevice {
		t.Fatalf("expected copied sps controller to get the next ga_device after %q, got %+v", gaDevice, copiedSPS.GADevice)
	}
	if got := spsLinks.spsControllerIDs(projectID); !sameUUIDSet(got, []uuid.UUID{copiedSPS.ID}) {
		t.Fatalf("expected project to link copied sps controller %s, got %v", copiedSPS.ID, got)
	}

	copiedSystemTypeIDs, err := spsSystemRepo.GetIDsBySPSControllerIDs(ctx, []uuid.UUID{copiedSPS.ID})
	if err != nil {
		t.Fatalf("expected copied system type lookup to succeed, got %v", err)
	}
	if len(copiedSystemTypeIDs) != 1 || copiedSystemTypeIDs[0] == originalSystemTypeID {
		t.Fatalf("expected one copied sps controller system type, got %v", copiedSystemTypeIDs)
	}

	copiedFieldDeviceIDs := fieldDeviceRepo.idsForSystemType(copiedSystemTypeIDs[0])
	if len(copiedFieldDeviceIDs) != 1 || copiedFieldDeviceIDs[0] == originalFieldDeviceID {
		t.Fatalf("expected one copied field device, got %v", copiedFieldDeviceIDs)
	}
	if got := fieldDeviceLinks.fieldDeviceIDs(projectID); !sameUUIDSet(got, copiedFieldDeviceIDs) {
		t.Fatalf("expected project to link copied field devices %v, got %v", copiedFieldDeviceIDs, got)
	}
}

func TestProjectService_DeleteSPSController_CharacterizesLinkAndHierarchyDeletion(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	linkID := uuid.New()
	spsID := uuid.New()
	systemTypeID := uuid.New()
	fieldDeviceID := uuid.New()

	spsLinks := newProjectSPSControllerRepo()
	fieldDeviceLinks := newProjectFieldDeviceRepo()
	spsRepo := newProjectSPSRepo()
	spsSystemRepo := newProjectSPSSystemTypeRepo()
	fieldDeviceRepo := newProjectFieldDeviceStore()
	specRepo := newProjectSpecificationRepo()
	bacnetRepo := newProjectBacnetObjectRepo()

	spsLinks.items[linkID] = &domainProject.ProjectSPSController{
		Base:            domain.Base{ID: linkID},
		ProjectID:       projectID,
		SPSControllerID: spsID,
	}
	fieldDeviceLinks.createWithID(projectID, fieldDeviceID)
	spsRepo.items[spsID] = &domainFacility.SPSController{Base: domain.Base{ID: spsID}}
	spsSystemRepo.items[systemTypeID] = &domainFacility.SPSControllerSystemType{
		Base:            domain.Base{ID: systemTypeID},
		SPSControllerID: spsID,
	}
	fieldDeviceRepo.items[fieldDeviceID] = &domainFacility.FieldDevice{
		Base:                      domain.Base{ID: fieldDeviceID},
		SPSControllerSystemTypeID: systemTypeID,
	}

	svc := newProjectCharacterizationServices(
		newProjectRepo(),
		newProjectControlCabinetRepo(),
		spsLinks,
		fieldDeviceLinks,
		newProjectObjectDataRepo(),
		bacnetRepo,
		specRepo,
		nil,
		spsRepo,
		spsSystemRepo,
		fieldDeviceRepo,
		nil,
	).FacilityLink

	if err := svc.DeleteSPSController(ctx, linkID, projectID); err != nil {
		t.Fatalf("expected delete to succeed, got %v", err)
	}

	if len(spsLinks.items) != 0 || len(fieldDeviceLinks.items) != 0 {
		t.Fatalf("expected project links to be removed, got sps=%d fd=%d", len(spsLinks.items), len(fieldDeviceLinks.items))
	}
	if _, ok := spsRepo.items[spsID]; ok {
		t.Fatal("expected copied/original sps controller to be deleted by current behavior")
	}
	if _, ok := spsSystemRepo.items[systemTypeID]; ok {
		t.Fatal("expected descendant sps controller system type to be deleted by current behavior")
	}
	if _, ok := fieldDeviceRepo.items[fieldDeviceID]; ok {
		t.Fatal("expected descendant field device to be deleted by current behavior")
	}
	if !sameUUIDSet(bacnetRepo.deletedFieldDeviceIDs, []uuid.UUID{fieldDeviceID}) {
		t.Fatalf("expected bacnet objects to be deleted for field device %s, got %v", fieldDeviceID, bacnetRepo.deletedFieldDeviceIDs)
	}
	if !sameUUIDSet(specRepo.deletedFieldDeviceIDs, []uuid.UUID{fieldDeviceID}) {
		t.Fatalf("expected specifications to be deleted for field device %s, got %v", fieldDeviceID, specRepo.deletedFieldDeviceIDs)
	}
}
