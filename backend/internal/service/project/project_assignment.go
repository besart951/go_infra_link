package project

import (
	"context"
	"fmt"

	domainFieldDevice "github.com/besart951/go_infra_link/backend/internal/domain/facility/fielddevice"
	domainHierarchy "github.com/besart951/go_infra_link/backend/internal/domain/facility/hierarchy"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type projectAssignment struct {
	deps projectAssignmentDependencies
}

type projectAssignmentDependencies struct {
	projectRepo               domainProject.ProjectRepository
	projectControlCabinetRepo domainProject.ProjectControlCabinetRepository
	projectSPSControllerRepo  domainProject.ProjectSPSControllerRepository
	projectFieldDeviceRepo    domainProject.ProjectFieldDeviceRepository
	spsControllerRepo         domainFacility.SPSControllerRepository
	spsControllerSystemRepo   domainHierarchy.SPSControllerSystemTypeStore
	fieldDeviceRepo           domainFieldDevice.FieldDeviceStore
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
	return projectAssignment{deps: projectAssignmentDependencies{
		projectRepo:               s.projectRepo,
		projectControlCabinetRepo: s.projectControlCabinetRepo,
		projectSPSControllerRepo:  s.projectSPSControllerRepo,
		projectFieldDeviceRepo:    s.projectFieldDeviceRepo,
		spsControllerRepo:         s.spsControllerRepo,
		spsControllerSystemRepo:   s.spsControllerSystemRepo,
		fieldDeviceRepo:           s.fieldDeviceRepo,
	}}
}

func (a projectAssignment) store() projectAssignmentStore {
	return newProjectAssignmentStore(a.deps)
}

func (a projectAssignment) assign(ctx context.Context, projectID uuid.UUID, target projectAssignmentTarget) (*projectAssignmentResult, error) {
	switch target.kind {
	case projectAssignmentControlCabinet:
		entity := &domainProject.ProjectControlCabinet{ProjectID: projectID, ControlCabinetID: target.id}
		if err := a.deps.projectControlCabinetRepo.Create(ctx, entity); err != nil {
			return nil, err
		}
		if err := a.assignControlCabinetDescendants(ctx, projectID, target.id); err != nil {
			return nil, err
		}
		return &projectAssignmentResult{controlCabinet: entity}, nil
	case projectAssignmentSPSController:
		existing, err := a.findProjectSPSControllerLink(ctx, projectID, target.id)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			if writer, ok := a.deps.projectSPSControllerRepo.(interface {
				AddAssignmentSource(
					context.Context,
					uuid.UUID,
					[]uuid.UUID,
					domainProject.AssignmentSource,
				) error
			}); ok {
				if err := writer.AddAssignmentSource(
					ctx,
					projectID,
					[]uuid.UUID{target.id},
					domainProject.ExplicitAssignmentSource(),
				); err != nil {
					return nil, err
				}
			}
			systemTypeIDs, err := a.deps.spsControllerSystemRepo.
				GetIDsBySPSControllerIDs(ctx, []uuid.UUID{target.id})
			if err != nil {
				return nil, err
			}
			if err := a.assignFieldDevicesForSystemTypes(
				ctx,
				projectID,
				systemTypeIDs,
				domainProject.AssignmentSource{
					Kind:           domainProject.AssignmentSourceSPSController,
					SourceEntityID: target.id,
				},
			); err != nil {
				return nil, err
			}
			return &projectAssignmentResult{spsController: existing}, nil
		}
		entity := &domainProject.ProjectSPSController{ProjectID: projectID, SPSControllerID: target.id}
		if err := a.deps.projectSPSControllerRepo.Create(ctx, entity); err != nil {
			return nil, err
		}
		if err := a.assignSPSControllerDescendants(
			ctx,
			projectID,
			[]uuid.UUID{target.id},
			domainProject.ExplicitAssignmentSource(),
			domainProject.AssignmentSource{
				Kind:           domainProject.AssignmentSourceSPSController,
				SourceEntityID: target.id,
			},
		); err != nil {
			return nil, err
		}
		return &projectAssignmentResult{spsController: entity}, nil
	case projectAssignmentSPSControllerSystemType:
		if err := a.assignFieldDevicesForSystemTypes(
			ctx,
			projectID,
			[]uuid.UUID{target.id},
			domainProject.AssignmentSource{
				Kind:           domainProject.AssignmentSourceSPSSystemType,
				SourceEntityID: target.id,
			},
		); err != nil {
			return nil, err
		}
		return &projectAssignmentResult{}, nil
	case projectAssignmentFieldDevice:
		existing, err := a.findProjectFieldDeviceLink(ctx, projectID, target.id)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			if writer, ok := a.deps.projectFieldDeviceRepo.(interface {
				AddAssignmentSource(
					context.Context,
					uuid.UUID,
					[]uuid.UUID,
					domainProject.AssignmentSource,
				) error
			}); ok {
				if err := writer.AddAssignmentSource(
					ctx,
					projectID,
					[]uuid.UUID{target.id},
					domainProject.ExplicitAssignmentSource(),
				); err != nil {
					return nil, err
				}
			}
			return &projectAssignmentResult{fieldDevice: existing}, nil
		}
		entity := &domainProject.ProjectFieldDevice{ProjectID: projectID, FieldDeviceID: target.id}
		if err := a.deps.projectFieldDeviceRepo.Create(ctx, entity); err != nil {
			return nil, err
		}
		return &projectAssignmentResult{fieldDevice: entity}, nil
	default:
		return nil, domain.ErrInvalidArgument
	}
}

func (a projectAssignment) findProjectSPSControllerLink(
	ctx context.Context,
	projectID uuid.UUID,
	spsControllerID uuid.UUID,
) (*domainProject.ProjectSPSController, error) {
	links, err := a.deps.projectSPSControllerRepo.GetBySPSControllerID(
		ctx,
		spsControllerID,
	)
	if err != nil {
		return nil, err
	}
	for _, link := range links {
		if link != nil && link.ProjectID == projectID {
			return link, nil
		}
	}
	return nil, nil
}

func (a projectAssignment) findProjectFieldDeviceLink(
	ctx context.Context,
	projectID uuid.UUID,
	fieldDeviceID uuid.UUID,
) (*domainProject.ProjectFieldDevice, error) {
	links, err := a.deps.projectFieldDeviceRepo.GetByFieldDeviceIDs(
		ctx,
		[]uuid.UUID{fieldDeviceID},
	)
	if err != nil {
		return nil, err
	}
	for _, link := range links {
		if link != nil && link.ProjectID == projectID {
			return link, nil
		}
	}
	return nil, nil
}

func (a projectAssignment) update(ctx context.Context, linkID, projectID uuid.UUID, target projectAssignmentTarget) (*projectAssignmentResult, error) {
	switch target.kind {
	case projectAssignmentControlCabinet:
		entity, err := domain.GetByID(ctx, a.deps.projectControlCabinetRepo, linkID)
		if err != nil {
			return nil, err
		}
		if entity.ProjectID != projectID {
			return nil, domain.ErrNotFound
		}
		previousControlCabinetID := entity.ControlCabinetID
		if previousControlCabinetID != target.id {
			source := domainProject.AssignmentSource{
				Kind:           domainProject.AssignmentSourceControlCabinet,
				SourceEntityID: previousControlCabinetID,
			}
			if err := pruneProjectAssignmentSource(
				ctx,
				a.deps.projectFieldDeviceRepo,
				projectID,
				source,
			); err != nil {
				return nil, err
			}
			if err := pruneProjectAssignmentSource(
				ctx,
				a.deps.projectSPSControllerRepo,
				projectID,
				source,
			); err != nil {
				return nil, err
			}
		}
		entity.ControlCabinetID = target.id
		if err := a.deps.projectControlCabinetRepo.Update(ctx, entity); err != nil {
			return nil, err
		}
		if err := a.assignControlCabinetDescendants(ctx, projectID, target.id); err != nil {
			return nil, err
		}
		return &projectAssignmentResult{controlCabinet: entity}, nil
	case projectAssignmentSPSController:
		entity, err := domain.GetByID(ctx, a.deps.projectSPSControllerRepo, linkID)
		if err != nil {
			return nil, err
		}
		if entity.ProjectID != projectID {
			return nil, domain.ErrNotFound
		}
		previousSPSControllerID := entity.SPSControllerID
		if previousSPSControllerID != target.id {
			explicitSource := domainProject.AssignmentSource{
				Kind:           domainProject.AssignmentSourceExplicit,
				SourceEntityID: previousSPSControllerID,
			}
			if err := requireExclusiveProjectAssignmentSource(
				ctx,
				a.deps.projectSPSControllerRepo,
				linkID,
				explicitSource,
			); err != nil {
				return nil, err
			}
			if err := pruneProjectAssignmentSource(
				ctx,
				a.deps.projectFieldDeviceRepo,
				projectID,
				domainProject.AssignmentSource{
					Kind:           domainProject.AssignmentSourceSPSController,
					SourceEntityID: previousSPSControllerID,
				},
			); err != nil {
				return nil, err
			}
		}
		entity.SPSControllerID = target.id
		if err := a.deps.projectSPSControllerRepo.Update(ctx, entity); err != nil {
			return nil, err
		}
		if previousSPSControllerID != target.id {
			if err := replaceExplicitProjectAssignmentSource(
				ctx,
				a.deps.projectSPSControllerRepo,
				linkID,
				previousSPSControllerID,
				target.id,
			); err != nil {
				return nil, err
			}
		}
		if err := a.assignSPSControllerDescendants(
			ctx,
			projectID,
			[]uuid.UUID{target.id},
			domainProject.ExplicitAssignmentSource(),
			domainProject.AssignmentSource{
				Kind:           domainProject.AssignmentSourceSPSController,
				SourceEntityID: target.id,
			},
		); err != nil {
			return nil, err
		}
		return &projectAssignmentResult{spsController: entity}, nil
	case projectAssignmentFieldDevice:
		entity, err := domain.GetByID(ctx, a.deps.projectFieldDeviceRepo, linkID)
		if err != nil {
			return nil, err
		}
		if entity.ProjectID != projectID {
			return nil, domain.ErrNotFound
		}
		previousFieldDeviceID := entity.FieldDeviceID
		if previousFieldDeviceID != target.id {
			explicitSource := domainProject.AssignmentSource{
				Kind:           domainProject.AssignmentSourceExplicit,
				SourceEntityID: previousFieldDeviceID,
			}
			if err := requireExclusiveProjectAssignmentSource(
				ctx,
				a.deps.projectFieldDeviceRepo,
				linkID,
				explicitSource,
			); err != nil {
				return nil, err
			}
		}
		entity.FieldDeviceID = target.id
		if err := a.deps.projectFieldDeviceRepo.Update(ctx, entity); err != nil {
			return nil, err
		}
		if previousFieldDeviceID != target.id {
			if err := replaceExplicitProjectAssignmentSource(
				ctx,
				a.deps.projectFieldDeviceRepo,
				linkID,
				previousFieldDeviceID,
				target.id,
			); err != nil {
				return nil, err
			}
		}
		return &projectAssignmentResult{fieldDevice: entity}, nil
	default:
		return nil, domain.ErrInvalidArgument
	}
}

func pruneProjectAssignmentSource(
	ctx context.Context,
	repository any,
	projectID uuid.UUID,
	source domainProject.AssignmentSource,
) error {
	_, err := pruneProjectAssignmentSourceHandled(ctx, repository, projectID, source)
	return err
}

func pruneProjectAssignmentSourceHandled(
	ctx context.Context,
	repository any,
	projectID uuid.UUID,
	source domainProject.AssignmentSource,
) (bool, error) {
	pruner, ok := repository.(interface {
		RemoveAssignmentSourceAndPrune(
			context.Context,
			uuid.UUID,
			domainProject.AssignmentSource,
		) (bool, error)
	})
	if !ok {
		return false, nil
	}
	_, err := pruner.RemoveAssignmentSourceAndPrune(ctx, projectID, source)
	return true, err
}

func requireExclusiveProjectAssignmentSource(
	ctx context.Context,
	repository any,
	linkID uuid.UUID,
	explicitSource domainProject.AssignmentSource,
) error {
	reader, ok := repository.(interface {
		HasAssignmentSourceOtherThan(
			context.Context,
			uuid.UUID,
			domainProject.AssignmentSource,
		) (bool, error)
	})
	if !ok {
		return nil
	}
	hasOtherSource, err := reader.HasAssignmentSourceOtherThan(
		ctx,
		linkID,
		explicitSource,
	)
	if err != nil {
		return err
	}
	if hasOtherSource {
		return fmt.Errorf(
			"project association is also inherited and cannot be reassigned: %w",
			domain.ErrConflict,
		)
	}
	return nil
}

func replaceExplicitProjectAssignmentSource(
	ctx context.Context,
	repository any,
	linkID uuid.UUID,
	fromEntityID uuid.UUID,
	toEntityID uuid.UUID,
) error {
	writer, ok := repository.(interface {
		ReplaceExplicitAssignmentSource(
			context.Context,
			uuid.UUID,
			uuid.UUID,
			uuid.UUID,
		) error
	})
	if !ok {
		return nil
	}
	return writer.ReplaceExplicitAssignmentSource(
		ctx,
		linkID,
		fromEntityID,
		toEntityID,
	)
}

func (a projectAssignment) remove(ctx context.Context, linkID, projectID uuid.UUID, kind projectAssignmentKind) error {
	switch kind {
	case projectAssignmentControlCabinet:
		entity, err := domain.GetByID(ctx, a.deps.projectControlCabinetRepo, linkID)
		if err != nil {
			return err
		}
		if entity.ProjectID != projectID {
			return domain.ErrNotFound
		}
		source := domainProject.AssignmentSource{
			Kind:           domainProject.AssignmentSourceControlCabinet,
			SourceEntityID: entity.ControlCabinetID,
		}
		if err := pruneProjectAssignmentSource(
			ctx,
			a.deps.projectFieldDeviceRepo,
			projectID,
			source,
		); err != nil {
			return err
		}
		if err := pruneProjectAssignmentSource(
			ctx,
			a.deps.projectSPSControllerRepo,
			projectID,
			source,
		); err != nil {
			return err
		}
		return a.deps.projectControlCabinetRepo.DeleteByIds(ctx, []uuid.UUID{linkID})
	case projectAssignmentSPSController:
		entity, err := domain.GetByID(ctx, a.deps.projectSPSControllerRepo, linkID)
		if err != nil {
			return err
		}
		if entity.ProjectID != projectID {
			return domain.ErrNotFound
		}
		if err := pruneProjectAssignmentSource(
			ctx,
			a.deps.projectFieldDeviceRepo,
			projectID,
			domainProject.AssignmentSource{
				Kind:           domainProject.AssignmentSourceSPSController,
				SourceEntityID: entity.SPSControllerID,
			},
		); err != nil {
			return err
		}
		handled, err := pruneProjectAssignmentSourceHandled(
			ctx,
			a.deps.projectSPSControllerRepo,
			projectID,
			domainProject.AssignmentSource{
				Kind:           domainProject.AssignmentSourceExplicit,
				SourceEntityID: entity.SPSControllerID,
			},
		)
		if err != nil || handled {
			return err
		}
		return a.deps.projectSPSControllerRepo.DeleteByIds(ctx, []uuid.UUID{linkID})
	case projectAssignmentFieldDevice:
		entity, err := domain.GetByID(ctx, a.deps.projectFieldDeviceRepo, linkID)
		if err != nil {
			return err
		}
		if entity.ProjectID != projectID {
			return domain.ErrNotFound
		}
		handled, err := pruneProjectAssignmentSourceHandled(
			ctx,
			a.deps.projectFieldDeviceRepo,
			projectID,
			domainProject.AssignmentSource{
				Kind:           domainProject.AssignmentSourceExplicit,
				SourceEntityID: entity.FieldDeviceID,
			},
		)
		if err != nil || handled {
			return err
		}
		return a.deps.projectFieldDeviceRepo.DeleteByIds(ctx, []uuid.UUID{linkID})
	default:
		return domain.ErrInvalidArgument
	}
}

func (a projectAssignment) multiAssignFieldDevices(ctx context.Context, projectID uuid.UUID, fieldDeviceIDs []uuid.UUID) ([]uuid.UUID, []string) {
	if _, err := domain.GetByID(ctx, a.deps.projectRepo, projectID); err != nil {
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
	return a.store().assignFieldDeviceIDs(
		ctx,
		projectID,
		fieldDeviceIDs,
		domainProject.ExplicitAssignmentSource(),
	)
}

func (a projectAssignment) assignControlCabinetDescendants(ctx context.Context, projectID, controlCabinetID uuid.UUID) error {
	spsControllerIDs, err := a.deps.spsControllerRepo.GetIDsByControlCabinetID(ctx, controlCabinetID)
	if err != nil {
		return err
	}
	source := domainProject.AssignmentSource{
		Kind:           domainProject.AssignmentSourceControlCabinet,
		SourceEntityID: controlCabinetID,
	}
	return a.assignSPSControllerDescendants(
		ctx,
		projectID,
		spsControllerIDs,
		source,
		source,
	)
}

func (a projectAssignment) assignSPSControllerDescendants(
	ctx context.Context,
	projectID uuid.UUID,
	spsControllerIDs []uuid.UUID,
	spsSource domainProject.AssignmentSource,
	fieldDeviceSource domainProject.AssignmentSource,
) error {
	return a.store().assignSPSControllerDescendants(
		ctx,
		projectID,
		spsControllerIDs,
		spsSource,
		fieldDeviceSource,
	)
}

func (a projectAssignment) assignFieldDevicesForSystemTypes(
	ctx context.Context,
	projectID uuid.UUID,
	systemTypeIDs []uuid.UUID,
	source domainProject.AssignmentSource,
) error {
	return a.store().assignFieldDevicesForSystemTypes(
		ctx,
		projectID,
		systemTypeIDs,
		source,
	)
}
