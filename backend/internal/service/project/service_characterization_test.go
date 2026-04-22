package project

import (
	"context"
	"sort"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/google/uuid"
)

func TestProjectService_Create_CharacterizesTemplateCopyAndCreatorMembership(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	creatorID := uuid.New()
	templateID := uuid.New()
	templateBO1ID := uuid.New()
	templateBO2ID := uuid.New()
	description := "Template Object"

	projectRepo := newProjectRepo()
	objectDataRepo := newProjectObjectDataRepo()
	bacnetRepo := newProjectBacnetObjectRepo()
	objectDataRepo.templates = []*domainFacility.ObjectData{
		{
			Base:        domain.Base{ID: templateID},
			Description: description,
			Version:     "1",
			IsActive:    true,
			BacnetObjects: []*domainFacility.BacnetObject{
				{
					Base:           domain.Base{ID: templateBO1ID},
					TextFix:        "BO-1",
					SoftwareType:   domainFacility.BacnetSoftwareTypeBI,
					SoftwareNumber: 1,
				},
				{
					Base:                domain.Base{ID: templateBO2ID},
					TextFix:             "BO-2",
					SoftwareType:        domainFacility.BacnetSoftwareTypeBV,
					SoftwareNumber:      2,
					SoftwareReferenceID: &templateBO1ID,
				},
			},
		},
	}

	svc := New(
		projectRepo,
		newProjectControlCabinetRepo(),
		newProjectSPSControllerRepo(),
		newProjectFieldDeviceRepo(),
		nil,
		nil,
		objectDataRepo,
		bacnetRepo,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	project := &domainProject.Project{
		Base:      domain.Base{ID: projectID},
		Name:      "Current project",
		CreatorID: creatorID,
	}

	if err := svc.Create(ctx, project); err != nil {
		t.Fatalf("expected create to succeed, got %v", err)
	}

	if project.Status != domainProject.StatusPlanned {
		t.Fatalf("expected empty project status to be defaulted to planned, got %q", project.Status)
	}
	if !projectRepo.hasUser(projectID, creatorID) {
		t.Fatal("expected creator to be added as a project user")
	}
	if len(objectDataRepo.created) != 1 {
		t.Fatalf("expected one project object-data copy, got %d", len(objectDataRepo.created))
	}
	copiedObjectData := objectDataRepo.created[0]
	if copiedObjectData.ID == templateID || copiedObjectData.ID == uuid.Nil {
		t.Fatalf("expected copied object-data to get a new non-template id, got %s", copiedObjectData.ID)
	}
	if copiedObjectData.ProjectID == nil || *copiedObjectData.ProjectID != projectID {
		t.Fatalf("expected copied object-data to belong to project %s, got %+v", projectID, copiedObjectData.ProjectID)
	}
	if copiedObjectData.Description != description {
		t.Fatalf("expected copied object-data description %q, got %q", description, copiedObjectData.Description)
	}
	if len(bacnetRepo.items) != 2 {
		t.Fatalf("expected two copied bacnet objects, got %d", len(bacnetRepo.items))
	}

	var referencedCopyID uuid.UUID
	for _, item := range bacnetRepo.items {
		if item.TextFix == "BO-2" && item.SoftwareReferenceID != nil {
			referencedCopyID = *item.SoftwareReferenceID
		}
		if item.ID == templateBO1ID || item.ID == templateBO2ID {
			t.Fatalf("expected bacnet object copy to get a new id, got template id %s", item.ID)
		}
	}
	if referencedCopyID == uuid.Nil {
		t.Fatal("expected copied BO-2 to retain an internal software reference")
	}
	referencedCopy, ok := bacnetRepo.items[referencedCopyID]
	if !ok || referencedCopy.TextFix != "BO-1" {
		t.Fatalf("expected copied BO-2 to reference copied BO-1, got %+v", referencedCopy)
	}
	if len(objectDataRepo.updated) != 1 || len(objectDataRepo.updated[0].BacnetObjects) != 2 {
		t.Fatalf("expected copied object-data to be updated with two copied bacnet objects, got %+v", objectDataRepo.updated)
	}
}

func TestProjectService_CreateControlCabinet_CharacterizesDescendantLinking(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	controlCabinetID := uuid.New()
	spsOneID := uuid.New()
	spsTwoID := uuid.New()
	systemTypeOneID := uuid.New()
	systemTypeTwoID := uuid.New()
	fieldDeviceOneID := uuid.New()
	fieldDeviceTwoID := uuid.New()

	controlCabinetLinks := newProjectControlCabinetRepo()
	spsLinks := newProjectSPSControllerRepo()
	fieldDeviceLinks := newProjectFieldDeviceRepo()
	spsRepo := newProjectSPSRepo()
	spsSystemRepo := newProjectSPSSystemTypeRepo()
	fieldDeviceRepo := newProjectFieldDeviceStore()

	spsRepo.items[spsOneID] = &domainFacility.SPSController{Base: domain.Base{ID: spsOneID}, ControlCabinetID: controlCabinetID}
	spsRepo.items[spsTwoID] = &domainFacility.SPSController{Base: domain.Base{ID: spsTwoID}, ControlCabinetID: controlCabinetID}
	spsSystemRepo.items[systemTypeOneID] = &domainFacility.SPSControllerSystemType{Base: domain.Base{ID: systemTypeOneID}, SPSControllerID: spsOneID}
	spsSystemRepo.items[systemTypeTwoID] = &domainFacility.SPSControllerSystemType{Base: domain.Base{ID: systemTypeTwoID}, SPSControllerID: spsTwoID}
	fieldDeviceRepo.items[fieldDeviceOneID] = &domainFacility.FieldDevice{Base: domain.Base{ID: fieldDeviceOneID}, SPSControllerSystemTypeID: systemTypeOneID}
	fieldDeviceRepo.items[fieldDeviceTwoID] = &domainFacility.FieldDevice{Base: domain.Base{ID: fieldDeviceTwoID}, SPSControllerSystemTypeID: systemTypeTwoID}

	svc := newProjectCharacterizationService(
		newProjectRepo(),
		controlCabinetLinks,
		spsLinks,
		fieldDeviceLinks,
		newProjectObjectDataRepo(),
		newProjectBacnetObjectRepo(),
		nil,
		nil,
		spsRepo,
		spsSystemRepo,
		fieldDeviceRepo,
		nil,
	)

	created, err := svc.CreateControlCabinet(ctx, projectID, controlCabinetID)
	if err != nil {
		t.Fatalf("expected link creation to succeed, got %v", err)
	}

	if created.ProjectID != projectID || created.ControlCabinetID != controlCabinetID {
		t.Fatalf("expected returned control cabinet link to preserve input ids, got %+v", created)
	}
	if got := controlCabinetLinks.controlCabinetIDs(projectID); !sameUUIDSet(got, []uuid.UUID{controlCabinetID}) {
		t.Fatalf("expected project control cabinet links to contain only %s, got %v", controlCabinetID, got)
	}
	if got := spsLinks.spsControllerIDs(projectID); !sameUUIDSet(got, []uuid.UUID{spsOneID, spsTwoID}) {
		t.Fatalf("expected descendant sps controller links, got %v", got)
	}
	if got := fieldDeviceLinks.fieldDeviceIDs(projectID); !sameUUIDSet(got, []uuid.UUID{fieldDeviceOneID, fieldDeviceTwoID}) {
		t.Fatalf("expected descendant field device links, got %v", got)
	}
}

func TestProjectService_DeleteControlCabinet_CharacterizesLinkAndHierarchyDeletion(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	controlCabinetID := uuid.New()
	linkID := uuid.New()
	spsID := uuid.New()
	systemTypeID := uuid.New()
	fieldDeviceID := uuid.New()

	controlCabinetLinks := newProjectControlCabinetRepo()
	spsLinks := newProjectSPSControllerRepo()
	fieldDeviceLinks := newProjectFieldDeviceRepo()
	controlCabinetRepo := newProjectControlCabinetStore()
	spsRepo := newProjectSPSRepo()
	spsSystemRepo := newProjectSPSSystemTypeRepo()
	fieldDeviceRepo := newProjectFieldDeviceStore()
	specRepo := newProjectSpecificationRepo()
	bacnetRepo := newProjectBacnetObjectRepo()

	controlCabinetLinks.items[linkID] = &domainProject.ProjectControlCabinet{
		Base:             domain.Base{ID: linkID},
		ProjectID:        projectID,
		ControlCabinetID: controlCabinetID,
	}
	spsLinks.createWithID(projectID, spsID)
	fieldDeviceLinks.createWithID(projectID, fieldDeviceID)
	controlCabinetRepo.items[controlCabinetID] = &domainFacility.ControlCabinet{Base: domain.Base{ID: controlCabinetID}}
	spsRepo.items[spsID] = &domainFacility.SPSController{Base: domain.Base{ID: spsID}, ControlCabinetID: controlCabinetID}
	spsSystemRepo.items[systemTypeID] = &domainFacility.SPSControllerSystemType{Base: domain.Base{ID: systemTypeID}, SPSControllerID: spsID}
	fieldDeviceRepo.items[fieldDeviceID] = &domainFacility.FieldDevice{Base: domain.Base{ID: fieldDeviceID}, SPSControllerSystemTypeID: systemTypeID}

	svc := newProjectCharacterizationService(
		newProjectRepo(),
		controlCabinetLinks,
		spsLinks,
		fieldDeviceLinks,
		newProjectObjectDataRepo(),
		bacnetRepo,
		specRepo,
		controlCabinetRepo,
		spsRepo,
		spsSystemRepo,
		fieldDeviceRepo,
		nil,
	)

	if err := svc.DeleteControlCabinet(ctx, linkID, projectID); err != nil {
		t.Fatalf("expected delete to succeed, got %v", err)
	}

	if len(controlCabinetLinks.items) != 0 || len(spsLinks.items) != 0 || len(fieldDeviceLinks.items) != 0 {
		t.Fatalf("expected project links to be removed, got cc=%d sps=%d fd=%d", len(controlCabinetLinks.items), len(spsLinks.items), len(fieldDeviceLinks.items))
	}
	if _, ok := controlCabinetRepo.items[controlCabinetID]; ok {
		t.Fatal("expected copied/original control cabinet to be deleted by current behavior")
	}
	if _, ok := spsRepo.items[spsID]; ok {
		t.Fatal("expected descendant sps controller to be deleted by current behavior")
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

func TestProjectService_CopySPSControllerSystemType_CharacterizesCopiedFieldDeviceLinking(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	spsID := uuid.New()
	systemTypeDefinitionID := uuid.New()
	originalSystemTypeID := uuid.New()
	originalFieldDeviceID := uuid.New()
	apparatID := uuid.New()

	spsSystemRepo := newProjectSPSSystemTypeRepo()
	fieldDeviceRepo := newProjectFieldDeviceStore()
	fieldDeviceLinks := newProjectFieldDeviceRepo()
	systemTypeRepo := newProjectSystemTypeRepo()
	specRepo := newProjectSpecificationRepo()
	bacnetRepo := newProjectBacnetObjectRepo()
	originalNumber := 1

	systemTypeRepo.items[systemTypeDefinitionID] = &domainFacility.SystemType{
		Base:      domain.Base{ID: systemTypeDefinitionID},
		NumberMin: 1,
		NumberMax: 2,
	}
	spsSystemRepo.items[originalSystemTypeID] = &domainFacility.SPSControllerSystemType{
		Base:            domain.Base{ID: originalSystemTypeID},
		Number:          &originalNumber,
		SPSControllerID: spsID,
		SystemTypeID:    systemTypeDefinitionID,
	}
	fieldDeviceRepo.items[originalFieldDeviceID] = &domainFacility.FieldDevice{
		Base:                      domain.Base{ID: originalFieldDeviceID},
		SPSControllerSystemTypeID: originalSystemTypeID,
		ApparatID:                 apparatID,
		ApparatNr:                 7,
	}

	hierarchyCopier := facilityservice.NewHierarchyCopier(
		newProjectControlCabinetStore(),
		newProjectBuildingRepo(),
		newProjectSPSRepo(),
		systemTypeRepo,
		spsSystemRepo,
		fieldDeviceRepo,
		specRepo,
		bacnetRepo,
	)
	svc := newProjectCharacterizationService(
		newProjectRepo(),
		newProjectControlCabinetRepo(),
		newProjectSPSControllerRepo(),
		fieldDeviceLinks,
		newProjectObjectDataRepo(),
		bacnetRepo,
		specRepo,
		newProjectControlCabinetStore(),
		newProjectSPSRepo(),
		spsSystemRepo,
		fieldDeviceRepo,
		hierarchyCopier,
	)

	copiedSystemType, err := svc.CopySPSControllerSystemType(ctx, projectID, originalSystemTypeID)
	if err != nil {
		t.Fatalf("expected system type copy to succeed, got %v", err)
	}

	if copiedSystemType.ID == originalSystemTypeID || copiedSystemType.ID == uuid.Nil {
		t.Fatalf("expected a new system type id, got %s", copiedSystemType.ID)
	}
	if copiedSystemType.Number == nil || *copiedSystemType.Number != 2 {
		t.Fatalf("expected copied system type to take next available number 2, got %+v", copiedSystemType.Number)
	}
	copiedFieldDeviceIDs := fieldDeviceRepo.idsForSystemType(copiedSystemType.ID)
	if len(copiedFieldDeviceIDs) != 1 {
		t.Fatalf("expected exactly one copied field device under copied system type, got %v", copiedFieldDeviceIDs)
	}
	if got := fieldDeviceLinks.fieldDeviceIDs(projectID); !sameUUIDSet(got, copiedFieldDeviceIDs) {
		t.Fatalf("expected project to link copied field device %v, got %v", copiedFieldDeviceIDs, got)
	}
}

func TestProjectService_CopyControlCabinet_CharacterizesDeepCopyAndProjectLinks(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	buildingID := uuid.New()
	originalControlCabinetID := uuid.New()
	originalSPSID := uuid.New()
	systemTypeDefinitionID := uuid.New()
	originalSystemTypeID := uuid.New()
	originalFieldDeviceID := uuid.New()
	apparatID := uuid.New()
	cabinetNr := "AK1"
	gaDevice := "A01"
	originalNumber := 1

	controlCabinetLinks := newProjectControlCabinetRepo()
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
	controlCabinetRepo.items[originalControlCabinetID] = &domainFacility.ControlCabinet{
		Base:             domain.Base{ID: originalControlCabinetID},
		BuildingID:       buildingID,
		ControlCabinetNr: &cabinetNr,
	}
	spsRepo.items[originalSPSID] = &domainFacility.SPSController{
		Base:             domain.Base{ID: originalSPSID},
		ControlCabinetID: originalControlCabinetID,
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
	svc := newProjectCharacterizationService(
		newProjectRepo(),
		controlCabinetLinks,
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
	)

	copiedControlCabinet, err := svc.CopyControlCabinet(ctx, projectID, originalControlCabinetID)
	if err != nil {
		t.Fatalf("expected control cabinet copy to succeed, got %v", err)
	}

	if copiedControlCabinet.ID == uuid.Nil || copiedControlCabinet.ID == originalControlCabinetID {
		t.Fatalf("expected copied control cabinet to get a new id, got %s", copiedControlCabinet.ID)
	}
	if got := controlCabinetLinks.controlCabinetIDs(projectID); !sameUUIDSet(got, []uuid.UUID{copiedControlCabinet.ID}) {
		t.Fatalf("expected project to link copied control cabinet %s, got %v", copiedControlCabinet.ID, got)
	}

	copiedSPSIDs, err := spsRepo.GetIDsByControlCabinetID(ctx, copiedControlCabinet.ID)
	if err != nil {
		t.Fatalf("expected copied sps lookup to succeed, got %v", err)
	}
	if len(copiedSPSIDs) != 1 || copiedSPSIDs[0] == originalSPSID {
		t.Fatalf("expected one copied sps controller, got %v", copiedSPSIDs)
	}
	if got := spsLinks.spsControllerIDs(projectID); !sameUUIDSet(got, copiedSPSIDs) {
		t.Fatalf("expected project to link copied sps controllers %v, got %v", copiedSPSIDs, got)
	}

	copiedSystemTypeIDs, err := spsSystemRepo.GetIDsBySPSControllerIDs(ctx, copiedSPSIDs)
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

func TestProjectService_AddAndRemoveObjectData_CharacterizesProjectActivation(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	objectDataID := uuid.New()

	projectRepo := newProjectRepo()
	projectRepo.items[projectID] = &domainProject.Project{Base: domain.Base{ID: projectID}, Name: "Current"}
	objectDataRepo := newProjectObjectDataRepo()
	objectDataRepo.items[objectDataID] = &domainFacility.ObjectData{
		Base:        domain.Base{ID: objectDataID},
		Description: "OD",
		Version:     "1",
		IsActive:    false,
	}
	svc := newProjectCharacterizationService(
		projectRepo,
		newProjectControlCabinetRepo(),
		newProjectSPSControllerRepo(),
		newProjectFieldDeviceRepo(),
		objectDataRepo,
		newProjectBacnetObjectRepo(),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	added, err := svc.AddObjectData(ctx, projectID, objectDataID)
	if err != nil {
		t.Fatalf("expected add object data to succeed, got %v", err)
	}
	if added.ProjectID == nil || *added.ProjectID != projectID || !added.IsActive {
		t.Fatalf("expected add to assign project and activate object data, got %+v", added)
	}

	removed, err := svc.RemoveObjectData(ctx, projectID, objectDataID)
	if err != nil {
		t.Fatalf("expected remove object data to succeed, got %v", err)
	}
	if removed.ProjectID == nil || *removed.ProjectID != projectID || removed.IsActive {
		t.Fatalf("expected remove to keep project assignment but mark inactive, got %+v", removed)
	}
}

func TestProjectService_ListObjectData_CharacterizesProjectFilterRouting(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	apparatID := uuid.New()
	systemPartID := uuid.New()
	objectDataRepo := newProjectObjectDataRepo()
	svc := newProjectCharacterizationService(
		newProjectRepo(),
		newProjectControlCabinetRepo(),
		newProjectSPSControllerRepo(),
		newProjectFieldDeviceRepo(),
		objectDataRepo,
		newProjectBacnetObjectRepo(),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	if _, err := svc.ListObjectData(ctx, projectID, 0, 0, "pump", nil, nil); err != nil {
		t.Fatalf("expected unfiltered list to succeed, got %v", err)
	}
	if objectDataRepo.lastListMethod != "project" || objectDataRepo.lastPagination.Page != 1 || objectDataRepo.lastPagination.Limit != 10 || objectDataRepo.lastPagination.Search != "pump" {
		t.Fatalf("expected unfiltered project list with normalized pagination, got method=%s params=%+v", objectDataRepo.lastListMethod, objectDataRepo.lastPagination)
	}

	if _, err := svc.ListObjectData(ctx, projectID, 2, 25, "", &apparatID, nil); err != nil {
		t.Fatalf("expected apparat-filtered list to succeed, got %v", err)
	}
	if objectDataRepo.lastListMethod != "project_apparat" {
		t.Fatalf("expected apparat filter route, got %s", objectDataRepo.lastListMethod)
	}

	if _, err := svc.ListObjectData(ctx, projectID, 2, 25, "", nil, &systemPartID); err != nil {
		t.Fatalf("expected system-part-filtered list to succeed, got %v", err)
	}
	if objectDataRepo.lastListMethod != "project_system_part" {
		t.Fatalf("expected system-part filter route, got %s", objectDataRepo.lastListMethod)
	}

	if _, err := svc.ListObjectData(ctx, projectID, 2, 25, "", &apparatID, &systemPartID); err != nil {
		t.Fatalf("expected combined-filter list to succeed, got %v", err)
	}
	if objectDataRepo.lastListMethod != "project_apparat_system_part" {
		t.Fatalf("expected combined filter route, got %s", objectDataRepo.lastListMethod)
	}
}

func newProjectCharacterizationService(
	projectRepo *projectRepoFake,
	controlCabinetLinks *projectControlCabinetRepoFake,
	spsLinks *projectSPSControllerRepoFake,
	fieldDeviceLinks *projectFieldDeviceRepoFake,
	objectDataRepo *projectObjectDataRepoFake,
	bacnetRepo *projectBacnetObjectRepoFake,
	specRepo *projectSpecificationRepoFake,
	controlCabinetRepo *projectControlCabinetStoreFake,
	spsRepo *projectSPSRepoFake,
	spsSystemRepo *projectSPSSystemTypeRepoFake,
	fieldDeviceRepo *projectFieldDeviceStoreFake,
	hierarchyCopier *facilityservice.HierarchyCopier,
) *Service {
	return New(
		projectRepo,
		controlCabinetLinks,
		spsLinks,
		fieldDeviceLinks,
		nil,
		nil,
		objectDataRepo,
		bacnetRepo,
		specRepo,
		controlCabinetRepo,
		spsRepo,
		spsSystemRepo,
		fieldDeviceRepo,
		hierarchyCopier,
		nil,
		nil,
	)
}

type projectRepoFake struct {
	items map[uuid.UUID]*domainProject.Project
	users map[uuid.UUID]map[uuid.UUID]struct{}
}

func newProjectRepo() *projectRepoFake {
	return &projectRepoFake{items: map[uuid.UUID]*domainProject.Project{}, users: map[uuid.UUID]map[uuid.UUID]struct{}{}}
}

func (r *projectRepoFake) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainProject.Project, error) {
	out := make([]*domainProject.Project, 0, len(ids))
	for _, id := range ids {
		if item, ok := r.items[id]; ok {
			clone := *item
			out = append(out, &clone)
		}
	}
	return out, nil
}

func (r *projectRepoFake) Create(_ context.Context, entity *domainProject.Project) error {
	if entity.ID == uuid.Nil {
		entity.ID = uuid.New()
	}
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *projectRepoFake) Update(_ context.Context, entity *domainProject.Project) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *projectRepoFake) DeleteByIds(_ context.Context, ids []uuid.UUID) error {
	for _, id := range ids {
		delete(r.items, id)
	}
	return nil
}

func (r *projectRepoFake) GetPaginatedList(_ context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainProject.Project], error) {
	return paginatedFromMap(r.items), nil
}

func (r *projectRepoFake) GetPaginatedListForUser(context.Context, domain.PaginationParams, uuid.UUID) (*domain.PaginatedList[domainProject.Project], error) {
	return paginatedFromMap(r.items), nil
}

func (r *projectRepoFake) GetPaginatedListWithStatus(context.Context, domain.PaginationParams, *domainProject.ProjectStatus) (*domain.PaginatedList[domainProject.Project], error) {
	return paginatedFromMap(r.items), nil
}

func (r *projectRepoFake) GetPaginatedListForUserWithStatus(context.Context, domain.PaginationParams, uuid.UUID, *domainProject.ProjectStatus) (*domain.PaginatedList[domainProject.Project], error) {
	return paginatedFromMap(r.items), nil
}

func (r *projectRepoFake) HasUser(_ context.Context, projectID, userID uuid.UUID) (bool, error) {
	return r.hasUser(projectID, userID), nil
}

func (r *projectRepoFake) AddUser(_ context.Context, projectID, userID uuid.UUID) error {
	if r.users[projectID] == nil {
		r.users[projectID] = map[uuid.UUID]struct{}{}
	}
	r.users[projectID][userID] = struct{}{}
	return nil
}

func (r *projectRepoFake) RemoveUser(_ context.Context, projectID, userID uuid.UUID) error {
	delete(r.users[projectID], userID)
	return nil
}

func (r *projectRepoFake) ListUsers(context.Context, uuid.UUID) ([]domainUser.User, error) {
	return nil, nil
}

func (r *projectRepoFake) hasUser(projectID, userID uuid.UUID) bool {
	_, ok := r.users[projectID][userID]
	return ok
}

type projectControlCabinetRepoFake struct {
	items map[uuid.UUID]*domainProject.ProjectControlCabinet
}

func newProjectControlCabinetRepo() *projectControlCabinetRepoFake {
	return &projectControlCabinetRepoFake{items: map[uuid.UUID]*domainProject.ProjectControlCabinet{}}
}

func (r *projectControlCabinetRepoFake) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainProject.ProjectControlCabinet, error) {
	return getProjectItemsByID(ids, r.items), nil
}

func (r *projectControlCabinetRepoFake) Create(_ context.Context, entity *domainProject.ProjectControlCabinet) error {
	if entity.ID == uuid.Nil {
		entity.ID = uuid.New()
	}
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *projectControlCabinetRepoFake) Update(_ context.Context, entity *domainProject.ProjectControlCabinet) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *projectControlCabinetRepoFake) DeleteByIds(_ context.Context, ids []uuid.UUID) error {
	deleteIDs(ids, r.items)
	return nil
}

func (r *projectControlCabinetRepoFake) GetPaginatedList(context.Context, domain.PaginationParams) (*domain.PaginatedList[domainProject.ProjectControlCabinet], error) {
	return paginatedFromMap(r.items), nil
}

func (r *projectControlCabinetRepoFake) GetPaginatedListByProjectID(_ context.Context, projectID uuid.UUID, _ domain.PaginationParams) (*domain.PaginatedList[domainProject.ProjectControlCabinet], error) {
	return paginatedFromFilter(r.items, func(item *domainProject.ProjectControlCabinet) bool { return item.ProjectID == projectID }), nil
}

func (r *projectControlCabinetRepoFake) GetByControlCabinetID(_ context.Context, controlCabinetID uuid.UUID) ([]*domainProject.ProjectControlCabinet, error) {
	out := make([]*domainProject.ProjectControlCabinet, 0)
	for _, item := range r.items {
		if item.ControlCabinetID == controlCabinetID {
			clone := *item
			out = append(out, &clone)
		}
	}
	return out, nil
}

func (r *projectControlCabinetRepoFake) controlCabinetIDs(projectID uuid.UUID) []uuid.UUID {
	out := make([]uuid.UUID, 0)
	for _, item := range r.items {
		if item.ProjectID == projectID {
			out = append(out, item.ControlCabinetID)
		}
	}
	return out
}

type projectSPSControllerRepoFake struct {
	items map[uuid.UUID]*domainProject.ProjectSPSController
}

func newProjectSPSControllerRepo() *projectSPSControllerRepoFake {
	return &projectSPSControllerRepoFake{items: map[uuid.UUID]*domainProject.ProjectSPSController{}}
}

func (r *projectSPSControllerRepoFake) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainProject.ProjectSPSController, error) {
	return getProjectItemsByID(ids, r.items), nil
}

func (r *projectSPSControllerRepoFake) Create(_ context.Context, entity *domainProject.ProjectSPSController) error {
	if entity.ID == uuid.Nil {
		entity.ID = uuid.New()
	}
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *projectSPSControllerRepoFake) Update(_ context.Context, entity *domainProject.ProjectSPSController) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *projectSPSControllerRepoFake) DeleteByIds(_ context.Context, ids []uuid.UUID) error {
	deleteIDs(ids, r.items)
	return nil
}

func (r *projectSPSControllerRepoFake) GetPaginatedList(context.Context, domain.PaginationParams) (*domain.PaginatedList[domainProject.ProjectSPSController], error) {
	return paginatedFromMap(r.items), nil
}

func (r *projectSPSControllerRepoFake) GetPaginatedListByProjectID(_ context.Context, projectID uuid.UUID, _ domain.PaginationParams) (*domain.PaginatedList[domainProject.ProjectSPSController], error) {
	return paginatedFromFilter(r.items, func(item *domainProject.ProjectSPSController) bool { return item.ProjectID == projectID }), nil
}

func (r *projectSPSControllerRepoFake) GetBySPSControllerID(_ context.Context, spsControllerID uuid.UUID) ([]*domainProject.ProjectSPSController, error) {
	out := make([]*domainProject.ProjectSPSController, 0)
	for _, item := range r.items {
		if item.SPSControllerID == spsControllerID {
			clone := *item
			out = append(out, &clone)
		}
	}
	return out, nil
}

func (r *projectSPSControllerRepoFake) createWithID(projectID, spsControllerID uuid.UUID) {
	_ = r.Create(context.Background(), &domainProject.ProjectSPSController{ProjectID: projectID, SPSControllerID: spsControllerID})
}

func (r *projectSPSControllerRepoFake) spsControllerIDs(projectID uuid.UUID) []uuid.UUID {
	out := make([]uuid.UUID, 0)
	for _, item := range r.items {
		if item.ProjectID == projectID {
			out = append(out, item.SPSControllerID)
		}
	}
	return out
}

type projectFieldDeviceRepoFake struct {
	items map[uuid.UUID]*domainProject.ProjectFieldDevice
}

func newProjectFieldDeviceRepo() *projectFieldDeviceRepoFake {
	return &projectFieldDeviceRepoFake{items: map[uuid.UUID]*domainProject.ProjectFieldDevice{}}
}

func (r *projectFieldDeviceRepoFake) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainProject.ProjectFieldDevice, error) {
	return getProjectItemsByID(ids, r.items), nil
}

func (r *projectFieldDeviceRepoFake) Create(_ context.Context, entity *domainProject.ProjectFieldDevice) error {
	if entity.ID == uuid.Nil {
		entity.ID = uuid.New()
	}
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *projectFieldDeviceRepoFake) Update(_ context.Context, entity *domainProject.ProjectFieldDevice) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *projectFieldDeviceRepoFake) DeleteByIds(_ context.Context, ids []uuid.UUID) error {
	deleteIDs(ids, r.items)
	return nil
}

func (r *projectFieldDeviceRepoFake) GetPaginatedList(context.Context, domain.PaginationParams) (*domain.PaginatedList[domainProject.ProjectFieldDevice], error) {
	return paginatedFromMap(r.items), nil
}

func (r *projectFieldDeviceRepoFake) GetPaginatedListByProjectID(_ context.Context, projectID uuid.UUID, _ domain.PaginationParams) (*domain.PaginatedList[domainProject.ProjectFieldDevice], error) {
	return paginatedFromFilter(r.items, func(item *domainProject.ProjectFieldDevice) bool { return item.ProjectID == projectID }), nil
}

func (r *projectFieldDeviceRepoFake) createWithID(projectID, fieldDeviceID uuid.UUID) {
	_ = r.Create(context.Background(), &domainProject.ProjectFieldDevice{ProjectID: projectID, FieldDeviceID: fieldDeviceID})
}

func (r *projectFieldDeviceRepoFake) fieldDeviceIDs(projectID uuid.UUID) []uuid.UUID {
	out := make([]uuid.UUID, 0)
	for _, item := range r.items {
		if item.ProjectID == projectID {
			out = append(out, item.FieldDeviceID)
		}
	}
	return out
}

type projectObjectDataRepoFake struct {
	items     map[uuid.UUID]*domainFacility.ObjectData
	templates []*domainFacility.ObjectData
	created   []*domainFacility.ObjectData
	updated   []*domainFacility.ObjectData

	lastListMethod string
	lastPagination domain.PaginationParams
}

func newProjectObjectDataRepo() *projectObjectDataRepoFake {
	return &projectObjectDataRepoFake{items: map[uuid.UUID]*domainFacility.ObjectData{}}
}

func (r *projectObjectDataRepoFake) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainFacility.ObjectData, error) {
	return getProjectItemsByID(ids, r.items), nil
}

func (r *projectObjectDataRepoFake) Create(_ context.Context, entity *domainFacility.ObjectData) error {
	if entity.ID == uuid.Nil {
		entity.ID = uuid.New()
	}
	clone := cloneObjectData(entity)
	r.items[entity.ID] = clone
	r.created = append(r.created, cloneObjectData(entity))
	return nil
}

func (r *projectObjectDataRepoFake) Update(_ context.Context, entity *domainFacility.ObjectData) error {
	clone := cloneObjectData(entity)
	r.items[entity.ID] = clone
	r.updated = append(r.updated, cloneObjectData(entity))
	return nil
}

func (r *projectObjectDataRepoFake) DeleteByIds(_ context.Context, ids []uuid.UUID) error {
	deleteIDs(ids, r.items)
	return nil
}

func (r *projectObjectDataRepoFake) GetPaginatedList(context.Context, domain.PaginationParams) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	return paginatedFromMap(r.items), nil
}

func (r *projectObjectDataRepoFake) GetBacnetObjectIDs(context.Context, uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

func (r *projectObjectDataRepoFake) ExistsByDescription(context.Context, *uuid.UUID, string, *uuid.UUID) (bool, error) {
	return false, nil
}

func (r *projectObjectDataRepoFake) GetTemplates(context.Context) ([]*domainFacility.ObjectData, error) {
	out := make([]*domainFacility.ObjectData, len(r.templates))
	for i, item := range r.templates {
		out[i] = cloneObjectData(item)
	}
	return out, nil
}

func (r *projectObjectDataRepoFake) GetTemplatesLite(context.Context) ([]*domainFacility.ObjectData, error) {
	return r.GetTemplates(context.Background())
}

func (r *projectObjectDataRepoFake) GetForProject(_ context.Context, projectID uuid.UUID) ([]*domainFacility.ObjectData, error) {
	return objectDataForProject(r.items, projectID), nil
}

func (r *projectObjectDataRepoFake) GetForProjectLite(ctx context.Context, projectID uuid.UUID) ([]*domainFacility.ObjectData, error) {
	return r.GetForProject(ctx, projectID)
}

func (r *projectObjectDataRepoFake) GetPaginatedListForProject(_ context.Context, projectID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	r.lastListMethod = "project"
	r.lastPagination = params
	return paginatedObjectData(objectDataForProject(r.items, projectID)), nil
}

func (r *projectObjectDataRepoFake) GetPaginatedListByApparatID(context.Context, uuid.UUID, domain.PaginationParams) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	return paginatedObjectData(nil), nil
}

func (r *projectObjectDataRepoFake) GetPaginatedListBySystemPartID(context.Context, uuid.UUID, domain.PaginationParams) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	return paginatedObjectData(nil), nil
}

func (r *projectObjectDataRepoFake) GetPaginatedListByApparatAndSystemPartID(context.Context, uuid.UUID, uuid.UUID, domain.PaginationParams) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	return paginatedObjectData(nil), nil
}

func (r *projectObjectDataRepoFake) GetPaginatedListForProjectByApparatID(ctx context.Context, projectID, _ uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	r.lastListMethod = "project_apparat"
	r.lastPagination = params
	return paginatedObjectData(objectDataForProject(r.items, projectID)), nil
}

func (r *projectObjectDataRepoFake) GetPaginatedListForProjectBySystemPartID(ctx context.Context, projectID, _ uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	r.lastListMethod = "project_system_part"
	r.lastPagination = params
	return paginatedObjectData(objectDataForProject(r.items, projectID)), nil
}

func (r *projectObjectDataRepoFake) GetPaginatedListForProjectByApparatAndSystemPartID(ctx context.Context, projectID, _, _ uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	r.lastListMethod = "project_apparat_system_part"
	r.lastPagination = params
	return paginatedObjectData(objectDataForProject(r.items, projectID)), nil
}

type projectBacnetObjectRepoFake struct {
	items                 map[uuid.UUID]*domainFacility.BacnetObject
	deletedFieldDeviceIDs []uuid.UUID
}

func newProjectBacnetObjectRepo() *projectBacnetObjectRepoFake {
	return &projectBacnetObjectRepoFake{items: map[uuid.UUID]*domainFacility.BacnetObject{}}
}

func (r *projectBacnetObjectRepoFake) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainFacility.BacnetObject, error) {
	return getProjectItemsByID(ids, r.items), nil
}

func (r *projectBacnetObjectRepoFake) Create(_ context.Context, entity *domainFacility.BacnetObject) error {
	if entity.ID == uuid.Nil {
		entity.ID = uuid.New()
	}
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *projectBacnetObjectRepoFake) BulkCreate(_ context.Context, entities []*domainFacility.BacnetObject, _ int) error {
	for _, entity := range entities {
		if err := r.Create(context.Background(), entity); err != nil {
			return err
		}
	}
	return nil
}

func (r *projectBacnetObjectRepoFake) Update(_ context.Context, entity *domainFacility.BacnetObject) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *projectBacnetObjectRepoFake) DeleteByIds(_ context.Context, ids []uuid.UUID) error {
	deleteIDs(ids, r.items)
	return nil
}

func (r *projectBacnetObjectRepoFake) GetPaginatedList(context.Context, domain.PaginationParams) (*domain.PaginatedList[domainFacility.BacnetObject], error) {
	return paginatedFromMap(r.items), nil
}

func (r *projectBacnetObjectRepoFake) GetByFieldDeviceIDs(_ context.Context, ids []uuid.UUID) ([]*domainFacility.BacnetObject, error) {
	idSet := uuidSet(ids)
	out := make([]*domainFacility.BacnetObject, 0)
	for _, item := range r.items {
		if item.FieldDeviceID != nil {
			if _, ok := idSet[*item.FieldDeviceID]; ok {
				clone := *item
				out = append(out, &clone)
			}
		}
	}
	return out, nil
}

func (r *projectBacnetObjectRepoFake) DeleteByFieldDeviceIDs(_ context.Context, ids []uuid.UUID) error {
	r.deletedFieldDeviceIDs = append(r.deletedFieldDeviceIDs, ids...)
	idSet := uuidSet(ids)
	for id, item := range r.items {
		if item.FieldDeviceID != nil {
			if _, ok := idSet[*item.FieldDeviceID]; ok {
				delete(r.items, id)
			}
		}
	}
	return nil
}

type projectSpecificationRepoFake struct {
	items                 map[uuid.UUID]*domainFacility.Specification
	deletedFieldDeviceIDs []uuid.UUID
}

func newProjectSpecificationRepo() *projectSpecificationRepoFake {
	return &projectSpecificationRepoFake{items: map[uuid.UUID]*domainFacility.Specification{}}
}

func (r *projectSpecificationRepoFake) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainFacility.Specification, error) {
	return getProjectItemsByID(ids, r.items), nil
}

func (r *projectSpecificationRepoFake) Create(_ context.Context, entity *domainFacility.Specification) error {
	if entity.ID == uuid.Nil {
		entity.ID = uuid.New()
	}
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *projectSpecificationRepoFake) BulkCreate(_ context.Context, entities []*domainFacility.Specification, _ int) error {
	for _, entity := range entities {
		if err := r.Create(context.Background(), entity); err != nil {
			return err
		}
	}
	return nil
}

func (r *projectSpecificationRepoFake) Update(_ context.Context, entity *domainFacility.Specification) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *projectSpecificationRepoFake) DeleteByIds(_ context.Context, ids []uuid.UUID) error {
	deleteIDs(ids, r.items)
	return nil
}

func (r *projectSpecificationRepoFake) GetPaginatedList(context.Context, domain.PaginationParams) (*domain.PaginatedList[domainFacility.Specification], error) {
	return paginatedFromMap(r.items), nil
}

func (r *projectSpecificationRepoFake) GetByFieldDeviceIDs(_ context.Context, ids []uuid.UUID) ([]*domainFacility.Specification, error) {
	idSet := uuidSet(ids)
	out := make([]*domainFacility.Specification, 0)
	for _, item := range r.items {
		if item.FieldDeviceID != nil {
			if _, ok := idSet[*item.FieldDeviceID]; ok {
				clone := *item
				out = append(out, &clone)
			}
		}
	}
	return out, nil
}

func (r *projectSpecificationRepoFake) DeleteByFieldDeviceIDs(_ context.Context, ids []uuid.UUID) error {
	r.deletedFieldDeviceIDs = append(r.deletedFieldDeviceIDs, ids...)
	idSet := uuidSet(ids)
	for id, item := range r.items {
		if item.FieldDeviceID != nil {
			if _, ok := idSet[*item.FieldDeviceID]; ok {
				delete(r.items, id)
			}
		}
	}
	return nil
}

type projectControlCabinetStoreFake struct {
	items map[uuid.UUID]*domainFacility.ControlCabinet
}

func newProjectControlCabinetStore() *projectControlCabinetStoreFake {
	return &projectControlCabinetStoreFake{items: map[uuid.UUID]*domainFacility.ControlCabinet{}}
}

func (r *projectControlCabinetStoreFake) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainFacility.ControlCabinet, error) {
	return getProjectItemsByID(ids, r.items), nil
}

func (r *projectControlCabinetStoreFake) Create(_ context.Context, entity *domainFacility.ControlCabinet) error {
	if entity.ID == uuid.Nil {
		entity.ID = uuid.New()
	}
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *projectControlCabinetStoreFake) Update(_ context.Context, entity *domainFacility.ControlCabinet) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *projectControlCabinetStoreFake) DeleteByIds(_ context.Context, ids []uuid.UUID) error {
	deleteIDs(ids, r.items)
	return nil
}

func (r *projectControlCabinetStoreFake) GetPaginatedList(context.Context, domain.PaginationParams) (*domain.PaginatedList[domainFacility.ControlCabinet], error) {
	return paginatedFromMap(r.items), nil
}

func (r *projectControlCabinetStoreFake) GetPaginatedListByBuildingID(_ context.Context, buildingID uuid.UUID, _ domain.PaginationParams) (*domain.PaginatedList[domainFacility.ControlCabinet], error) {
	return paginatedFromFilter(r.items, func(item *domainFacility.ControlCabinet) bool { return item.BuildingID == buildingID }), nil
}

func (r *projectControlCabinetStoreFake) GetIDsByBuildingID(_ context.Context, buildingID uuid.UUID) ([]uuid.UUID, error) {
	out := make([]uuid.UUID, 0)
	for id, item := range r.items {
		if item.BuildingID == buildingID {
			out = append(out, id)
		}
	}
	return out, nil
}

func (r *projectControlCabinetStoreFake) ExistsControlCabinetNr(_ context.Context, buildingID uuid.UUID, controlCabinetNr string, excludeID *uuid.UUID) (bool, error) {
	for id, item := range r.items {
		if excludeID != nil && id == *excludeID {
			continue
		}
		if item.BuildingID == buildingID && item.ControlCabinetNr != nil && *item.ControlCabinetNr == controlCabinetNr {
			return true, nil
		}
	}
	return false, nil
}

type projectSPSRepoFake struct {
	items map[uuid.UUID]*domainFacility.SPSController
}

func newProjectSPSRepo() *projectSPSRepoFake {
	return &projectSPSRepoFake{items: map[uuid.UUID]*domainFacility.SPSController{}}
}

func (r *projectSPSRepoFake) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainFacility.SPSController, error) {
	return getProjectItemsByID(ids, r.items), nil
}

func (r *projectSPSRepoFake) Create(_ context.Context, entity *domainFacility.SPSController) error {
	if entity.ID == uuid.Nil {
		entity.ID = uuid.New()
	}
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *projectSPSRepoFake) Update(_ context.Context, entity *domainFacility.SPSController) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *projectSPSRepoFake) DeleteByIds(_ context.Context, ids []uuid.UUID) error {
	deleteIDs(ids, r.items)
	return nil
}

func (r *projectSPSRepoFake) GetPaginatedList(context.Context, domain.PaginationParams) (*domain.PaginatedList[domainFacility.SPSController], error) {
	return paginatedFromMap(r.items), nil
}

func (r *projectSPSRepoFake) GetPaginatedListByControlCabinetID(_ context.Context, controlCabinetID uuid.UUID, _ domain.PaginationParams) (*domain.PaginatedList[domainFacility.SPSController], error) {
	return paginatedFromFilter(r.items, func(item *domainFacility.SPSController) bool { return item.ControlCabinetID == controlCabinetID }), nil
}

func (r *projectSPSRepoFake) GetIDsByControlCabinetID(_ context.Context, controlCabinetID uuid.UUID) ([]uuid.UUID, error) {
	out := make([]uuid.UUID, 0)
	for id, item := range r.items {
		if item.ControlCabinetID == controlCabinetID {
			out = append(out, id)
		}
	}
	return out, nil
}

func (r *projectSPSRepoFake) GetIDsByControlCabinetIDs(_ context.Context, controlCabinetIDs []uuid.UUID) ([]uuid.UUID, error) {
	set := uuidSet(controlCabinetIDs)
	out := make([]uuid.UUID, 0)
	for id, item := range r.items {
		if _, ok := set[item.ControlCabinetID]; ok {
			out = append(out, id)
		}
	}
	return out, nil
}

func (r *projectSPSRepoFake) ListGADevicesByControlCabinetID(_ context.Context, controlCabinetID uuid.UUID) ([]string, error) {
	out := make([]string, 0)
	for _, item := range r.items {
		if item.ControlCabinetID == controlCabinetID && item.GADevice != nil {
			out = append(out, *item.GADevice)
		}
	}
	return out, nil
}

func (r *projectSPSRepoFake) ExistsGADevice(_ context.Context, controlCabinetID uuid.UUID, gaDevice string, excludeID *uuid.UUID) (bool, error) {
	for id, item := range r.items {
		if excludeID != nil && id == *excludeID {
			continue
		}
		if item.ControlCabinetID == controlCabinetID && item.GADevice != nil && *item.GADevice == gaDevice {
			return true, nil
		}
	}
	return false, nil
}

func (r *projectSPSRepoFake) ExistsIPAddressVlan(context.Context, string, string, *uuid.UUID) (bool, error) {
	return false, nil
}

func (r *projectSPSRepoFake) GetByIdsForExport(_ context.Context, ids []uuid.UUID) ([]domainFacility.SPSController, error) {
	ptrs, _ := r.GetByIds(context.Background(), ids)
	out := make([]domainFacility.SPSController, 0, len(ptrs))
	for _, item := range ptrs {
		out = append(out, *item)
	}
	return out, nil
}

type projectSPSSystemTypeRepoFake struct {
	items map[uuid.UUID]*domainFacility.SPSControllerSystemType
}

func newProjectSPSSystemTypeRepo() *projectSPSSystemTypeRepoFake {
	return &projectSPSSystemTypeRepoFake{items: map[uuid.UUID]*domainFacility.SPSControllerSystemType{}}
}

func (r *projectSPSSystemTypeRepoFake) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainFacility.SPSControllerSystemType, error) {
	return getProjectItemsByID(ids, r.items), nil
}

func (r *projectSPSSystemTypeRepoFake) Create(_ context.Context, entity *domainFacility.SPSControllerSystemType) error {
	if entity.ID == uuid.Nil {
		entity.ID = uuid.New()
	}
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *projectSPSSystemTypeRepoFake) Update(_ context.Context, entity *domainFacility.SPSControllerSystemType) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *projectSPSSystemTypeRepoFake) DeleteByIds(_ context.Context, ids []uuid.UUID) error {
	deleteIDs(ids, r.items)
	return nil
}

func (r *projectSPSSystemTypeRepoFake) GetPaginatedList(context.Context, domain.PaginationParams) (*domain.PaginatedList[domainFacility.SPSControllerSystemType], error) {
	return paginatedFromMap(r.items), nil
}

func (r *projectSPSSystemTypeRepoFake) GetPaginatedListBySPSControllerID(_ context.Context, spsControllerID uuid.UUID, _ domain.PaginationParams) (*domain.PaginatedList[domainFacility.SPSControllerSystemType], error) {
	return paginatedFromFilter(r.items, func(item *domainFacility.SPSControllerSystemType) bool {
		return item.SPSControllerID == spsControllerID
	}), nil
}

func (r *projectSPSSystemTypeRepoFake) GetPaginatedListByProjectID(context.Context, uuid.UUID, domain.PaginationParams) (*domain.PaginatedList[domainFacility.SPSControllerSystemType], error) {
	return paginatedFromMap(r.items), nil
}

func (r *projectSPSSystemTypeRepoFake) ListBySPSControllerID(_ context.Context, spsControllerID uuid.UUID) ([]*domainFacility.SPSControllerSystemType, error) {
	out := make([]*domainFacility.SPSControllerSystemType, 0)
	for _, item := range r.items {
		if item.SPSControllerID == spsControllerID {
			clone := *item
			out = append(out, &clone)
		}
	}
	return out, nil
}

func (r *projectSPSSystemTypeRepoFake) GetIDsBySPSControllerIDs(_ context.Context, ids []uuid.UUID) ([]uuid.UUID, error) {
	set := uuidSet(ids)
	out := make([]uuid.UUID, 0)
	for id, item := range r.items {
		if _, ok := set[item.SPSControllerID]; ok {
			out = append(out, id)
		}
	}
	return out, nil
}

func (r *projectSPSSystemTypeRepoFake) DeleteBySPSControllerIDs(_ context.Context, ids []uuid.UUID) error {
	set := uuidSet(ids)
	for id, item := range r.items {
		if _, ok := set[item.SPSControllerID]; ok {
			delete(r.items, id)
		}
	}
	return nil
}

type projectFieldDeviceStoreFake struct {
	items map[uuid.UUID]*domainFacility.FieldDevice
}

func newProjectFieldDeviceStore() *projectFieldDeviceStoreFake {
	return &projectFieldDeviceStoreFake{items: map[uuid.UUID]*domainFacility.FieldDevice{}}
}

func (r *projectFieldDeviceStoreFake) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainFacility.FieldDevice, error) {
	return getProjectItemsByID(ids, r.items), nil
}

func (r *projectFieldDeviceStoreFake) Create(_ context.Context, entity *domainFacility.FieldDevice) error {
	if entity.ID == uuid.Nil {
		entity.ID = uuid.New()
	}
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *projectFieldDeviceStoreFake) BulkCreate(_ context.Context, entities []*domainFacility.FieldDevice, _ int) error {
	for _, entity := range entities {
		if err := r.Create(context.Background(), entity); err != nil {
			return err
		}
	}
	return nil
}

func (r *projectFieldDeviceStoreFake) Update(_ context.Context, entity *domainFacility.FieldDevice) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *projectFieldDeviceStoreFake) DeleteByIds(_ context.Context, ids []uuid.UUID) error {
	deleteIDs(ids, r.items)
	return nil
}

func (r *projectFieldDeviceStoreFake) GetPaginatedList(context.Context, domain.PaginationParams) (*domain.PaginatedList[domainFacility.FieldDevice], error) {
	return paginatedFromMap(r.items), nil
}

func (r *projectFieldDeviceStoreFake) GetPaginatedListWithFilters(context.Context, domain.PaginationParams, domainFacility.FieldDeviceFilterParams) (*domain.PaginatedList[domainFacility.FieldDevice], error) {
	return paginatedFromMap(r.items), nil
}

func (r *projectFieldDeviceStoreFake) GetIDsBySPSControllerSystemTypeIDs(_ context.Context, ids []uuid.UUID) ([]uuid.UUID, error) {
	set := uuidSet(ids)
	out := make([]uuid.UUID, 0)
	for id, item := range r.items {
		if _, ok := set[item.SPSControllerSystemTypeID]; ok {
			out = append(out, id)
		}
	}
	return out, nil
}

func (r *projectFieldDeviceStoreFake) ExistsApparatNrConflict(context.Context, uuid.UUID, *uuid.UUID, uuid.UUID, int, []uuid.UUID) (bool, error) {
	return false, nil
}

func (r *projectFieldDeviceStoreFake) GetUsedApparatNumbers(context.Context, uuid.UUID, *uuid.UUID, uuid.UUID) ([]int, error) {
	return nil, nil
}

func (r *projectFieldDeviceStoreFake) idsForSystemType(systemTypeID uuid.UUID) []uuid.UUID {
	out := make([]uuid.UUID, 0)
	for id, item := range r.items {
		if item.SPSControllerSystemTypeID == systemTypeID {
			out = append(out, id)
		}
	}
	return out
}

type projectSystemTypeRepoFake struct {
	items map[uuid.UUID]*domainFacility.SystemType
}

func newProjectSystemTypeRepo() *projectSystemTypeRepoFake {
	return &projectSystemTypeRepoFake{items: map[uuid.UUID]*domainFacility.SystemType{}}
}

func (r *projectSystemTypeRepoFake) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainFacility.SystemType, error) {
	return getProjectItemsByID(ids, r.items), nil
}

func (r *projectSystemTypeRepoFake) Create(_ context.Context, entity *domainFacility.SystemType) error {
	if entity.ID == uuid.Nil {
		entity.ID = uuid.New()
	}
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *projectSystemTypeRepoFake) Update(_ context.Context, entity *domainFacility.SystemType) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *projectSystemTypeRepoFake) DeleteByIds(_ context.Context, ids []uuid.UUID) error {
	deleteIDs(ids, r.items)
	return nil
}

func (r *projectSystemTypeRepoFake) GetPaginatedList(context.Context, domain.PaginationParams) (*domain.PaginatedList[domainFacility.SystemType], error) {
	return paginatedFromMap(r.items), nil
}

func (r *projectSystemTypeRepoFake) ExistsName(context.Context, string, *uuid.UUID) (bool, error) {
	return false, nil
}

func (r *projectSystemTypeRepoFake) ExistsOverlappingRange(context.Context, int, int, *uuid.UUID) (bool, error) {
	return false, nil
}

type projectBuildingRepoFake struct {
	items map[uuid.UUID]*domainFacility.Building
}

func newProjectBuildingRepo() *projectBuildingRepoFake {
	return &projectBuildingRepoFake{items: map[uuid.UUID]*domainFacility.Building{}}
}

func (r *projectBuildingRepoFake) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainFacility.Building, error) {
	return getProjectItemsByID(ids, r.items), nil
}

func (r *projectBuildingRepoFake) Create(_ context.Context, entity *domainFacility.Building) error {
	if entity.ID == uuid.Nil {
		entity.ID = uuid.New()
	}
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *projectBuildingRepoFake) Update(_ context.Context, entity *domainFacility.Building) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *projectBuildingRepoFake) DeleteByIds(_ context.Context, ids []uuid.UUID) error {
	deleteIDs(ids, r.items)
	return nil
}

func (r *projectBuildingRepoFake) GetPaginatedList(context.Context, domain.PaginationParams) (*domain.PaginatedList[domainFacility.Building], error) {
	return paginatedFromMap(r.items), nil
}

func (r *projectBuildingRepoFake) ExistsIWSCodeGroup(context.Context, string, int, *uuid.UUID) (bool, error) {
	return false, nil
}

func getProjectItemsByID[T any](ids []uuid.UUID, items map[uuid.UUID]*T) []*T {
	out := make([]*T, 0, len(ids))
	for _, id := range ids {
		if item, ok := items[id]; ok {
			clone := *item
			out = append(out, &clone)
		}
	}
	return out
}

func deleteIDs[T any](ids []uuid.UUID, items map[uuid.UUID]*T) {
	for _, id := range ids {
		delete(items, id)
	}
}

func paginatedFromMap[T any](items map[uuid.UUID]*T) *domain.PaginatedList[T] {
	out := make([]T, 0, len(items))
	for _, item := range items {
		out = append(out, *item)
	}
	return &domain.PaginatedList[T]{Items: out, Total: int64(len(out)), Page: 1, TotalPages: 1}
}

func paginatedFromFilter[T any](items map[uuid.UUID]*T, include func(*T) bool) *domain.PaginatedList[T] {
	out := make([]T, 0)
	for _, item := range items {
		if include(item) {
			out = append(out, *item)
		}
	}
	return &domain.PaginatedList[T]{Items: out, Total: int64(len(out)), Page: 1, TotalPages: 1}
}

func paginatedObjectData(items []*domainFacility.ObjectData) *domain.PaginatedList[domainFacility.ObjectData] {
	out := make([]domainFacility.ObjectData, 0, len(items))
	for _, item := range items {
		out = append(out, *item)
	}
	return &domain.PaginatedList[domainFacility.ObjectData]{Items: out, Total: int64(len(out)), Page: 1, TotalPages: 1}
}

func objectDataForProject(items map[uuid.UUID]*domainFacility.ObjectData, projectID uuid.UUID) []*domainFacility.ObjectData {
	out := make([]*domainFacility.ObjectData, 0)
	for _, item := range items {
		if item.ProjectID != nil && *item.ProjectID == projectID {
			out = append(out, cloneObjectData(item))
		}
	}
	return out
}

func cloneObjectData(item *domainFacility.ObjectData) *domainFacility.ObjectData {
	clone := *item
	if item.BacnetObjects != nil {
		clone.BacnetObjects = make([]*domainFacility.BacnetObject, len(item.BacnetObjects))
		for i, bo := range item.BacnetObjects {
			boClone := *bo
			clone.BacnetObjects[i] = &boClone
		}
	}
	return &clone
}

func uuidSet(ids []uuid.UUID) map[uuid.UUID]struct{} {
	out := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		out[id] = struct{}{}
	}
	return out
}

func sameUUIDSet(left, right []uuid.UUID) bool {
	if len(left) != len(right) {
		return false
	}
	leftStrings := make([]string, len(left))
	rightStrings := make([]string, len(right))
	for i := range left {
		leftStrings[i] = left[i].String()
		rightStrings[i] = right[i].String()
	}
	sort.Strings(leftStrings)
	sort.Strings(rightStrings)
	for i := range leftStrings {
		if leftStrings[i] != rightStrings[i] {
			return false
		}
	}
	return true
}
