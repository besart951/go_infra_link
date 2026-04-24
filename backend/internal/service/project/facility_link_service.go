package project

import (
	"context"
	"errors"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/google/uuid"
	"gorm.io/gorm"
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
	txRunner                  TxRunner
	txFactory                 func(tx *gorm.DB) (*Services, error)
}

func (s *ProjectFacilityLinkService) withTx(fn func(*ProjectFacilityLinkService) error) error {
	if s.txRunner == nil || s.txFactory == nil {
		return fn(s)
	}

	return s.txRunner(func(tx *gorm.DB) error {
		txServices, err := s.txFactory(tx)
		if err != nil {
			return err
		}
		return fn(txServices.FacilityLink)
	})
}

func withProjectFacilityLinkTxResult[T any](s *ProjectFacilityLinkService, fn func(*ProjectFacilityLinkService) (T, error)) (T, error) {
	var zero T
	if s.txRunner == nil || s.txFactory == nil {
		return fn(s)
	}

	var result T
	err := s.txRunner(func(tx *gorm.DB) error {
		txServices, buildErr := s.txFactory(tx)
		if buildErr != nil {
			return buildErr
		}

		value, runErr := fn(txServices.FacilityLink)
		if runErr != nil {
			return runErr
		}

		result = value
		return nil
	})
	if err != nil {
		return zero, err
	}

	return result, nil
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
		return txService.createControlCabinet(ctx, projectID, controlCabinetID)
	})
}

func (s *ProjectFacilityLinkService) createControlCabinet(ctx context.Context, projectID, controlCabinetID uuid.UUID) (*domainProject.ProjectControlCabinet, error) {
	entity := &domainProject.ProjectControlCabinet{ProjectID: projectID, ControlCabinetID: controlCabinetID}
	if err := s.projectControlCabinetRepo.Create(ctx, entity); err != nil {
		return nil, err
	}

	if err := s.linkDescendantsForControlCabinet(ctx, projectID, controlCabinetID); err != nil {
		_ = s.cleanupProjectLinksForControlCabinetHierarchy(ctx, controlCabinetID)
		return nil, err
	}

	return entity, nil
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

	if _, err := s.createControlCabinet(ctx, projectID, copyEntity.ID); err != nil {
		return nil, err
	}

	return copyEntity, nil
}

func (s *ProjectFacilityLinkService) UpdateControlCabinet(ctx context.Context, linkID, projectID, controlCabinetID uuid.UUID) (*domainProject.ProjectControlCabinet, error) {
	return withProjectFacilityLinkTxResult(s, func(txService *ProjectFacilityLinkService) (*domainProject.ProjectControlCabinet, error) {
		return txService.updateControlCabinet(ctx, linkID, projectID, controlCabinetID)
	})
}

func (s *ProjectFacilityLinkService) updateControlCabinet(ctx context.Context, linkID, projectID, controlCabinetID uuid.UUID) (*domainProject.ProjectControlCabinet, error) {
	entity, err := domain.GetByID(ctx, s.projectControlCabinetRepo, linkID)
	if err != nil {
		return nil, err
	}
	if entity.ProjectID != projectID {
		return nil, domain.ErrNotFound
	}

	previousControlCabinetID := entity.ControlCabinetID
	entity.ControlCabinetID = controlCabinetID
	if err := s.projectControlCabinetRepo.Update(ctx, entity); err != nil {
		return nil, err
	}

	if err := s.linkDescendantsForControlCabinet(ctx, projectID, controlCabinetID); err != nil {
		_ = s.cleanupProjectLinksForControlCabinetHierarchy(ctx, controlCabinetID)
		entity.ControlCabinetID = previousControlCabinetID
		_ = s.projectControlCabinetRepo.Update(ctx, entity)
		return nil, err
	}

	return entity, nil
}

func (s *ProjectFacilityLinkService) DeleteControlCabinet(ctx context.Context, linkID, projectID uuid.UUID) error {
	return s.withTx(func(txService *ProjectFacilityLinkService) error {
		return txService.deleteControlCabinet(ctx, linkID, projectID)
	})
}

func (s *ProjectFacilityLinkService) deleteControlCabinet(ctx context.Context, linkID, projectID uuid.UUID) error {
	entity, err := domain.GetByID(ctx, s.projectControlCabinetRepo, linkID)
	if err != nil {
		return err
	}
	if entity.ProjectID != projectID {
		return domain.ErrNotFound
	}

	controlCabinetID := entity.ControlCabinetID
	spsControllerIDs, _, fieldDeviceIDs, err := s.collectDescendantIDsForControlCabinet(ctx, controlCabinetID)
	if err != nil {
		return err
	}

	if err := s.deleteProjectControlCabinetLinksByControlCabinetIDs(ctx, []uuid.UUID{controlCabinetID}); err != nil {
		return err
	}
	if err := s.deleteProjectSPSControllerLinksBySPSControllerIDs(ctx, spsControllerIDs); err != nil {
		return err
	}
	if err := s.deleteProjectFieldDeviceLinksByFieldDeviceIDs(ctx, fieldDeviceIDs); err != nil {
		return err
	}

	if err := s.deleteFieldDevicesWithChildren(ctx, fieldDeviceIDs); err != nil {
		return err
	}
	if len(spsControllerIDs) > 0 {
		if err := s.spsControllerSystemRepo.DeleteBySPSControllerIDs(ctx, spsControllerIDs); err != nil {
			return err
		}
		if err := s.spsControllerRepo.DeleteByIds(ctx, spsControllerIDs); err != nil {
			return err
		}
	}

	return s.controlCabinetRepo.DeleteByIds(ctx, []uuid.UUID{controlCabinetID})
}

func (s *ProjectFacilityLinkService) CreateSPSController(ctx context.Context, projectID, spsControllerID uuid.UUID) (*domainProject.ProjectSPSController, error) {
	return withProjectFacilityLinkTxResult(s, func(txService *ProjectFacilityLinkService) (*domainProject.ProjectSPSController, error) {
		return txService.createSPSController(ctx, projectID, spsControllerID)
	})
}

func (s *ProjectFacilityLinkService) createSPSController(ctx context.Context, projectID, spsControllerID uuid.UUID) (*domainProject.ProjectSPSController, error) {
	entity := &domainProject.ProjectSPSController{ProjectID: projectID, SPSControllerID: spsControllerID}
	if err := s.projectSPSControllerRepo.Create(ctx, entity); err != nil {
		return nil, err
	}

	if err := s.linkDescendantsForSPSControllers(ctx, projectID, []uuid.UUID{spsControllerID}); err != nil {
		_ = s.cleanupProjectLinksForSPSControllers(ctx, []uuid.UUID{spsControllerID})
		return nil, err
	}

	return entity, nil
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

	if _, err := s.createSPSController(ctx, projectID, copyEntity.ID); err != nil {
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

	if err := s.linkFieldDevicesForSystemTypes(ctx, projectID, []uuid.UUID{copyEntity.ID}); err != nil {
		_ = s.cleanupProjectLinksForSystemTypes(ctx, []uuid.UUID{copyEntity.ID})
		return nil, err
	}

	return copyEntity, nil
}

func (s *ProjectFacilityLinkService) UpdateSPSController(ctx context.Context, linkID, projectID, spsControllerID uuid.UUID) (*domainProject.ProjectSPSController, error) {
	return withProjectFacilityLinkTxResult(s, func(txService *ProjectFacilityLinkService) (*domainProject.ProjectSPSController, error) {
		return txService.updateSPSController(ctx, linkID, projectID, spsControllerID)
	})
}

func (s *ProjectFacilityLinkService) updateSPSController(ctx context.Context, linkID, projectID, spsControllerID uuid.UUID) (*domainProject.ProjectSPSController, error) {
	entity, err := domain.GetByID(ctx, s.projectSPSControllerRepo, linkID)
	if err != nil {
		return nil, err
	}
	if entity.ProjectID != projectID {
		return nil, domain.ErrNotFound
	}

	previousSPSControllerID := entity.SPSControllerID
	entity.SPSControllerID = spsControllerID
	if err := s.projectSPSControllerRepo.Update(ctx, entity); err != nil {
		return nil, err
	}

	if err := s.linkDescendantsForSPSControllers(ctx, projectID, []uuid.UUID{spsControllerID}); err != nil {
		_ = s.cleanupProjectLinksForSPSControllers(ctx, []uuid.UUID{spsControllerID})
		entity.SPSControllerID = previousSPSControllerID
		_ = s.projectSPSControllerRepo.Update(ctx, entity)
		return nil, err
	}

	return entity, nil
}

func (s *ProjectFacilityLinkService) DeleteSPSController(ctx context.Context, linkID, projectID uuid.UUID) error {
	return s.withTx(func(txService *ProjectFacilityLinkService) error {
		return txService.deleteSPSController(ctx, linkID, projectID)
	})
}

func (s *ProjectFacilityLinkService) deleteSPSController(ctx context.Context, linkID, projectID uuid.UUID) error {
	entity, err := domain.GetByID(ctx, s.projectSPSControllerRepo, linkID)
	if err != nil {
		return err
	}
	if entity.ProjectID != projectID {
		return domain.ErrNotFound
	}

	spsControllerID := entity.SPSControllerID
	_, fieldDeviceIDs, err := s.collectDescendantIDsForSPSControllers(ctx, []uuid.UUID{spsControllerID})
	if err != nil {
		return err
	}

	if err := s.deleteProjectSPSControllerLinksBySPSControllerIDs(ctx, []uuid.UUID{spsControllerID}); err != nil {
		return err
	}
	if err := s.deleteProjectFieldDeviceLinksByFieldDeviceIDs(ctx, fieldDeviceIDs); err != nil {
		return err
	}

	if err := s.deleteFieldDevicesWithChildren(ctx, fieldDeviceIDs); err != nil {
		return err
	}
	if err := s.spsControllerSystemRepo.DeleteBySPSControllerIDs(ctx, []uuid.UUID{spsControllerID}); err != nil {
		return err
	}
	return s.spsControllerRepo.DeleteByIds(ctx, []uuid.UUID{spsControllerID})
}

func (s *ProjectFacilityLinkService) CreateFieldDevice(ctx context.Context, projectID, fieldDeviceID uuid.UUID) (*domainProject.ProjectFieldDevice, error) {
	return withProjectFacilityLinkTxResult(s, func(txService *ProjectFacilityLinkService) (*domainProject.ProjectFieldDevice, error) {
		return txService.createFieldDevice(ctx, projectID, fieldDeviceID)
	})
}

func (s *ProjectFacilityLinkService) createFieldDevice(ctx context.Context, projectID, fieldDeviceID uuid.UUID) (*domainProject.ProjectFieldDevice, error) {
	entity := &domainProject.ProjectFieldDevice{ProjectID: projectID, FieldDeviceID: fieldDeviceID}
	if err := s.projectFieldDeviceRepo.Create(ctx, entity); err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *ProjectFacilityLinkService) UpdateFieldDevice(ctx context.Context, linkID, projectID, fieldDeviceID uuid.UUID) (*domainProject.ProjectFieldDevice, error) {
	entity, err := domain.GetByID(ctx, s.projectFieldDeviceRepo, linkID)
	if err != nil {
		return nil, err
	}
	if entity.ProjectID != projectID {
		return nil, domain.ErrNotFound
	}
	entity.FieldDeviceID = fieldDeviceID
	if err := s.projectFieldDeviceRepo.Update(ctx, entity); err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *ProjectFacilityLinkService) DeleteFieldDevice(ctx context.Context, linkID, projectID uuid.UUID) error {
	return s.withTx(func(txService *ProjectFacilityLinkService) error {
		return txService.deleteFieldDevice(ctx, linkID, projectID)
	})
}

func (s *ProjectFacilityLinkService) deleteFieldDevice(ctx context.Context, linkID, projectID uuid.UUID) error {
	entity, err := domain.GetByID(ctx, s.projectFieldDeviceRepo, linkID)
	if err != nil {
		return err
	}
	if entity.ProjectID != projectID {
		return domain.ErrNotFound
	}

	fieldDeviceID := entity.FieldDeviceID
	if err := s.deleteProjectFieldDeviceLinksByFieldDeviceIDs(ctx, []uuid.UUID{fieldDeviceID}); err != nil {
		return err
	}
	return s.deleteFieldDevicesWithChildren(ctx, []uuid.UUID{fieldDeviceID})
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

	switch {
	case apparatID != nil && systemPartID != nil:
		return s.objectDataRepo.GetPaginatedListForProjectByApparatAndSystemPartID(ctx, projectID, *apparatID, *systemPartID, params)
	case apparatID != nil:
		return s.objectDataRepo.GetPaginatedListForProjectByApparatID(ctx, projectID, *apparatID, params)
	case systemPartID != nil:
		return s.objectDataRepo.GetPaginatedListForProjectBySystemPartID(ctx, projectID, *systemPartID, params)
	default:
		return s.objectDataRepo.GetPaginatedListForProject(ctx, projectID, params)
	}
}

func (s *ProjectFacilityLinkService) MultiCreateFieldDevices(ctx context.Context, projectID uuid.UUID, fieldDeviceIDs []uuid.UUID) ([]uuid.UUID, []string) {
	return s.multiCreateFieldDevices(ctx, projectID, fieldDeviceIDs)
}

func (s *ProjectFacilityLinkService) multiCreateFieldDevices(ctx context.Context, projectID uuid.UUID, fieldDeviceIDs []uuid.UUID) ([]uuid.UUID, []string) {
	if _, err := domain.GetByID(ctx, s.projectRepo, projectID); err != nil {
		return nil, []string{"project not found"}
	}

	successIDs := make([]uuid.UUID, 0, len(fieldDeviceIDs))
	associationErrors := make([]string, 0)

	for _, fieldDeviceID := range fieldDeviceIDs {
		entity := &domainProject.ProjectFieldDevice{ProjectID: projectID, FieldDeviceID: fieldDeviceID}
		if err := s.projectFieldDeviceRepo.Create(ctx, entity); err != nil {
			associationErrors = append(associationErrors, err.Error())
			continue
		}
		successIDs = append(successIDs, fieldDeviceID)
	}

	return successIDs, associationErrors
}

func (s *ProjectFacilityLinkService) collectDescendantIDsForControlCabinet(ctx context.Context, controlCabinetID uuid.UUID) ([]uuid.UUID, []uuid.UUID, []uuid.UUID, error) {
	spsControllerIDs, err := s.spsControllerRepo.GetIDsByControlCabinetID(ctx, controlCabinetID)
	if err != nil {
		return nil, nil, nil, err
	}

	systemTypeIDs, fieldDeviceIDs, err := s.collectDescendantIDsForSPSControllers(ctx, spsControllerIDs)
	if err != nil {
		return nil, nil, nil, err
	}

	return spsControllerIDs, systemTypeIDs, fieldDeviceIDs, nil
}

func (s *ProjectFacilityLinkService) collectDescendantIDsForSPSControllers(ctx context.Context, spsControllerIDs []uuid.UUID) ([]uuid.UUID, []uuid.UUID, error) {
	if len(spsControllerIDs) == 0 {
		return nil, nil, nil
	}

	systemTypeIDs, err := s.spsControllerSystemRepo.GetIDsBySPSControllerIDs(ctx, spsControllerIDs)
	if err != nil {
		return nil, nil, err
	}
	if len(systemTypeIDs) == 0 {
		return nil, nil, nil
	}

	fieldDeviceIDs, err := s.fieldDeviceRepo.GetIDsBySPSControllerSystemTypeIDs(ctx, systemTypeIDs)
	if err != nil {
		return nil, nil, err
	}

	return systemTypeIDs, fieldDeviceIDs, nil
}

func (s *ProjectFacilityLinkService) deleteFieldDevicesWithChildren(ctx context.Context, fieldDeviceIDs []uuid.UUID) error {
	if len(fieldDeviceIDs) == 0 {
		return nil
	}

	if err := s.bacnetObjectRepo.DeleteByFieldDeviceIDs(ctx, fieldDeviceIDs); err != nil {
		return err
	}
	if err := s.specificationRepo.DeleteByFieldDeviceIDs(ctx, fieldDeviceIDs); err != nil {
		return err
	}
	return s.fieldDeviceRepo.DeleteByIds(ctx, fieldDeviceIDs)
}

func (s *ProjectFacilityLinkService) deleteProjectControlCabinetLinksByControlCabinetIDs(ctx context.Context, controlCabinetIDs []uuid.UUID) error {
	if len(controlCabinetIDs) == 0 {
		return nil
	}
	return s.projectControlCabinetRepo.DeleteByControlCabinetIDs(ctx, controlCabinetIDs)
}

func (s *ProjectFacilityLinkService) deleteProjectSPSControllerLinksBySPSControllerIDs(ctx context.Context, spsControllerIDs []uuid.UUID) error {
	if len(spsControllerIDs) == 0 {
		return nil
	}
	return s.projectSPSControllerRepo.DeleteBySPSControllerIDs(ctx, spsControllerIDs)
}

func (s *ProjectFacilityLinkService) deleteProjectFieldDeviceLinksByFieldDeviceIDs(ctx context.Context, fieldDeviceIDs []uuid.UUID) error {
	if len(fieldDeviceIDs) == 0 {
		return nil
	}
	return s.projectFieldDeviceRepo.DeleteByFieldDeviceIDs(ctx, fieldDeviceIDs)
}

func (s *ProjectFacilityLinkService) linkDescendantsForControlCabinet(ctx context.Context, projectID, controlCabinetID uuid.UUID) error {
	spsControllerIDs, err := s.spsControllerRepo.GetIDsByControlCabinetID(ctx, controlCabinetID)
	if err != nil {
		return err
	}
	return s.linkDescendantsForSPSControllers(ctx, projectID, spsControllerIDs)
}

func (s *ProjectFacilityLinkService) linkDescendantsForSPSControllers(ctx context.Context, projectID uuid.UUID, spsControllerIDs []uuid.UUID) error {
	if len(spsControllerIDs) == 0 {
		return nil
	}

	existingSPS, err := s.listProjectSPSControllerIDSet(ctx, projectID)
	if err != nil {
		return err
	}
	for _, spsID := range spsControllerIDs {
		if _, ok := existingSPS[spsID]; ok {
			continue
		}
		if err := s.createProjectSPSControllerLink(ctx, projectID, spsID); err != nil {
			return err
		}
		existingSPS[spsID] = struct{}{}
	}

	systemTypeIDs, err := s.spsControllerSystemRepo.GetIDsBySPSControllerIDs(ctx, spsControllerIDs)
	if err != nil {
		return err
	}
	return s.linkFieldDevicesForSystemTypes(ctx, projectID, systemTypeIDs)
}

func (s *ProjectFacilityLinkService) linkFieldDevicesForSystemTypes(ctx context.Context, projectID uuid.UUID, systemTypeIDs []uuid.UUID) error {
	if len(systemTypeIDs) == 0 {
		return nil
	}

	fieldDeviceIDs, err := s.fieldDeviceRepo.GetIDsBySPSControllerSystemTypeIDs(ctx, systemTypeIDs)
	if err != nil {
		return err
	}
	if len(fieldDeviceIDs) == 0 {
		return nil
	}

	existingFieldDevices, err := s.listProjectFieldDeviceIDSet(ctx, projectID)
	if err != nil {
		return err
	}
	for _, fieldDeviceID := range fieldDeviceIDs {
		if _, ok := existingFieldDevices[fieldDeviceID]; ok {
			continue
		}
		if err := s.createProjectFieldDeviceLink(ctx, projectID, fieldDeviceID); err != nil {
			return err
		}
		existingFieldDevices[fieldDeviceID] = struct{}{}
	}

	return nil
}

func (s *ProjectFacilityLinkService) cleanupProjectLinksForControlCabinetHierarchy(ctx context.Context, controlCabinetID uuid.UUID) error {
	spsControllerIDs, _, fieldDeviceIDs, err := s.collectDescendantIDsForControlCabinet(ctx, controlCabinetID)
	if err != nil {
		return err
	}
	if err := s.deleteProjectFieldDeviceLinksByFieldDeviceIDs(ctx, fieldDeviceIDs); err != nil {
		return err
	}
	if err := s.deleteProjectSPSControllerLinksBySPSControllerIDs(ctx, spsControllerIDs); err != nil {
		return err
	}
	return s.deleteProjectControlCabinetLinksByControlCabinetIDs(ctx, []uuid.UUID{controlCabinetID})
}

func (s *ProjectFacilityLinkService) cleanupProjectLinksForSPSControllers(ctx context.Context, spsControllerIDs []uuid.UUID) error {
	if len(spsControllerIDs) == 0 {
		return nil
	}

	_, fieldDeviceIDs, err := s.collectDescendantIDsForSPSControllers(ctx, spsControllerIDs)
	if err != nil {
		return err
	}
	if err := s.deleteProjectFieldDeviceLinksByFieldDeviceIDs(ctx, fieldDeviceIDs); err != nil {
		return err
	}
	return s.deleteProjectSPSControllerLinksBySPSControllerIDs(ctx, spsControllerIDs)
}

func (s *ProjectFacilityLinkService) cleanupProjectLinksForSystemTypes(ctx context.Context, systemTypeIDs []uuid.UUID) error {
	if len(systemTypeIDs) == 0 {
		return nil
	}

	fieldDeviceIDs, err := s.fieldDeviceRepo.GetIDsBySPSControllerSystemTypeIDs(ctx, systemTypeIDs)
	if err != nil {
		return err
	}
	return s.deleteProjectFieldDeviceLinksByFieldDeviceIDs(ctx, fieldDeviceIDs)
}

func (s *ProjectFacilityLinkService) listProjectSPSControllerIDSet(ctx context.Context, projectID uuid.UUID) (map[uuid.UUID]struct{}, error) {
	result := make(map[uuid.UUID]struct{})
	page := 1

	for {
		items, err := s.projectSPSControllerRepo.GetPaginatedListByProjectID(ctx, projectID, domain.PaginationParams{Page: page, Limit: 500})
		if err != nil {
			return nil, err
		}

		for _, item := range items.Items {
			result[item.SPSControllerID] = struct{}{}
		}

		if page >= items.TotalPages || len(items.Items) == 0 {
			break
		}
		page++
	}

	return result, nil
}

func (s *ProjectFacilityLinkService) listProjectFieldDeviceIDSet(ctx context.Context, projectID uuid.UUID) (map[uuid.UUID]struct{}, error) {
	result := make(map[uuid.UUID]struct{})
	page := 1

	for {
		items, err := s.projectFieldDeviceRepo.GetPaginatedListByProjectID(ctx, projectID, domain.PaginationParams{Page: page, Limit: 500})
		if err != nil {
			return nil, err
		}

		for _, item := range items.Items {
			result[item.FieldDeviceID] = struct{}{}
		}

		if page >= items.TotalPages || len(items.Items) == 0 {
			break
		}
		page++
	}

	return result, nil
}

func (s *ProjectFacilityLinkService) createProjectSPSControllerLink(ctx context.Context, projectID, spsControllerID uuid.UUID) error {
	entity := &domainProject.ProjectSPSController{ProjectID: projectID, SPSControllerID: spsControllerID}
	if err := s.projectSPSControllerRepo.Create(ctx, entity); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil
		}
		return err
	}
	return nil
}

func (s *ProjectFacilityLinkService) createProjectFieldDeviceLink(ctx context.Context, projectID, fieldDeviceID uuid.UUID) error {
	entity := &domainProject.ProjectFieldDevice{ProjectID: projectID, FieldDeviceID: fieldDeviceID}
	if err := s.projectFieldDeviceRepo.Create(ctx, entity); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil
		}
		return err
	}
	return nil
}
