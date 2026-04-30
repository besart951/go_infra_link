package project

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/google/uuid"
)

type ProjectFacilityLinkService struct {
	projectRepo               domainProject.ProjectRepository
	projectControlCabinetRepo domainProject.ProjectControlCabinetRepository
	projectSPSControllerRepo  domainProject.ProjectSPSControllerRepository
	projectFieldDeviceRepo    domainProject.ProjectFieldDeviceRepository
	objectDataRepo            domainFacility.ObjectDataStore
	bacnetObjectRepo          domainFacility.BacnetObjectStore
	specificationRepo         domainFacility.SpecificationStore
	controlCabinetRepo        domainFacility.ControlCabinetRepository
	spsControllerRepo         domainFacility.SPSControllerRepository
	spsControllerSystemRepo   domainFacility.SPSControllerSystemTypeStore
	fieldDeviceRepo           domainFacility.FieldDeviceStore
	hierarchyCopier           *facilityservice.HierarchyCopier
	fieldDeviceCreator        fieldDeviceCreator
	tx                        txCoordinator
}

type fieldDeviceCreator interface {
	MultiCreate(ctx context.Context, items []domainFacility.FieldDeviceCreateItem) *domainFacility.FieldDeviceMultiCreateResult
}

func (s *ProjectFacilityLinkService) bindTransactions(tx txCoordinator) {
	s.tx = tx
}

func (s *ProjectFacilityLinkService) transaction() projectTx[*ProjectFacilityLinkService] {
	return newProjectTx(s.tx, s, func(services *Services) *ProjectFacilityLinkService {
		return services.FacilityLink
	})
}

func (s *ProjectFacilityLinkService) withTx(fn func(*ProjectFacilityLinkService) error) error {
	return s.transaction().run(fn)
}

func withProjectFacilityLinkTxResult[T any](s *ProjectFacilityLinkService, fn func(*ProjectFacilityLinkService) (T, error)) (T, error) {
	return runProjectTxResult(s.transaction(), fn)
}

func (s *ProjectFacilityLinkService) ListProjectIDsByControlCabinetID(ctx context.Context, controlCabinetID uuid.UUID) ([]uuid.UUID, error) {
	items, err := s.projectControlCabinetRepo.GetByControlCabinetID(ctx, controlCabinetID)
	if err != nil {
		return nil, err
	}

	projectIDSet := make(map[uuid.UUID]struct{}, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		projectIDSet[item.ProjectID] = struct{}{}
	}

	projectIDs := make([]uuid.UUID, 0, len(projectIDSet))
	for projectID := range projectIDSet {
		projectIDs = append(projectIDs, projectID)
	}
	return projectIDs, nil
}

func (s *ProjectFacilityLinkService) ListProjectIDsBySPSControllerID(ctx context.Context, spsControllerID uuid.UUID) ([]uuid.UUID, error) {
	items, err := s.projectSPSControllerRepo.GetBySPSControllerID(ctx, spsControllerID)
	if err != nil {
		return nil, err
	}

	projectIDSet := make(map[uuid.UUID]struct{}, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		projectIDSet[item.ProjectID] = struct{}{}
	}

	projectIDs := make([]uuid.UUID, 0, len(projectIDSet))
	for projectID := range projectIDSet {
		projectIDs = append(projectIDs, projectID)
	}
	return projectIDs, nil
}

func (s *ProjectFacilityLinkService) CreateControlCabinet(ctx context.Context, projectID, controlCabinetID uuid.UUID) (*domainProject.ProjectControlCabinet, error) {
	return withProjectFacilityLinkTxResult(s, func(txService *ProjectFacilityLinkService) (*domainProject.ProjectControlCabinet, error) {
		return txService.assignments().assignControlCabinet(ctx, projectID, controlCabinetID)
	})
}

func (s *ProjectFacilityLinkService) CopyControlCabinet(ctx context.Context, projectID, controlCabinetID uuid.UUID) (*domainFacility.ControlCabinet, error) {
	return withProjectFacilityLinkTxResult(s, func(txService *ProjectFacilityLinkService) (*domainFacility.ControlCabinet, error) {
		return txService.copyControlCabinet(ctx, projectID, controlCabinetID)
	})
}

func (s *ProjectFacilityLinkService) copyControlCabinet(ctx context.Context, projectID, controlCabinetID uuid.UUID) (*domainFacility.ControlCabinet, error) {
	copyEntity, err := s.hierarchyCopier.CopyControlCabinetByID(ctx, controlCabinetID)
	if err != nil {
		return nil, err
	}

	if _, err := s.assignments().assignControlCabinet(ctx, projectID, copyEntity.ID); err != nil {
		return nil, err
	}

	return copyEntity, nil
}

func (s *ProjectFacilityLinkService) UpdateControlCabinet(ctx context.Context, linkID, projectID, controlCabinetID uuid.UUID) (*domainProject.ProjectControlCabinet, error) {
	return withProjectFacilityLinkTxResult(s, func(txService *ProjectFacilityLinkService) (*domainProject.ProjectControlCabinet, error) {
		return txService.assignments().updateControlCabinet(ctx, linkID, projectID, controlCabinetID)
	})
}

func (s *ProjectFacilityLinkService) DeleteControlCabinet(ctx context.Context, linkID, projectID uuid.UUID) error {
	return s.withTx(func(txService *ProjectFacilityLinkService) error {
		return txService.assignments().removeControlCabinet(ctx, linkID, projectID)
	})
}

func (s *ProjectFacilityLinkService) CreateSPSController(ctx context.Context, projectID, spsControllerID uuid.UUID) (*domainProject.ProjectSPSController, error) {
	return withProjectFacilityLinkTxResult(s, func(txService *ProjectFacilityLinkService) (*domainProject.ProjectSPSController, error) {
		return txService.assignments().assignSPSController(ctx, projectID, spsControllerID)
	})
}

func (s *ProjectFacilityLinkService) CopySPSController(ctx context.Context, projectID, spsControllerID uuid.UUID) (*domainFacility.SPSController, error) {
	return withProjectFacilityLinkTxResult(s, func(txService *ProjectFacilityLinkService) (*domainFacility.SPSController, error) {
		return txService.copySPSController(ctx, projectID, spsControllerID)
	})
}

func (s *ProjectFacilityLinkService) copySPSController(ctx context.Context, projectID, spsControllerID uuid.UUID) (*domainFacility.SPSController, error) {
	copyEntity, err := s.hierarchyCopier.CopySPSControllerByID(ctx, spsControllerID)
	if err != nil {
		return nil, err
	}

	if _, err := s.assignments().assignSPSController(ctx, projectID, copyEntity.ID); err != nil {
		return nil, err
	}

	return copyEntity, nil
}

func (s *ProjectFacilityLinkService) CopySPSControllerSystemType(ctx context.Context, projectID, systemTypeID uuid.UUID) (*domainFacility.SPSControllerSystemType, error) {
	return withProjectFacilityLinkTxResult(s, func(txService *ProjectFacilityLinkService) (*domainFacility.SPSControllerSystemType, error) {
		return txService.copySPSControllerSystemType(ctx, projectID, systemTypeID)
	})
}

func (s *ProjectFacilityLinkService) copySPSControllerSystemType(ctx context.Context, projectID, systemTypeID uuid.UUID) (*domainFacility.SPSControllerSystemType, error) {
	copyEntity, err := s.hierarchyCopier.CopySPSControllerSystemTypeByID(ctx, systemTypeID)
	if err != nil {
		return nil, err
	}

	if err := s.assignments().assignSPSControllerSystemType(ctx, projectID, copyEntity.ID); err != nil {
		return nil, err
	}

	return copyEntity, nil
}

func (s *ProjectFacilityLinkService) UpdateSPSController(ctx context.Context, linkID, projectID, spsControllerID uuid.UUID) (*domainProject.ProjectSPSController, error) {
	return withProjectFacilityLinkTxResult(s, func(txService *ProjectFacilityLinkService) (*domainProject.ProjectSPSController, error) {
		return txService.assignments().updateSPSController(ctx, linkID, projectID, spsControllerID)
	})
}

func (s *ProjectFacilityLinkService) DeleteSPSController(ctx context.Context, linkID, projectID uuid.UUID) error {
	return s.withTx(func(txService *ProjectFacilityLinkService) error {
		return txService.assignments().removeSPSController(ctx, linkID, projectID)
	})
}

func (s *ProjectFacilityLinkService) CreateFieldDevice(ctx context.Context, projectID, fieldDeviceID uuid.UUID) (*domainProject.ProjectFieldDevice, error) {
	return withProjectFacilityLinkTxResult(s, func(txService *ProjectFacilityLinkService) (*domainProject.ProjectFieldDevice, error) {
		return txService.assignments().assignFieldDevice(ctx, projectID, fieldDeviceID)
	})
}

func (s *ProjectFacilityLinkService) UpdateFieldDevice(ctx context.Context, linkID, projectID, fieldDeviceID uuid.UUID) (*domainProject.ProjectFieldDevice, error) {
	return s.assignments().updateFieldDevice(ctx, linkID, projectID, fieldDeviceID)
}

func (s *ProjectFacilityLinkService) DeleteFieldDevice(ctx context.Context, linkID, projectID uuid.UUID) error {
	return s.withTx(func(txService *ProjectFacilityLinkService) error {
		return txService.assignments().removeFieldDevice(ctx, linkID, projectID)
	})
}

func (s *ProjectFacilityLinkService) AddObjectData(ctx context.Context, projectID, objectDataID uuid.UUID) (*domainFacility.ObjectData, error) {
	if _, err := domain.GetByID(ctx, s.projectRepo, projectID); err != nil {
		return nil, err
	}
	obj, err := domain.GetByID(ctx, s.objectDataRepo, objectDataID)
	if err != nil {
		return nil, err
	}
	if obj.ProjectID != nil && *obj.ProjectID != projectID {
		return nil, domain.ErrConflict
	}
	if obj.ProjectID == nil {
		obj.ProjectID = &projectID
	}
	obj.IsActive = true
	if err := s.objectDataRepo.Update(ctx, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func (s *ProjectFacilityLinkService) RemoveObjectData(ctx context.Context, projectID, objectDataID uuid.UUID) (*domainFacility.ObjectData, error) {
	if _, err := domain.GetByID(ctx, s.projectRepo, projectID); err != nil {
		return nil, err
	}
	obj, err := domain.GetByID(ctx, s.objectDataRepo, objectDataID)
	if err != nil {
		return nil, err
	}
	if obj.ProjectID == nil || *obj.ProjectID != projectID {
		return nil, domain.ErrNotFound
	}
	obj.IsActive = false
	if err := s.objectDataRepo.Update(ctx, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func (s *ProjectFacilityLinkService) ListControlCabinets(ctx context.Context, projectID uuid.UUID, page, limit int) (*domain.PaginatedList[domainProject.ProjectControlCabinet], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.projectControlCabinetRepo.GetPaginatedListByProjectID(ctx, projectID, domain.PaginationParams{Page: page, Limit: limit})
}

func (s *ProjectFacilityLinkService) ListSPSControllers(ctx context.Context, projectID uuid.UUID, page, limit int) (*domain.PaginatedList[domainProject.ProjectSPSController], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.projectSPSControllerRepo.GetPaginatedListByProjectID(ctx, projectID, domain.PaginationParams{Page: page, Limit: limit})
}

func (s *ProjectFacilityLinkService) ListFieldDevices(ctx context.Context, projectID uuid.UUID, page, limit int) (*domain.PaginatedList[domainProject.ProjectFieldDevice], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.projectFieldDeviceRepo.GetPaginatedListByProjectID(ctx, projectID, domain.PaginationParams{Page: page, Limit: limit})
}

func (s *ProjectFacilityLinkService) ListObjectData(ctx context.Context, projectID uuid.UUID, page, limit int, search string, apparatID, systemPartID *uuid.UUID) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	params := domain.PaginationParams{Page: page, Limit: limit, Search: search}
	return s.objectDataRepo.GetPaginatedListWithFilters(ctx, params, domainFacility.ObjectDataFilterParams{
		ProjectID:    &projectID,
		ApparatID:    apparatID,
		SystemPartID: systemPartID,
	})
}

func (s *ProjectFacilityLinkService) MultiCreateFieldDevices(ctx context.Context, projectID uuid.UUID, fieldDeviceIDs []uuid.UUID) ([]uuid.UUID, []string) {
	return s.assignments().multiAssignFieldDevices(ctx, projectID, fieldDeviceIDs)
}

func (s *ProjectFacilityLinkService) MultiCreateAndAssignFieldDevices(ctx context.Context, projectID uuid.UUID, items []domainFacility.FieldDeviceCreateItem) (*domainFacility.FieldDeviceMultiCreateResult, error) {
	return withProjectFacilityLinkTxResult(s, func(txService *ProjectFacilityLinkService) (*domainFacility.FieldDeviceMultiCreateResult, error) {
		return txService.multiCreateAndAssignFieldDevices(ctx, projectID, items)
	})
}

func (s *ProjectFacilityLinkService) multiCreateAndAssignFieldDevices(ctx context.Context, projectID uuid.UUID, items []domainFacility.FieldDeviceCreateItem) (*domainFacility.FieldDeviceMultiCreateResult, error) {
	if _, err := domain.GetByID(ctx, s.projectRepo, projectID); err != nil {
		return nil, err
	}
	if s.fieldDeviceCreator == nil {
		return nil, domain.ErrInvalidArgument
	}

	result := s.fieldDeviceCreator.MultiCreate(ctx, items)
	fieldDeviceIDs := successfulFieldDeviceIDs(result)
	if len(fieldDeviceIDs) == 0 {
		return result, nil
	}

	if err := s.assignments().assignFieldDeviceIDs(ctx, projectID, fieldDeviceIDs); err != nil {
		return nil, err
	}

	return result, nil
}

func successfulFieldDeviceIDs(result *domainFacility.FieldDeviceMultiCreateResult) []uuid.UUID {
	if result == nil {
		return nil
	}

	ids := make([]uuid.UUID, 0, result.SuccessCount)
	for _, item := range result.Results {
		if item.Success && item.FieldDevice != nil {
			ids = append(ids, item.FieldDevice.ID)
		}
	}
	return ids
}
