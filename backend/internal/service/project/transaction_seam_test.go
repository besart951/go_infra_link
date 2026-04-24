package project

import (
	"context"
	"errors"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestProjectTransaction_CopyControlCabinetFailureDoesNotUseRollbackDeletes(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	buildingID := uuid.New()
	originalControlCabinetID := uuid.New()
	originalSPSID := uuid.New()
	systemTypeDefinitionID := uuid.New()
	originalSystemTypeID := uuid.New()
	originalFieldDeviceID := uuid.New()
	linkErr := errors.New("link control cabinet failed")
	cabinetNr := "AK1"
	gaDevice := "A01"
	originalNumber := 1

	baseControlCabinetRepo := newProjectControlCabinetStore()
	baseBuildingRepo := newProjectBuildingRepo()
	baseSPSRepo := newProjectSPSRepo()
	baseSystemTypeRepo := newProjectSystemTypeRepo()
	baseSPSSystemRepo := newProjectSPSSystemTypeRepo()
	baseFieldDeviceRepo := newProjectFieldDeviceStore()
	baseSpecRepo := newProjectSpecificationRepo()
	baseBacnetRepo := newProjectBacnetObjectRepo()

	txControlCabinetRepo := newProjectControlCabinetStore()
	txBuildingRepo := newProjectBuildingRepo()
	txSPSRepo := newProjectSPSRepo()
	txSystemTypeRepo := newProjectSystemTypeRepo()
	txSPSSystemRepo := newProjectSPSSystemTypeRepo()
	txFieldDeviceRepo := newProjectFieldDeviceStore()
	txSpecRepo := newProjectSpecificationRepo()
	txBacnetRepo := newProjectBacnetObjectRepo()

	seedProjectControlCabinetCopyHierarchy(
		buildingID,
		originalControlCabinetID,
		originalSPSID,
		systemTypeDefinitionID,
		originalSystemTypeID,
		originalFieldDeviceID,
		cabinetNr,
		gaDevice,
		originalNumber,
		baseControlCabinetRepo,
		baseBuildingRepo,
		baseSPSRepo,
		baseSystemTypeRepo,
		baseSPSSystemRepo,
		baseFieldDeviceRepo,
	)
	seedProjectControlCabinetCopyHierarchy(
		buildingID,
		originalControlCabinetID,
		originalSPSID,
		systemTypeDefinitionID,
		originalSystemTypeID,
		originalFieldDeviceID,
		cabinetNr,
		gaDevice,
		originalNumber,
		txControlCabinetRepo,
		txBuildingRepo,
		txSPSRepo,
		txSystemTypeRepo,
		txSPSSystemRepo,
		txFieldDeviceRepo,
	)

	baseDeps := Dependencies{
		Projects:                 newProjectRepo(),
		ProjectControlCabinets:   newProjectControlCabinetRepo(),
		ProjectSPSControllers:    newProjectSPSControllerRepo(),
		ProjectFieldDevices:      newProjectFieldDeviceRepo(),
		ObjectData:               newProjectObjectDataRepo(),
		BacnetObjects:            baseBacnetRepo,
		Specifications:           baseSpecRepo,
		ControlCabinets:          baseControlCabinetRepo,
		SPSControllers:           baseSPSRepo,
		SPSControllerSystemTypes: baseSPSSystemRepo,
		FieldDevices:             baseFieldDeviceRepo,
		HierarchyCopier: facilityservice.NewHierarchyCopier(
			baseControlCabinetRepo,
			baseBuildingRepo,
			baseSPSRepo,
			baseSystemTypeRepo,
			baseSPSSystemRepo,
			baseFieldDeviceRepo,
			baseSpecRepo,
			baseBacnetRepo,
		),
	}
	txDeps := Dependencies{
		Projects:                 newProjectRepo(),
		ProjectControlCabinets:   &failingProjectControlCabinetLinkRepo{projectControlCabinetRepoFake: newProjectControlCabinetRepo(), createErr: linkErr},
		ProjectSPSControllers:    newProjectSPSControllerRepo(),
		ProjectFieldDevices:      newProjectFieldDeviceRepo(),
		ObjectData:               newProjectObjectDataRepo(),
		BacnetObjects:            txBacnetRepo,
		Specifications:           txSpecRepo,
		ControlCabinets:          txControlCabinetRepo,
		SPSControllers:           txSPSRepo,
		SPSControllerSystemTypes: txSPSSystemRepo,
		FieldDevices:             txFieldDeviceRepo,
		HierarchyCopier: facilityservice.NewHierarchyCopier(
			txControlCabinetRepo,
			txBuildingRepo,
			txSPSRepo,
			txSystemTypeRepo,
			txSPSSystemRepo,
			txFieldDeviceRepo,
			txSpecRepo,
			txBacnetRepo,
		),
	}

	runnerCalls := 0
	services := newProjectTxServices(baseDeps, txDeps, &runnerCalls)

	_, err := services.FacilityLink.CopyControlCabinet(ctx, projectID, originalControlCabinetID)
	if !errors.Is(err, linkErr) {
		t.Fatalf("expected link error, got %v", err)
	}
	if runnerCalls != 1 {
		t.Fatalf("expected one transaction run, got %d", runnerCalls)
	}
	if len(baseControlCabinetRepo.items) != 1 || len(baseSPSRepo.items) != 1 || len(baseSPSSystemRepo.items) != 1 || len(baseFieldDeviceRepo.items) != 1 {
		t.Fatalf("expected base repositories to remain unchanged, got cabinets=%d sps=%d systemTypes=%d fieldDevices=%d", len(baseControlCabinetRepo.items), len(baseSPSRepo.items), len(baseSPSSystemRepo.items), len(baseFieldDeviceRepo.items))
	}
	if len(txControlCabinetRepo.items) != 2 || len(txSPSRepo.items) != 2 || len(txSPSSystemRepo.items) != 2 || len(txFieldDeviceRepo.items) != 2 {
		t.Fatalf("expected copied hierarchy to remain in tx repositories without manual rollback, got cabinets=%d sps=%d systemTypes=%d fieldDevices=%d", len(txControlCabinetRepo.items), len(txSPSRepo.items), len(txSPSSystemRepo.items), len(txFieldDeviceRepo.items))
	}
}

func TestProjectTransaction_CopySPSControllerFailureDoesNotUseRollbackDeletes(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	buildingID := uuid.New()
	controlCabinetID := uuid.New()
	originalSPSID := uuid.New()
	systemTypeDefinitionID := uuid.New()
	originalSystemTypeID := uuid.New()
	originalFieldDeviceID := uuid.New()
	linkErr := errors.New("link sps controller failed")
	cabinetNr := "AK1"
	gaDevice := "A01"
	originalNumber := 1

	baseControlCabinetRepo := newProjectControlCabinetStore()
	baseBuildingRepo := newProjectBuildingRepo()
	baseSPSRepo := newProjectSPSRepo()
	baseSystemTypeRepo := newProjectSystemTypeRepo()
	baseSPSSystemRepo := newProjectSPSSystemTypeRepo()
	baseFieldDeviceRepo := newProjectFieldDeviceStore()
	baseSpecRepo := newProjectSpecificationRepo()
	baseBacnetRepo := newProjectBacnetObjectRepo()

	txControlCabinetRepo := newProjectControlCabinetStore()
	txBuildingRepo := newProjectBuildingRepo()
	txSPSRepo := newProjectSPSRepo()
	txSystemTypeRepo := newProjectSystemTypeRepo()
	txSPSSystemRepo := newProjectSPSSystemTypeRepo()
	txFieldDeviceRepo := newProjectFieldDeviceStore()
	txSpecRepo := newProjectSpecificationRepo()
	txBacnetRepo := newProjectBacnetObjectRepo()

	seedProjectControlCabinetCopyHierarchy(
		buildingID,
		controlCabinetID,
		originalSPSID,
		systemTypeDefinitionID,
		originalSystemTypeID,
		originalFieldDeviceID,
		cabinetNr,
		gaDevice,
		originalNumber,
		baseControlCabinetRepo,
		baseBuildingRepo,
		baseSPSRepo,
		baseSystemTypeRepo,
		baseSPSSystemRepo,
		baseFieldDeviceRepo,
	)
	seedProjectControlCabinetCopyHierarchy(
		buildingID,
		controlCabinetID,
		originalSPSID,
		systemTypeDefinitionID,
		originalSystemTypeID,
		originalFieldDeviceID,
		cabinetNr,
		gaDevice,
		originalNumber,
		txControlCabinetRepo,
		txBuildingRepo,
		txSPSRepo,
		txSystemTypeRepo,
		txSPSSystemRepo,
		txFieldDeviceRepo,
	)

	baseDeps := Dependencies{
		Projects:                 newProjectRepo(),
		ProjectControlCabinets:   newProjectControlCabinetRepo(),
		ProjectSPSControllers:    newProjectSPSControllerRepo(),
		ProjectFieldDevices:      newProjectFieldDeviceRepo(),
		ObjectData:               newProjectObjectDataRepo(),
		BacnetObjects:            baseBacnetRepo,
		Specifications:           baseSpecRepo,
		ControlCabinets:          baseControlCabinetRepo,
		SPSControllers:           baseSPSRepo,
		SPSControllerSystemTypes: baseSPSSystemRepo,
		FieldDevices:             baseFieldDeviceRepo,
		HierarchyCopier: facilityservice.NewHierarchyCopier(
			baseControlCabinetRepo,
			baseBuildingRepo,
			baseSPSRepo,
			baseSystemTypeRepo,
			baseSPSSystemRepo,
			baseFieldDeviceRepo,
			baseSpecRepo,
			baseBacnetRepo,
		),
	}
	txDeps := Dependencies{
		Projects:                 newProjectRepo(),
		ProjectControlCabinets:   newProjectControlCabinetRepo(),
		ProjectSPSControllers:    &failingProjectSPSControllerLinkRepo{projectSPSControllerRepoFake: newProjectSPSControllerRepo(), createErr: linkErr},
		ProjectFieldDevices:      newProjectFieldDeviceRepo(),
		ObjectData:               newProjectObjectDataRepo(),
		BacnetObjects:            txBacnetRepo,
		Specifications:           txSpecRepo,
		ControlCabinets:          txControlCabinetRepo,
		SPSControllers:           txSPSRepo,
		SPSControllerSystemTypes: txSPSSystemRepo,
		FieldDevices:             txFieldDeviceRepo,
		HierarchyCopier: facilityservice.NewHierarchyCopier(
			txControlCabinetRepo,
			txBuildingRepo,
			txSPSRepo,
			txSystemTypeRepo,
			txSPSSystemRepo,
			txFieldDeviceRepo,
			txSpecRepo,
			txBacnetRepo,
		),
	}

	runnerCalls := 0
	services := newProjectTxServices(baseDeps, txDeps, &runnerCalls)

	_, err := services.FacilityLink.CopySPSController(ctx, projectID, originalSPSID)
	if !errors.Is(err, linkErr) {
		t.Fatalf("expected link error, got %v", err)
	}
	if runnerCalls != 1 {
		t.Fatalf("expected one transaction run, got %d", runnerCalls)
	}
	if len(baseSPSRepo.items) != 1 || len(baseSPSSystemRepo.items) != 1 || len(baseFieldDeviceRepo.items) != 1 {
		t.Fatalf("expected base repositories to remain unchanged, got sps=%d systemTypes=%d fieldDevices=%d", len(baseSPSRepo.items), len(baseSPSSystemRepo.items), len(baseFieldDeviceRepo.items))
	}
	if len(txSPSRepo.items) != 2 || len(txSPSSystemRepo.items) != 2 || len(txFieldDeviceRepo.items) != 2 {
		t.Fatalf("expected copied hierarchy to remain in tx repositories without manual rollback, got sps=%d systemTypes=%d fieldDevices=%d", len(txSPSRepo.items), len(txSPSSystemRepo.items), len(txFieldDeviceRepo.items))
	}
}

func TestProjectTransaction_CopySPSControllerSystemTypeFailureDoesNotUseRollbackDeletes(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	originalSPSID := uuid.New()
	systemTypeDefinitionID := uuid.New()
	originalSystemTypeID := uuid.New()
	originalFieldDeviceID := uuid.New()
	linkErr := errors.New("link field device failed")
	originalNumber := 1

	baseSPSSystemRepo := newProjectSPSSystemTypeRepo()
	baseFieldDeviceRepo := newProjectFieldDeviceStore()
	baseSystemTypeRepo := newProjectSystemTypeRepo()
	baseSpecRepo := newProjectSpecificationRepo()
	baseBacnetRepo := newProjectBacnetObjectRepo()

	txSPSSystemRepo := newProjectSPSSystemTypeRepo()
	txFieldDeviceRepo := newProjectFieldDeviceStore()
	txSystemTypeRepo := newProjectSystemTypeRepo()
	txSpecRepo := newProjectSpecificationRepo()
	txBacnetRepo := newProjectBacnetObjectRepo()

	seedProjectSystemTypeCopyHierarchy(
		originalSPSID,
		systemTypeDefinitionID,
		originalSystemTypeID,
		originalFieldDeviceID,
		originalNumber,
		baseSystemTypeRepo,
		baseSPSSystemRepo,
		baseFieldDeviceRepo,
	)
	seedProjectSystemTypeCopyHierarchy(
		originalSPSID,
		systemTypeDefinitionID,
		originalSystemTypeID,
		originalFieldDeviceID,
		originalNumber,
		txSystemTypeRepo,
		txSPSSystemRepo,
		txFieldDeviceRepo,
	)

	baseDeps := Dependencies{
		Projects:                 newProjectRepo(),
		ProjectControlCabinets:   newProjectControlCabinetRepo(),
		ProjectSPSControllers:    newProjectSPSControllerRepo(),
		ProjectFieldDevices:      newProjectFieldDeviceRepo(),
		ObjectData:               newProjectObjectDataRepo(),
		BacnetObjects:            baseBacnetRepo,
		Specifications:           baseSpecRepo,
		ControlCabinets:          newProjectControlCabinetStore(),
		SPSControllers:           newProjectSPSRepo(),
		SPSControllerSystemTypes: baseSPSSystemRepo,
		FieldDevices:             baseFieldDeviceRepo,
		HierarchyCopier: facilityservice.NewHierarchyCopier(
			newProjectControlCabinetStore(),
			newProjectBuildingRepo(),
			newProjectSPSRepo(),
			baseSystemTypeRepo,
			baseSPSSystemRepo,
			baseFieldDeviceRepo,
			baseSpecRepo,
			baseBacnetRepo,
		),
	}
	txDeps := Dependencies{
		Projects:                 newProjectRepo(),
		ProjectControlCabinets:   newProjectControlCabinetRepo(),
		ProjectSPSControllers:    newProjectSPSControllerRepo(),
		ProjectFieldDevices:      &failingProjectFieldDeviceLinkRepo{projectFieldDeviceRepoFake: newProjectFieldDeviceRepo(), createErr: linkErr},
		ObjectData:               newProjectObjectDataRepo(),
		BacnetObjects:            txBacnetRepo,
		Specifications:           txSpecRepo,
		ControlCabinets:          newProjectControlCabinetStore(),
		SPSControllers:           newProjectSPSRepo(),
		SPSControllerSystemTypes: txSPSSystemRepo,
		FieldDevices:             txFieldDeviceRepo,
		HierarchyCopier: facilityservice.NewHierarchyCopier(
			newProjectControlCabinetStore(),
			newProjectBuildingRepo(),
			newProjectSPSRepo(),
			txSystemTypeRepo,
			txSPSSystemRepo,
			txFieldDeviceRepo,
			txSpecRepo,
			txBacnetRepo,
		),
	}

	runnerCalls := 0
	services := newProjectTxServices(baseDeps, txDeps, &runnerCalls)

	_, err := services.FacilityLink.CopySPSControllerSystemType(ctx, projectID, originalSystemTypeID)
	if !errors.Is(err, linkErr) {
		t.Fatalf("expected link error, got %v", err)
	}
	if runnerCalls != 1 {
		t.Fatalf("expected one transaction run, got %d", runnerCalls)
	}
	if len(baseSPSSystemRepo.items) != 1 || len(baseFieldDeviceRepo.items) != 1 {
		t.Fatalf("expected base repositories to remain unchanged, got systemTypes=%d fieldDevices=%d", len(baseSPSSystemRepo.items), len(baseFieldDeviceRepo.items))
	}
	if len(txSPSSystemRepo.items) != 2 || len(txFieldDeviceRepo.items) != 2 {
		t.Fatalf("expected copied hierarchy to remain in tx repositories without manual rollback, got systemTypes=%d fieldDevices=%d", len(txSPSSystemRepo.items), len(txFieldDeviceRepo.items))
	}
}

type failingProjectControlCabinetLinkRepo struct {
	*projectControlCabinetRepoFake
	createErr error
}

func (r *failingProjectControlCabinetLinkRepo) Create(_ context.Context, entity *domainProject.ProjectControlCabinet) error {
	if r.createErr != nil {
		return r.createErr
	}
	return r.projectControlCabinetRepoFake.Create(context.Background(), entity)
}

type failingProjectSPSControllerLinkRepo struct {
	*projectSPSControllerRepoFake
	createErr error
}

func (r *failingProjectSPSControllerLinkRepo) Create(_ context.Context, entity *domainProject.ProjectSPSController) error {
	if r.createErr != nil {
		return r.createErr
	}
	return r.projectSPSControllerRepoFake.Create(context.Background(), entity)
}

type failingProjectFieldDeviceLinkRepo struct {
	*projectFieldDeviceRepoFake
	createErr error
}

func (r *failingProjectFieldDeviceLinkRepo) Create(_ context.Context, entity *domainProject.ProjectFieldDevice) error {
	if r.createErr != nil {
		return r.createErr
	}
	return r.projectFieldDeviceRepoFake.Create(context.Background(), entity)
}

func newProjectTxServices(baseDeps Dependencies, txDeps Dependencies, runnerCalls *int) *Services {
	return NewServices(baseDeps, Config{
		TxRunner: func(run func(tx *gorm.DB) error) error {
			*runnerCalls = *runnerCalls + 1
			return run(&gorm.DB{})
		},
		TxFactory: func(_ *gorm.DB) (*Services, error) {
			return NewServices(txDeps), nil
		},
	})
}

func seedProjectControlCabinetCopyHierarchy(
	buildingID uuid.UUID,
	controlCabinetID uuid.UUID,
	spsControllerID uuid.UUID,
	systemTypeDefinitionID uuid.UUID,
	systemTypeID uuid.UUID,
	fieldDeviceID uuid.UUID,
	cabinetNr string,
	gaDevice string,
	number int,
	controlCabinetRepo *projectControlCabinetStoreFake,
	buildingRepo *projectBuildingRepoFake,
	spsRepo *projectSPSRepoFake,
	systemTypeRepo *projectSystemTypeRepoFake,
	spsSystemRepo *projectSPSSystemTypeRepoFake,
	fieldDeviceRepo *projectFieldDeviceStoreFake,
) {
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
	spsRepo.items[spsControllerID] = &domainFacility.SPSController{
		Base:             domain.Base{ID: spsControllerID},
		ControlCabinetID: controlCabinetID,
		GADevice:         &gaDevice,
		DeviceName:       "B01_AK1_A01",
	}
	systemTypeRepo.items[systemTypeDefinitionID] = &domainFacility.SystemType{
		Base:      domain.Base{ID: systemTypeDefinitionID},
		NumberMin: 1,
		NumberMax: 99,
	}
	spsSystemRepo.items[systemTypeID] = &domainFacility.SPSControllerSystemType{
		Base:            domain.Base{ID: systemTypeID},
		Number:          &number,
		SPSControllerID: spsControllerID,
		SystemTypeID:    systemTypeDefinitionID,
	}
	fieldDeviceRepo.items[fieldDeviceID] = &domainFacility.FieldDevice{
		Base:                      domain.Base{ID: fieldDeviceID},
		SPSControllerSystemTypeID: systemTypeID,
		ApparatNr:                 3,
	}
}

func seedProjectSystemTypeCopyHierarchy(
	spsControllerID uuid.UUID,
	systemTypeDefinitionID uuid.UUID,
	systemTypeID uuid.UUID,
	fieldDeviceID uuid.UUID,
	number int,
	systemTypeRepo *projectSystemTypeRepoFake,
	spsSystemRepo *projectSPSSystemTypeRepoFake,
	fieldDeviceRepo *projectFieldDeviceStoreFake,
) {
	systemTypeRepo.items[systemTypeDefinitionID] = &domainFacility.SystemType{
		Base:      domain.Base{ID: systemTypeDefinitionID},
		NumberMin: 1,
		NumberMax: 99,
	}
	spsSystemRepo.items[systemTypeID] = &domainFacility.SPSControllerSystemType{
		Base:            domain.Base{ID: systemTypeID},
		Number:          &number,
		SPSControllerID: spsControllerID,
		SystemTypeID:    systemTypeDefinitionID,
	}
	fieldDeviceRepo.items[fieldDeviceID] = &domainFacility.FieldDevice{
		Base:                      domain.Base{ID: fieldDeviceID},
		SPSControllerSystemTypeID: systemTypeID,
		ApparatNr:                 7,
	}
}
