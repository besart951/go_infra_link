package project

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
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

func (a projectAssignment) store() projectAssignmentStore {
	return newProjectAssignmentStore(a.service)
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
		spsControllerIDs, err := a.service.spsControllerRepo.GetIDsByControlCabinetID(ctx, controlCabinetID)
		if err != nil {
			return err
		}
		systemTypeIDs, err := a.collectSystemTypeIDsForSPSControllers(ctx, spsControllerIDs)
		if err != nil {
			return err
		}

		if err := a.deleteControlCabinetAssignments(ctx, []uuid.UUID{controlCabinetID}); err != nil {
			return err
		}
		if err := a.deleteSPSControllerAssignments(ctx, spsControllerIDs); err != nil {
			return err
		}
		if err := a.deleteFieldDeviceHierarchyForSystemTypes(ctx, systemTypeIDs); err != nil {
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
		systemTypeIDs, err := a.collectSystemTypeIDsForSPSControllers(ctx, []uuid.UUID{spsControllerID})
		if err != nil {
			return err
		}

		if err := a.deleteSPSControllerAssignments(ctx, []uuid.UUID{spsControllerID}); err != nil {
			return err
		}
		if err := a.deleteFieldDeviceHierarchyForSystemTypes(ctx, systemTypeIDs); err != nil {
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

func (a projectAssignment) assignFieldDeviceIDs(ctx context.Context, projectID uuid.UUID, fieldDeviceIDs []uuid.UUID) error {
	return a.store().assignFieldDeviceIDs(ctx, projectID, fieldDeviceIDs)
}

func (a projectAssignment) assignControlCabinetDescendants(ctx context.Context, projectID, controlCabinetID uuid.UUID) error {
	spsControllerIDs, err := a.service.spsControllerRepo.GetIDsByControlCabinetID(ctx, controlCabinetID)
	if err != nil {
		return err
	}
	return a.assignSPSControllerDescendants(ctx, projectID, spsControllerIDs)
}

func (a projectAssignment) assignSPSControllerDescendants(ctx context.Context, projectID uuid.UUID, spsControllerIDs []uuid.UUID) error {
	return a.store().assignSPSControllerDescendants(ctx, projectID, spsControllerIDs)
}

func (a projectAssignment) assignFieldDevicesForSystemTypes(ctx context.Context, projectID uuid.UUID, systemTypeIDs []uuid.UUID) error {
	return a.store().assignFieldDevicesForSystemTypes(ctx, projectID, systemTypeIDs)
}

func (a projectAssignment) collectSystemTypeIDsForSPSControllers(ctx context.Context, spsControllerIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(spsControllerIDs) == 0 {
		return nil, nil
	}
	return a.service.spsControllerSystemRepo.GetIDsBySPSControllerIDs(ctx, spsControllerIDs)
}

func (a projectAssignment) deleteFieldDeviceHierarchyForSystemTypes(ctx context.Context, systemTypeIDs []uuid.UUID) error {
	return a.store().deleteFieldDeviceHierarchyForSystemTypes(ctx, systemTypeIDs)
}

func (a projectAssignment) deleteFieldDevicesWithChildren(ctx context.Context, fieldDeviceIDs []uuid.UUID) error {
	return a.store().deleteFieldDevicesWithChildren(ctx, fieldDeviceIDs)
}

func (a projectAssignment) deleteControlCabinetAssignments(ctx context.Context, controlCabinetIDs []uuid.UUID) error {
	return a.store().deleteControlCabinetAssignments(ctx, controlCabinetIDs)
}

func (a projectAssignment) deleteSPSControllerAssignments(ctx context.Context, spsControllerIDs []uuid.UUID) error {
	return a.store().deleteSPSControllerAssignments(ctx, spsControllerIDs)
}

func (a projectAssignment) deleteFieldDeviceAssignments(ctx context.Context, fieldDeviceIDs []uuid.UUID) error {
	return a.store().deleteFieldDeviceAssignments(ctx, fieldDeviceIDs)
}
