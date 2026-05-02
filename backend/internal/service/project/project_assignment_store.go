package project

import (
	"context"
	"errors"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type projectSPSControllerBulkCreator interface {
	BulkCreate(ctx context.Context, entities []*domainProject.ProjectSPSController, batchSize int) error
}

type projectSPSControllerSetCreator interface {
	BulkCreateBySPSControllerIDs(ctx context.Context, projectID uuid.UUID, spsControllerIDs []uuid.UUID) error
}

type projectFieldDeviceBulkCreator interface {
	BulkCreate(ctx context.Context, entities []*domainProject.ProjectFieldDevice, batchSize int) error
}

type projectFieldDeviceSetCreator interface {
	BulkCreateByFieldDeviceIDs(ctx context.Context, projectID uuid.UUID, fieldDeviceIDs []uuid.UUID) error
	BulkCreateBySPSControllerSystemTypeIDs(ctx context.Context, projectID uuid.UUID, systemTypeIDs []uuid.UUID) error
}

type spsControllerSystemTypeFieldDeviceDeleter interface {
	DeleteBySPSControllerSystemTypeIDs(ctx context.Context, systemTypeIDs []uuid.UUID) error
}

type projectAssignmentStore struct {
	projectControlCabinetRepo domainProject.ProjectControlCabinetRepository
	projectSPSControllerRepo  domainProject.ProjectSPSControllerRepository
	projectFieldDeviceRepo    domainProject.ProjectFieldDeviceRepository
	spsControllerSystemRepo   domainFacility.SPSControllerSystemTypeStore
	fieldDeviceRepo           domainFacility.FieldDeviceStore
	specificationRepo         domainFacility.SpecificationStore
	bacnetObjectRepo          domainFacility.BacnetObjectStore
}

func newProjectAssignmentStore(deps projectAssignmentDependencies) projectAssignmentStore {
	return projectAssignmentStore{
		projectControlCabinetRepo: deps.projectControlCabinetRepo,
		projectSPSControllerRepo:  deps.projectSPSControllerRepo,
		projectFieldDeviceRepo:    deps.projectFieldDeviceRepo,
		spsControllerSystemRepo:   deps.spsControllerSystemRepo,
		fieldDeviceRepo:           deps.fieldDeviceRepo,
		specificationRepo:         deps.specificationRepo,
		bacnetObjectRepo:          deps.bacnetObjectRepo,
	}
}

func (s projectAssignmentStore) assignFieldDeviceIDs(ctx context.Context, projectID uuid.UUID, fieldDeviceIDs []uuid.UUID) error {
	if len(fieldDeviceIDs) == 0 {
		return nil
	}

	if repo, ok := s.projectFieldDeviceRepo.(projectFieldDeviceSetCreator); ok {
		return repo.BulkCreateByFieldDeviceIDs(ctx, projectID, fieldDeviceIDs)
	}

	existingFieldDevices, err := s.listProjectFieldDeviceIDSet(ctx, projectID)
	if err != nil {
		return err
	}

	toCreate := make([]uuid.UUID, 0, len(fieldDeviceIDs))
	for _, fieldDeviceID := range fieldDeviceIDs {
		if _, ok := existingFieldDevices[fieldDeviceID]; ok {
			continue
		}
		toCreate = append(toCreate, fieldDeviceID)
		existingFieldDevices[fieldDeviceID] = struct{}{}
	}

	return s.createProjectFieldDeviceAssignments(ctx, projectID, toCreate)
}

func (s projectAssignmentStore) assignSPSControllerDescendants(ctx context.Context, projectID uuid.UUID, spsControllerIDs []uuid.UUID) error {
	if len(spsControllerIDs) == 0 {
		return nil
	}

	if repo, ok := s.projectSPSControllerRepo.(projectSPSControllerSetCreator); ok {
		if err := repo.BulkCreateBySPSControllerIDs(ctx, projectID, spsControllerIDs); err != nil {
			return err
		}
	} else {
		existingSPS, err := s.listProjectSPSControllerIDSet(ctx, projectID)
		if err != nil {
			return err
		}
		missingSPS := make([]uuid.UUID, 0, len(spsControllerIDs))
		for _, spsID := range spsControllerIDs {
			if _, ok := existingSPS[spsID]; ok {
				continue
			}
			missingSPS = append(missingSPS, spsID)
			existingSPS[spsID] = struct{}{}
		}
		if err := s.createProjectSPSControllerAssignments(ctx, projectID, missingSPS); err != nil {
			return err
		}
	}

	systemTypeIDs, err := s.spsControllerSystemRepo.GetIDsBySPSControllerIDs(ctx, spsControllerIDs)
	if err != nil {
		return err
	}
	return s.assignFieldDevicesForSystemTypes(ctx, projectID, systemTypeIDs)
}

func (s projectAssignmentStore) assignFieldDevicesForSystemTypes(ctx context.Context, projectID uuid.UUID, systemTypeIDs []uuid.UUID) error {
	if len(systemTypeIDs) == 0 {
		return nil
	}

	if repo, ok := s.projectFieldDeviceRepo.(projectFieldDeviceSetCreator); ok {
		return repo.BulkCreateBySPSControllerSystemTypeIDs(ctx, projectID, systemTypeIDs)
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
	missingFieldDeviceIDs := make([]uuid.UUID, 0, len(fieldDeviceIDs))
	for _, fieldDeviceID := range fieldDeviceIDs {
		if _, ok := existingFieldDevices[fieldDeviceID]; ok {
			continue
		}
		missingFieldDeviceIDs = append(missingFieldDeviceIDs, fieldDeviceID)
		existingFieldDevices[fieldDeviceID] = struct{}{}
	}

	return s.createProjectFieldDeviceAssignments(ctx, projectID, missingFieldDeviceIDs)
}

func (s projectAssignmentStore) deleteFieldDeviceHierarchyForSystemTypes(ctx context.Context, systemTypeIDs []uuid.UUID) error {
	if len(systemTypeIDs) == 0 {
		return nil
	}

	projectLinks, okProjectLinks := s.projectFieldDeviceRepo.(spsControllerSystemTypeFieldDeviceDeleter)
	bacnetObjects, okBacnetObjects := s.bacnetObjectRepo.(spsControllerSystemTypeFieldDeviceDeleter)
	specifications, okSpecifications := s.specificationRepo.(spsControllerSystemTypeFieldDeviceDeleter)
	fieldDevices, okFieldDevices := s.fieldDeviceRepo.(spsControllerSystemTypeFieldDeviceDeleter)
	if okProjectLinks && okBacnetObjects && okSpecifications && okFieldDevices {
		if err := projectLinks.DeleteBySPSControllerSystemTypeIDs(ctx, systemTypeIDs); err != nil {
			return err
		}
		if err := bacnetObjects.DeleteBySPSControllerSystemTypeIDs(ctx, systemTypeIDs); err != nil {
			return err
		}
		if err := specifications.DeleteBySPSControllerSystemTypeIDs(ctx, systemTypeIDs); err != nil {
			return err
		}
		return fieldDevices.DeleteBySPSControllerSystemTypeIDs(ctx, systemTypeIDs)
	}

	fieldDeviceIDs, err := s.fieldDeviceRepo.GetIDsBySPSControllerSystemTypeIDs(ctx, systemTypeIDs)
	if err != nil {
		return err
	}
	if err := s.deleteFieldDeviceAssignments(ctx, fieldDeviceIDs); err != nil {
		return err
	}
	return s.deleteFieldDevicesWithChildren(ctx, fieldDeviceIDs)
}

func (s projectAssignmentStore) deleteFieldDevicesWithChildren(ctx context.Context, fieldDeviceIDs []uuid.UUID) error {
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

func (s projectAssignmentStore) deleteControlCabinetAssignments(ctx context.Context, controlCabinetIDs []uuid.UUID) error {
	if len(controlCabinetIDs) == 0 {
		return nil
	}
	return s.projectControlCabinetRepo.DeleteByControlCabinetIDs(ctx, controlCabinetIDs)
}

func (s projectAssignmentStore) deleteSPSControllerAssignments(ctx context.Context, spsControllerIDs []uuid.UUID) error {
	if len(spsControllerIDs) == 0 {
		return nil
	}
	return s.projectSPSControllerRepo.DeleteBySPSControllerIDs(ctx, spsControllerIDs)
}

func (s projectAssignmentStore) deleteFieldDeviceAssignments(ctx context.Context, fieldDeviceIDs []uuid.UUID) error {
	if len(fieldDeviceIDs) == 0 {
		return nil
	}
	return s.projectFieldDeviceRepo.DeleteByFieldDeviceIDs(ctx, fieldDeviceIDs)
}

func (s projectAssignmentStore) listProjectSPSControllerIDSet(ctx context.Context, projectID uuid.UUID) (map[uuid.UUID]struct{}, error) {
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

func (s projectAssignmentStore) listProjectFieldDeviceIDSet(ctx context.Context, projectID uuid.UUID) (map[uuid.UUID]struct{}, error) {
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

func (s projectAssignmentStore) createProjectSPSControllerAssignments(ctx context.Context, projectID uuid.UUID, spsControllerIDs []uuid.UUID) error {
	if len(spsControllerIDs) == 0 {
		return nil
	}

	entities := make([]*domainProject.ProjectSPSController, 0, len(spsControllerIDs))
	for _, spsControllerID := range spsControllerIDs {
		entities = append(entities, &domainProject.ProjectSPSController{ProjectID: projectID, SPSControllerID: spsControllerID})
	}

	if repo, ok := s.projectSPSControllerRepo.(projectSPSControllerBulkCreator); ok {
		return repo.BulkCreate(ctx, entities, 200)
	}

	for _, entity := range entities {
		if err := s.projectSPSControllerRepo.Create(ctx, entity); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				continue
			}
			return err
		}
	}
	return nil
}

func (s projectAssignmentStore) createProjectFieldDeviceAssignments(ctx context.Context, projectID uuid.UUID, fieldDeviceIDs []uuid.UUID) error {
	if len(fieldDeviceIDs) == 0 {
		return nil
	}

	entities := make([]*domainProject.ProjectFieldDevice, 0, len(fieldDeviceIDs))
	for _, fieldDeviceID := range fieldDeviceIDs {
		entities = append(entities, &domainProject.ProjectFieldDevice{ProjectID: projectID, FieldDeviceID: fieldDeviceID})
	}

	if repo, ok := s.projectFieldDeviceRepo.(projectFieldDeviceBulkCreator); ok {
		return repo.BulkCreate(ctx, entities, 200)
	}

	for _, entity := range entities {
		if err := s.projectFieldDeviceRepo.Create(ctx, entity); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				continue
			}
			return err
		}
	}
	return nil
}
