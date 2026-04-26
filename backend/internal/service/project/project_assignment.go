package project

import (
	"context"
	"errors"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type projectAssignment struct {
	service *ProjectFacilityLinkService
}

type projectAssignmentKind int

const (
	projectAssignmentControlCabinet projectAssignmentKind = iota
	projectAssignmentSPSController
	projectAssignmentSPSControllerSystemType
	projectAssignmentFieldDevice
)

type projectAssignmentTarget struct {
	kind projectAssignmentKind
	id   uuid.UUID
}

type projectAssignmentResult struct {
	controlCabinet *domainProject.ProjectControlCabinet
	spsController  *domainProject.ProjectSPSController
	fieldDevice    *domainProject.ProjectFieldDevice
}

func (a projectAssignment) assignControlCabinet(ctx context.Context, projectID, controlCabinetID uuid.UUID) (*domainProject.ProjectControlCabinet, error) {
	result, err := a.assign(ctx, projectID, projectAssignmentTarget{kind: projectAssignmentControlCabinet, id: controlCabinetID})
	if err != nil {
		return nil, err
	}
	return result.controlCabinet, nil
}

func (a projectAssignment) assignSPSController(ctx context.Context, projectID, spsControllerID uuid.UUID) (*domainProject.ProjectSPSController, error) {
	result, err := a.assign(ctx, projectID, projectAssignmentTarget{kind: projectAssignmentSPSController, id: spsControllerID})
	if err != nil {
		return nil, err
	}
	return result.spsController, nil
}

func (a projectAssignment) assignSPSControllerSystemType(ctx context.Context, projectID, systemTypeID uuid.UUID) error {
	_, err := a.assign(ctx, projectID, projectAssignmentTarget{kind: projectAssignmentSPSControllerSystemType, id: systemTypeID})
	return err
}

func (a projectAssignment) assignFieldDevice(ctx context.Context, projectID, fieldDeviceID uuid.UUID) (*domainProject.ProjectFieldDevice, error) {
	result, err := a.assign(ctx, projectID, projectAssignmentTarget{kind: projectAssignmentFieldDevice, id: fieldDeviceID})
	if err != nil {
		return nil, err
	}
	return result.fieldDevice, nil
}

func (a projectAssignment) updateControlCabinet(ctx context.Context, linkID, projectID, controlCabinetID uuid.UUID) (*domainProject.ProjectControlCabinet, error) {
	result, err := a.update(ctx, linkID, projectID, projectAssignmentTarget{kind: projectAssignmentControlCabinet, id: controlCabinetID})
	if err != nil {
		return nil, err
	}
	return result.controlCabinet, nil
}

func (a projectAssignment) updateSPSController(ctx context.Context, linkID, projectID, spsControllerID uuid.UUID) (*domainProject.ProjectSPSController, error) {
	result, err := a.update(ctx, linkID, projectID, projectAssignmentTarget{kind: projectAssignmentSPSController, id: spsControllerID})
	if err != nil {
		return nil, err
	}
	return result.spsController, nil
}

func (a projectAssignment) updateFieldDevice(ctx context.Context, linkID, projectID, fieldDeviceID uuid.UUID) (*domainProject.ProjectFieldDevice, error) {
	result, err := a.update(ctx, linkID, projectID, projectAssignmentTarget{kind: projectAssignmentFieldDevice, id: fieldDeviceID})
	if err != nil {
		return nil, err
	}
	return result.fieldDevice, nil
}

func (a projectAssignment) removeControlCabinet(ctx context.Context, linkID, projectID uuid.UUID) error {
	return a.remove(ctx, linkID, projectID, projectAssignmentControlCabinet)
}

func (a projectAssignment) removeSPSController(ctx context.Context, linkID, projectID uuid.UUID) error {
	return a.remove(ctx, linkID, projectID, projectAssignmentSPSController)
}

func (a projectAssignment) removeFieldDevice(ctx context.Context, linkID, projectID uuid.UUID) error {
	return a.remove(ctx, linkID, projectID, projectAssignmentFieldDevice)
}

func (s *ProjectFacilityLinkService) assignments() projectAssignment {
	return projectAssignment{service: s}
}

func (a projectAssignment) assign(ctx context.Context, projectID uuid.UUID, target projectAssignmentTarget) (*projectAssignmentResult, error) {
	switch target.kind {
	case projectAssignmentControlCabinet:
		entity := &domainProject.ProjectControlCabinet{ProjectID: projectID, ControlCabinetID: target.id}
		if err := a.service.projectControlCabinetRepo.Create(ctx, entity); err != nil {
			return nil, err
		}
		if err := a.assignControlCabinetDescendants(ctx, projectID, target.id); err != nil {
			return nil, err
		}
		return &projectAssignmentResult{controlCabinet: entity}, nil
	case projectAssignmentSPSController:
		entity := &domainProject.ProjectSPSController{ProjectID: projectID, SPSControllerID: target.id}
		if err := a.service.projectSPSControllerRepo.Create(ctx, entity); err != nil {
			return nil, err
		}
		if err := a.assignSPSControllerDescendants(ctx, projectID, []uuid.UUID{target.id}); err != nil {
			return nil, err
		}
		return &projectAssignmentResult{spsController: entity}, nil
	case projectAssignmentSPSControllerSystemType:
		if err := a.assignFieldDevicesForSystemTypes(ctx, projectID, []uuid.UUID{target.id}); err != nil {
			return nil, err
		}
		return &projectAssignmentResult{}, nil
	case projectAssignmentFieldDevice:
		entity := &domainProject.ProjectFieldDevice{ProjectID: projectID, FieldDeviceID: target.id}
		if err := a.service.projectFieldDeviceRepo.Create(ctx, entity); err != nil {
			return nil, err
		}
		return &projectAssignmentResult{fieldDevice: entity}, nil
	default:
		return nil, domain.ErrInvalidArgument
	}
}

func (a projectAssignment) update(ctx context.Context, linkID, projectID uuid.UUID, target projectAssignmentTarget) (*projectAssignmentResult, error) {
	switch target.kind {
	case projectAssignmentControlCabinet:
		entity, err := domain.GetByID(ctx, a.service.projectControlCabinetRepo, linkID)
		if err != nil {
			return nil, err
		}
		if entity.ProjectID != projectID {
			return nil, domain.ErrNotFound
		}
		entity.ControlCabinetID = target.id
		if err := a.service.projectControlCabinetRepo.Update(ctx, entity); err != nil {
			return nil, err
		}
		if err := a.assignControlCabinetDescendants(ctx, projectID, target.id); err != nil {
			return nil, err
		}
		return &projectAssignmentResult{controlCabinet: entity}, nil
	case projectAssignmentSPSController:
		entity, err := domain.GetByID(ctx, a.service.projectSPSControllerRepo, linkID)
		if err != nil {
			return nil, err
		}
		if entity.ProjectID != projectID {
			return nil, domain.ErrNotFound
		}
		entity.SPSControllerID = target.id
		if err := a.service.projectSPSControllerRepo.Update(ctx, entity); err != nil {
			return nil, err
		}
		if err := a.assignSPSControllerDescendants(ctx, projectID, []uuid.UUID{target.id}); err != nil {
			return nil, err
		}
		return &projectAssignmentResult{spsController: entity}, nil
	case projectAssignmentFieldDevice:
		entity, err := domain.GetByID(ctx, a.service.projectFieldDeviceRepo, linkID)
		if err != nil {
			return nil, err
		}
		if entity.ProjectID != projectID {
			return nil, domain.ErrNotFound
		}
		entity.FieldDeviceID = target.id
		if err := a.service.projectFieldDeviceRepo.Update(ctx, entity); err != nil {
			return nil, err
		}
		return &projectAssignmentResult{fieldDevice: entity}, nil
	default:
		return nil, domain.ErrInvalidArgument
	}
}

func (a projectAssignment) remove(ctx context.Context, linkID, projectID uuid.UUID, kind projectAssignmentKind) error {
	switch kind {
	case projectAssignmentControlCabinet:
		entity, err := domain.GetByID(ctx, a.service.projectControlCabinetRepo, linkID)
		if err != nil {
			return err
		}
		if entity.ProjectID != projectID {
			return domain.ErrNotFound
		}

		controlCabinetID := entity.ControlCabinetID
		spsControllerIDs, _, fieldDeviceIDs, err := a.collectDescendantIDsForControlCabinet(ctx, controlCabinetID)
		if err != nil {
			return err
		}

		if err := a.deleteControlCabinetAssignments(ctx, []uuid.UUID{controlCabinetID}); err != nil {
			return err
		}
		if err := a.deleteSPSControllerAssignments(ctx, spsControllerIDs); err != nil {
			return err
		}
		if err := a.deleteFieldDeviceAssignments(ctx, fieldDeviceIDs); err != nil {
			return err
		}

		if err := a.deleteFieldDevicesWithChildren(ctx, fieldDeviceIDs); err != nil {
			return err
		}
		if len(spsControllerIDs) > 0 {
			if err := a.service.spsControllerSystemRepo.DeleteBySPSControllerIDs(ctx, spsControllerIDs); err != nil {
				return err
			}
			if err := a.service.spsControllerRepo.DeleteByIds(ctx, spsControllerIDs); err != nil {
				return err
			}
		}

		return a.service.controlCabinetRepo.DeleteByIds(ctx, []uuid.UUID{controlCabinetID})
	case projectAssignmentSPSController:
		entity, err := domain.GetByID(ctx, a.service.projectSPSControllerRepo, linkID)
		if err != nil {
			return err
		}
		if entity.ProjectID != projectID {
			return domain.ErrNotFound
		}

		spsControllerID := entity.SPSControllerID
		_, fieldDeviceIDs, err := a.collectDescendantIDsForSPSControllers(ctx, []uuid.UUID{spsControllerID})
		if err != nil {
			return err
		}

		if err := a.deleteSPSControllerAssignments(ctx, []uuid.UUID{spsControllerID}); err != nil {
			return err
		}
		if err := a.deleteFieldDeviceAssignments(ctx, fieldDeviceIDs); err != nil {
			return err
		}

		if err := a.deleteFieldDevicesWithChildren(ctx, fieldDeviceIDs); err != nil {
			return err
		}
		if err := a.service.spsControllerSystemRepo.DeleteBySPSControllerIDs(ctx, []uuid.UUID{spsControllerID}); err != nil {
			return err
		}
		return a.service.spsControllerRepo.DeleteByIds(ctx, []uuid.UUID{spsControllerID})
	case projectAssignmentFieldDevice:
		entity, err := domain.GetByID(ctx, a.service.projectFieldDeviceRepo, linkID)
		if err != nil {
			return err
		}
		if entity.ProjectID != projectID {
			return domain.ErrNotFound
		}

		fieldDeviceID := entity.FieldDeviceID
		if err := a.deleteFieldDeviceAssignments(ctx, []uuid.UUID{fieldDeviceID}); err != nil {
			return err
		}
		return a.deleteFieldDevicesWithChildren(ctx, []uuid.UUID{fieldDeviceID})
	default:
		return domain.ErrInvalidArgument
	}
}

func (a projectAssignment) multiAssignFieldDevices(ctx context.Context, projectID uuid.UUID, fieldDeviceIDs []uuid.UUID) ([]uuid.UUID, []string) {
	if _, err := domain.GetByID(ctx, a.service.projectRepo, projectID); err != nil {
		return nil, []string{"project not found"}
	}

	successIDs := make([]uuid.UUID, 0, len(fieldDeviceIDs))
	associationErrors := make([]string, 0)

	for _, fieldDeviceID := range fieldDeviceIDs {
		if _, err := a.assign(ctx, projectID, projectAssignmentTarget{kind: projectAssignmentFieldDevice, id: fieldDeviceID}); err != nil {
			associationErrors = append(associationErrors, err.Error())
			continue
		}
		successIDs = append(successIDs, fieldDeviceID)
	}

	return successIDs, associationErrors
}

func (a projectAssignment) assignControlCabinetDescendants(ctx context.Context, projectID, controlCabinetID uuid.UUID) error {
	spsControllerIDs, err := a.service.spsControllerRepo.GetIDsByControlCabinetID(ctx, controlCabinetID)
	if err != nil {
		return err
	}
	return a.assignSPSControllerDescendants(ctx, projectID, spsControllerIDs)
}

func (a projectAssignment) assignSPSControllerDescendants(ctx context.Context, projectID uuid.UUID, spsControllerIDs []uuid.UUID) error {
	if len(spsControllerIDs) == 0 {
		return nil
	}

	existingSPS, err := a.listProjectSPSControllerIDSet(ctx, projectID)
	if err != nil {
		return err
	}
	for _, spsID := range spsControllerIDs {
		if _, ok := existingSPS[spsID]; ok {
			continue
		}
		if err := a.createProjectSPSControllerAssignment(ctx, projectID, spsID); err != nil {
			return err
		}
		existingSPS[spsID] = struct{}{}
	}

	systemTypeIDs, err := a.service.spsControllerSystemRepo.GetIDsBySPSControllerIDs(ctx, spsControllerIDs)
	if err != nil {
		return err
	}
	return a.assignFieldDevicesForSystemTypes(ctx, projectID, systemTypeIDs)
}

func (a projectAssignment) assignFieldDevicesForSystemTypes(ctx context.Context, projectID uuid.UUID, systemTypeIDs []uuid.UUID) error {
	if len(systemTypeIDs) == 0 {
		return nil
	}

	fieldDeviceIDs, err := a.service.fieldDeviceRepo.GetIDsBySPSControllerSystemTypeIDs(ctx, systemTypeIDs)
	if err != nil {
		return err
	}
	if len(fieldDeviceIDs) == 0 {
		return nil
	}

	existingFieldDevices, err := a.listProjectFieldDeviceIDSet(ctx, projectID)
	if err != nil {
		return err
	}
	for _, fieldDeviceID := range fieldDeviceIDs {
		if _, ok := existingFieldDevices[fieldDeviceID]; ok {
			continue
		}
		if err := a.createProjectFieldDeviceAssignment(ctx, projectID, fieldDeviceID); err != nil {
			return err
		}
		existingFieldDevices[fieldDeviceID] = struct{}{}
	}

	return nil
}

func (a projectAssignment) collectDescendantIDsForControlCabinet(ctx context.Context, controlCabinetID uuid.UUID) ([]uuid.UUID, []uuid.UUID, []uuid.UUID, error) {
	spsControllerIDs, err := a.service.spsControllerRepo.GetIDsByControlCabinetID(ctx, controlCabinetID)
	if err != nil {
		return nil, nil, nil, err
	}

	systemTypeIDs, fieldDeviceIDs, err := a.collectDescendantIDsForSPSControllers(ctx, spsControllerIDs)
	if err != nil {
		return nil, nil, nil, err
	}

	return spsControllerIDs, systemTypeIDs, fieldDeviceIDs, nil
}

func (a projectAssignment) collectDescendantIDsForSPSControllers(ctx context.Context, spsControllerIDs []uuid.UUID) ([]uuid.UUID, []uuid.UUID, error) {
	if len(spsControllerIDs) == 0 {
		return nil, nil, nil
	}

	systemTypeIDs, err := a.service.spsControllerSystemRepo.GetIDsBySPSControllerIDs(ctx, spsControllerIDs)
	if err != nil {
		return nil, nil, err
	}
	if len(systemTypeIDs) == 0 {
		return nil, nil, nil
	}

	fieldDeviceIDs, err := a.service.fieldDeviceRepo.GetIDsBySPSControllerSystemTypeIDs(ctx, systemTypeIDs)
	if err != nil {
		return nil, nil, err
	}

	return systemTypeIDs, fieldDeviceIDs, nil
}

func (a projectAssignment) deleteFieldDevicesWithChildren(ctx context.Context, fieldDeviceIDs []uuid.UUID) error {
	if len(fieldDeviceIDs) == 0 {
		return nil
	}

	if err := a.service.bacnetObjectRepo.DeleteByFieldDeviceIDs(ctx, fieldDeviceIDs); err != nil {
		return err
	}
	if err := a.service.specificationRepo.DeleteByFieldDeviceIDs(ctx, fieldDeviceIDs); err != nil {
		return err
	}
	return a.service.fieldDeviceRepo.DeleteByIds(ctx, fieldDeviceIDs)
}

func (a projectAssignment) deleteControlCabinetAssignments(ctx context.Context, controlCabinetIDs []uuid.UUID) error {
	if len(controlCabinetIDs) == 0 {
		return nil
	}
	return a.service.projectControlCabinetRepo.DeleteByControlCabinetIDs(ctx, controlCabinetIDs)
}

func (a projectAssignment) deleteSPSControllerAssignments(ctx context.Context, spsControllerIDs []uuid.UUID) error {
	if len(spsControllerIDs) == 0 {
		return nil
	}
	return a.service.projectSPSControllerRepo.DeleteBySPSControllerIDs(ctx, spsControllerIDs)
}

func (a projectAssignment) deleteFieldDeviceAssignments(ctx context.Context, fieldDeviceIDs []uuid.UUID) error {
	if len(fieldDeviceIDs) == 0 {
		return nil
	}
	return a.service.projectFieldDeviceRepo.DeleteByFieldDeviceIDs(ctx, fieldDeviceIDs)
}

func (a projectAssignment) listProjectSPSControllerIDSet(ctx context.Context, projectID uuid.UUID) (map[uuid.UUID]struct{}, error) {
	result := make(map[uuid.UUID]struct{})
	page := 1

	for {
		items, err := a.service.projectSPSControllerRepo.GetPaginatedListByProjectID(ctx, projectID, domain.PaginationParams{Page: page, Limit: 500})
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

func (a projectAssignment) listProjectFieldDeviceIDSet(ctx context.Context, projectID uuid.UUID) (map[uuid.UUID]struct{}, error) {
	result := make(map[uuid.UUID]struct{})
	page := 1

	for {
		items, err := a.service.projectFieldDeviceRepo.GetPaginatedListByProjectID(ctx, projectID, domain.PaginationParams{Page: page, Limit: 500})
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

func (a projectAssignment) createProjectSPSControllerAssignment(ctx context.Context, projectID, spsControllerID uuid.UUID) error {
	entity := &domainProject.ProjectSPSController{ProjectID: projectID, SPSControllerID: spsControllerID}
	if err := a.service.projectSPSControllerRepo.Create(ctx, entity); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil
		}
		return err
	}
	return nil
}

func (a projectAssignment) createProjectFieldDeviceAssignment(ctx context.Context, projectID, fieldDeviceID uuid.UUID) error {
	entity := &domainProject.ProjectFieldDevice{ProjectID: projectID, FieldDeviceID: fieldDeviceID}
	if err := a.service.projectFieldDeviceRepo.Create(ctx, entity); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil
		}
		return err
	}
	return nil
}
