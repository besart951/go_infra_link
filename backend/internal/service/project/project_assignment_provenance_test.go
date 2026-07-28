package project

import (
	"context"
	"reflect"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type assignmentSourceCall struct {
	projectID uuid.UUID
	ids       []uuid.UUID
	source    domainProject.AssignmentSource
}

type provenanceSPSLinkRepo struct {
	*projectSPSControllerRepoFake
	calls            []assignmentSourceCall
	pruneCalls       []assignmentSourceCall
	explicitProjects map[uuid.UUID]struct{}
}

func (repo *provenanceSPSLinkRepo) BulkCreateBySPSControllerIDsWithSource(
	_ context.Context,
	projectID uuid.UUID,
	ids []uuid.UUID,
	source domainProject.AssignmentSource,
) error {
	repo.calls = append(repo.calls, assignmentSourceCall{
		projectID: projectID,
		ids:       append([]uuid.UUID(nil), ids...),
		source:    source,
	})
	return nil
}

func (repo *provenanceSPSLinkRepo) RemoveAssignmentSourceAndPrune(
	_ context.Context,
	projectID uuid.UUID,
	source domainProject.AssignmentSource,
) (bool, error) {
	repo.pruneCalls = append(repo.pruneCalls, assignmentSourceCall{
		projectID: projectID,
		source:    source,
	})
	return true, nil
}

func (repo *provenanceSPSLinkRepo) ListProjectIDsByAssignmentSource(
	_ context.Context,
	source domainProject.AssignmentSource,
) ([]uuid.UUID, error) {
	projectIDs := make([]uuid.UUID, 0)
	for _, link := range repo.items {
		if source.Kind == domainProject.AssignmentSourceExplicit &&
			link.SPSControllerID == source.SourceEntityID {
			if repo.explicitProjects != nil {
				if _, ok := repo.explicitProjects[link.ProjectID]; !ok {
					continue
				}
			}
			projectIDs = append(projectIDs, link.ProjectID)
		}
	}
	return projectIDs, nil
}

type provenanceFieldDeviceLinkRepo struct {
	*projectFieldDeviceRepoFake
	systemTypeCalls []assignmentSourceCall
	directCalls     []assignmentSourceCall
	pruneCalls      []assignmentSourceCall
}

func (repo *provenanceFieldDeviceLinkRepo) BulkCreateByFieldDeviceIDsWithSource(
	_ context.Context,
	projectID uuid.UUID,
	ids []uuid.UUID,
	source domainProject.AssignmentSource,
) error {
	repo.directCalls = append(repo.directCalls, assignmentSourceCall{
		projectID: projectID,
		ids:       append([]uuid.UUID(nil), ids...),
		source:    source,
	})
	return nil
}

func (repo *provenanceFieldDeviceLinkRepo) RemoveAssignmentSourceAndPrune(
	_ context.Context,
	projectID uuid.UUID,
	source domainProject.AssignmentSource,
) (bool, error) {
	repo.pruneCalls = append(repo.pruneCalls, assignmentSourceCall{
		projectID: projectID,
		source:    source,
	})
	return true, nil
}

func (repo *provenanceFieldDeviceLinkRepo) BulkCreateBySPSControllerSystemTypeIDsWithSource(
	_ context.Context,
	projectID uuid.UUID,
	ids []uuid.UUID,
	source domainProject.AssignmentSource,
) error {
	repo.systemTypeCalls = append(repo.systemTypeCalls, assignmentSourceCall{
		projectID: projectID,
		ids:       append([]uuid.UUID(nil), ids...),
		source:    source,
	})
	return nil
}

func TestCabinetAssignmentPropagatesOneInheritedSourceToBothDescendantLevels(
	t *testing.T,
) {
	projectID := uuid.New()
	cabinetID := uuid.New()
	controllerID := uuid.New()
	systemTypeID := uuid.New()
	spsLinks := &provenanceSPSLinkRepo{
		projectSPSControllerRepoFake: newProjectSPSControllerRepo(),
	}
	fieldLinks := &provenanceFieldDeviceLinkRepo{
		projectFieldDeviceRepoFake: newProjectFieldDeviceRepo(),
	}
	systemTypes := newProjectSPSSystemTypeRepo()
	systemTypes.items[systemTypeID] = &domainFacility.SPSControllerSystemType{
		SPSControllerID: controllerID,
	}
	store := projectAssignmentStore{
		projectSPSControllerRepo: spsLinks,
		projectFieldDeviceRepo:   fieldLinks,
		spsControllerSystemRepo:  systemTypes,
	}
	source := domainProject.AssignmentSource{
		Kind:           domainProject.AssignmentSourceControlCabinet,
		SourceEntityID: cabinetID,
	}

	if err := store.assignSPSControllerDescendants(
		context.Background(),
		projectID,
		[]uuid.UUID{controllerID},
		source,
		source,
	); err != nil {
		t.Fatalf("assign cabinet descendants: %v", err)
	}

	wantSPS := assignmentSourceCall{
		projectID: projectID,
		ids:       []uuid.UUID{controllerID},
		source:    source,
	}
	if !reflect.DeepEqual(spsLinks.calls, []assignmentSourceCall{wantSPS}) {
		t.Fatalf("SPS source calls: got %+v, want %+v", spsLinks.calls, wantSPS)
	}
	wantField := assignmentSourceCall{
		projectID: projectID,
		ids:       []uuid.UUID{systemTypeID},
		source:    source,
	}
	if !reflect.DeepEqual(
		fieldLinks.systemTypeCalls,
		[]assignmentSourceCall{wantField},
	) {
		t.Fatalf(
			"FieldDevice source calls: got %+v, want %+v",
			fieldLinks.systemTypeCalls,
			wantField,
		)
	}
}

func TestDirectSPSAssignmentUsesExplicitRootAndInheritedFieldDeviceSources(
	t *testing.T,
) {
	projectID := uuid.New()
	controllerID := uuid.New()
	systemTypeID := uuid.New()
	spsLinks := &provenanceSPSLinkRepo{
		projectSPSControllerRepoFake: newProjectSPSControllerRepo(),
	}
	fieldLinks := &provenanceFieldDeviceLinkRepo{
		projectFieldDeviceRepoFake: newProjectFieldDeviceRepo(),
	}
	systemTypes := newProjectSPSSystemTypeRepo()
	systemTypes.items[systemTypeID] = &domainFacility.SPSControllerSystemType{
		SPSControllerID: controllerID,
	}
	store := projectAssignmentStore{
		projectSPSControllerRepo: spsLinks,
		projectFieldDeviceRepo:   fieldLinks,
		spsControllerSystemRepo:  systemTypes,
	}
	fieldSource := domainProject.AssignmentSource{
		Kind:           domainProject.AssignmentSourceSPSController,
		SourceEntityID: controllerID,
	}

	if err := store.assignSPSControllerDescendants(
		context.Background(),
		projectID,
		[]uuid.UUID{controllerID},
		domainProject.ExplicitAssignmentSource(),
		fieldSource,
	); err != nil {
		t.Fatalf("assign SPS descendants: %v", err)
	}

	if got := spsLinks.calls[0].source; got.Kind != domainProject.AssignmentSourceExplicit {
		t.Fatalf("root source: got %+v", got)
	}
	if got := fieldLinks.systemTypeCalls[0].source; got != fieldSource {
		t.Fatalf("FieldDevice source: got %+v, want %+v", got, fieldSource)
	}
}

func TestFieldDeviceMoveTransfersLiveParentClaimsAndKeepsAffectedProjects(
	t *testing.T,
) {
	oldSPSID := uuid.New()
	newSPSID := uuid.New()
	oldCabinetID := uuid.New()
	newCabinetID := uuid.New()
	oldSystemTypeID := uuid.New()
	newSystemTypeID := uuid.New()
	fieldDeviceID := uuid.New()
	oldSPSProjectID := uuid.New()
	newSPSProjectID := uuid.New()
	inheritedNewSPSProjectID := uuid.New()
	oldCabinetProjectID := uuid.New()
	newCabinetProjectID := uuid.New()

	cabinetLinks := newProjectControlCabinetRepo()
	_ = cabinetLinks.Create(context.Background(), &domainProject.ProjectControlCabinet{
		ProjectID:        oldCabinetProjectID,
		ControlCabinetID: oldCabinetID,
	})
	_ = cabinetLinks.Create(context.Background(), &domainProject.ProjectControlCabinet{
		ProjectID:        newCabinetProjectID,
		ControlCabinetID: newCabinetID,
	})
	spsLinks := &provenanceSPSLinkRepo{
		projectSPSControllerRepoFake: newProjectSPSControllerRepo(),
		explicitProjects: map[uuid.UUID]struct{}{
			oldSPSProjectID: {},
			newSPSProjectID: {},
		},
	}
	spsLinks.createWithID(oldSPSProjectID, oldSPSID)
	spsLinks.createWithID(newSPSProjectID, newSPSID)
	spsLinks.createWithID(inheritedNewSPSProjectID, newSPSID)
	fieldLinks := &provenanceFieldDeviceLinkRepo{
		projectFieldDeviceRepoFake: newProjectFieldDeviceRepo(),
	}
	systemTypes := newProjectSPSSystemTypeRepo()
	systemTypes.items[oldSystemTypeID] = &domainFacility.SPSControllerSystemType{
		SPSControllerID: oldSPSID,
	}
	systemTypes.items[newSystemTypeID] = &domainFacility.SPSControllerSystemType{
		SPSControllerID: newSPSID,
	}
	controllers := newProjectSPSRepo()
	controllers.items[oldSPSID] = &domainFacility.SPSController{
		Base:             domain.Base{ID: oldSPSID},
		ControlCabinetID: oldCabinetID,
	}
	controllers.items[newSPSID] = &domainFacility.SPSController{
		Base:             domain.Base{ID: newSPSID},
		ControlCabinetID: newCabinetID,
	}
	service := &ProjectFacilityLinkService{
		projectControlCabinetRepo: cabinetLinks,
		projectSPSControllerRepo:  spsLinks,
		projectFieldDeviceRepo:    fieldLinks,
		spsControllerRepo:         controllers,
		spsControllerSystemRepo:   systemTypes,
	}

	affected, err := service.ReconcileFieldDeviceMove(
		context.Background(),
		fieldDeviceID,
		oldSystemTypeID,
		newSystemTypeID,
	)
	if err != nil {
		t.Fatalf("reconcile FieldDevice move: %v", err)
	}
	wantAffected := []uuid.UUID{
		oldSPSProjectID,
		newSPSProjectID,
		inheritedNewSPSProjectID,
		oldCabinetProjectID,
		newCabinetProjectID,
	}
	assertSameUUIDs(t, affected, wantAffected)

	assertAssignmentSourceCall(
		t,
		fieldLinks.directCalls,
		newSPSProjectID,
		fieldDeviceID,
		domainProject.AssignmentSourceSPSController,
		newSPSID,
	)
	assertNoAssignmentSourceCall(
		t,
		fieldLinks.directCalls,
		inheritedNewSPSProjectID,
		domainProject.AssignmentSourceSPSController,
		newSPSID,
	)
	assertAssignmentSourceCall(
		t,
		fieldLinks.directCalls,
		newCabinetProjectID,
		fieldDeviceID,
		domainProject.AssignmentSourceControlCabinet,
		newCabinetID,
	)
	assertPruneSourceCall(
		t,
		fieldLinks.pruneCalls,
		oldSPSProjectID,
		domainProject.AssignmentSourceSPSController,
		oldSPSID,
	)
	assertPruneSourceCall(
		t,
		fieldLinks.pruneCalls,
		oldCabinetProjectID,
		domainProject.AssignmentSourceControlCabinet,
		oldCabinetID,
	)
}

func TestSPSControllerMoveTransfersCabinetClaimsForRootAndDescendants(
	t *testing.T,
) {
	controllerID := uuid.New()
	systemTypeID := uuid.New()
	oldCabinetID := uuid.New()
	newCabinetID := uuid.New()
	oldProjectID := uuid.New()
	newProjectID := uuid.New()

	cabinetLinks := newProjectControlCabinetRepo()
	_ = cabinetLinks.Create(context.Background(), &domainProject.ProjectControlCabinet{
		ProjectID:        oldProjectID,
		ControlCabinetID: oldCabinetID,
	})
	_ = cabinetLinks.Create(context.Background(), &domainProject.ProjectControlCabinet{
		ProjectID:        newProjectID,
		ControlCabinetID: newCabinetID,
	})
	spsLinks := &provenanceSPSLinkRepo{
		projectSPSControllerRepoFake: newProjectSPSControllerRepo(),
	}
	fieldLinks := &provenanceFieldDeviceLinkRepo{
		projectFieldDeviceRepoFake: newProjectFieldDeviceRepo(),
	}
	systemTypes := newProjectSPSSystemTypeRepo()
	systemTypes.items[systemTypeID] = &domainFacility.SPSControllerSystemType{
		SPSControllerID: controllerID,
	}
	service := &ProjectFacilityLinkService{
		projectControlCabinetRepo: cabinetLinks,
		projectSPSControllerRepo:  spsLinks,
		projectFieldDeviceRepo:    fieldLinks,
		spsControllerSystemRepo:   systemTypes,
	}

	affected, err := service.ReconcileSPSControllerMove(
		context.Background(),
		controllerID,
		oldCabinetID,
		newCabinetID,
	)
	if err != nil {
		t.Fatalf("reconcile SPSController move: %v", err)
	}
	assertSameUUIDs(t, affected, []uuid.UUID{oldProjectID, newProjectID})
	assertAssignmentSourceCall(
		t,
		spsLinks.calls,
		newProjectID,
		controllerID,
		domainProject.AssignmentSourceControlCabinet,
		newCabinetID,
	)
	assertAssignmentSourceCall(
		t,
		fieldLinks.systemTypeCalls,
		newProjectID,
		systemTypeID,
		domainProject.AssignmentSourceControlCabinet,
		newCabinetID,
	)
	assertPruneSourceCall(
		t,
		spsLinks.pruneCalls,
		oldProjectID,
		domainProject.AssignmentSourceControlCabinet,
		oldCabinetID,
	)
	assertPruneSourceCall(
		t,
		fieldLinks.pruneCalls,
		oldProjectID,
		domainProject.AssignmentSourceControlCabinet,
		oldCabinetID,
	)
}

func assertAssignmentSourceCall(
	t *testing.T,
	calls []assignmentSourceCall,
	projectID uuid.UUID,
	entityID uuid.UUID,
	kind domainProject.AssignmentSourceKind,
	sourceEntityID uuid.UUID,
) {
	t.Helper()
	for _, call := range calls {
		if call.projectID == projectID &&
			reflect.DeepEqual(call.ids, []uuid.UUID{entityID}) &&
			call.source.Kind == kind &&
			call.source.SourceEntityID == sourceEntityID {
			return
		}
	}
	t.Fatalf(
		"missing assignment source call project=%s entity=%s kind=%s source=%s in %+v",
		projectID,
		entityID,
		kind,
		sourceEntityID,
		calls,
	)
}

func assertPruneSourceCall(
	t *testing.T,
	calls []assignmentSourceCall,
	projectID uuid.UUID,
	kind domainProject.AssignmentSourceKind,
	sourceEntityID uuid.UUID,
) {
	t.Helper()
	for _, call := range calls {
		if call.projectID == projectID &&
			call.source.Kind == kind &&
			call.source.SourceEntityID == sourceEntityID {
			return
		}
	}
	t.Fatalf(
		"missing prune call project=%s kind=%s source=%s in %+v",
		projectID,
		kind,
		sourceEntityID,
		calls,
	)
}

func assertNoAssignmentSourceCall(
	t *testing.T,
	calls []assignmentSourceCall,
	projectID uuid.UUID,
	kind domainProject.AssignmentSourceKind,
	sourceEntityID uuid.UUID,
) {
	t.Helper()
	for _, call := range calls {
		if call.projectID == projectID &&
			call.source.Kind == kind &&
			call.source.SourceEntityID == sourceEntityID {
			t.Fatalf(
				"unexpected assignment source call project=%s kind=%s source=%s in %+v",
				projectID,
				kind,
				sourceEntityID,
				calls,
			)
		}
	}
}

func assertSameUUIDs(t *testing.T, got []uuid.UUID, want []uuid.UUID) {
	t.Helper()
	gotSet := make(map[uuid.UUID]int, len(got))
	for _, id := range got {
		gotSet[id]++
	}
	for _, id := range want {
		gotSet[id]--
	}
	for id, count := range gotSet {
		if count != 0 {
			t.Fatalf("UUID set differs at %s: got %v, want %v", id, got, want)
		}
	}
	if len(got) != len(want) {
		t.Fatalf("UUID count: got %v, want %v", got, want)
	}
}
