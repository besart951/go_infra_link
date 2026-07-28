package project

import (
	"context"
	"sort"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type fieldDeviceHierarchyPlacement struct {
	spsControllerID  uuid.UUID
	controlCabinetID uuid.UUID
}

// ReconcileFieldDeviceMove updates only live inherited project-assignment
// claims. Explicit links, copied-system-type claims, and conservative legacy
// backfill claims survive the move.
func (s *ProjectFacilityLinkService) ReconcileFieldDeviceMove(
	ctx context.Context,
	fieldDeviceID uuid.UUID,
	fromSystemTypeID uuid.UUID,
	toSystemTypeID uuid.UUID,
) ([]uuid.UUID, error) {
	if fieldDeviceID == uuid.Nil ||
		fromSystemTypeID == uuid.Nil ||
		toSystemTypeID == uuid.Nil {
		return nil, domain.ErrInvalidArgument
	}
	if fromSystemTypeID == toSystemTypeID {
		return nil, nil
	}

	from, err := s.loadFieldDeviceHierarchyPlacement(ctx, fromSystemTypeID)
	if err != nil {
		return nil, err
	}
	to, err := s.loadFieldDeviceHierarchyPlacement(ctx, toSystemTypeID)
	if err != nil {
		return nil, err
	}

	affected := make(map[uuid.UUID]struct{})
	assignments := s.assignments()
	if from.spsControllerID != to.spsControllerID {
		fromProjects, err := s.ListProjectIDsBySPSControllerID(
			ctx,
			from.spsControllerID,
		)
		if err != nil {
			return nil, err
		}
		toProjects, err := s.ListProjectIDsBySPSControllerID(
			ctx,
			to.spsControllerID,
		)
		if err != nil {
			return nil, err
		}
		toExplicitProjects, err := s.listProjectIDsByExplicitSPSController(
			ctx,
			to.spsControllerID,
		)
		if err != nil {
			return nil, err
		}
		addProjectIDs(affected, fromProjects)
		addProjectIDs(affected, toProjects)

		toSource := domainProject.AssignmentSource{
			Kind:           domainProject.AssignmentSourceSPSController,
			SourceEntityID: to.spsControllerID,
		}
		for _, projectID := range toExplicitProjects {
			if err := assignments.store().assignFieldDeviceIDs(
				ctx,
				projectID,
				[]uuid.UUID{fieldDeviceID},
				toSource,
			); err != nil {
				return nil, err
			}
		}
		fromSource := domainProject.AssignmentSource{
			Kind:           domainProject.AssignmentSourceSPSController,
			SourceEntityID: from.spsControllerID,
		}
		for _, projectID := range fromProjects {
			if err := pruneProjectAssignmentSource(
				ctx,
				s.projectFieldDeviceRepo,
				projectID,
				fromSource,
			); err != nil {
				return nil, err
			}
		}
	}

	if from.controlCabinetID != to.controlCabinetID {
		fromProjects, err := s.ListProjectIDsByControlCabinetID(
			ctx,
			from.controlCabinetID,
		)
		if err != nil {
			return nil, err
		}
		toProjects, err := s.ListProjectIDsByControlCabinetID(
			ctx,
			to.controlCabinetID,
		)
		if err != nil {
			return nil, err
		}
		addProjectIDs(affected, fromProjects)
		addProjectIDs(affected, toProjects)

		toSource := domainProject.AssignmentSource{
			Kind:           domainProject.AssignmentSourceControlCabinet,
			SourceEntityID: to.controlCabinetID,
		}
		for _, projectID := range toProjects {
			if err := assignments.store().assignFieldDeviceIDs(
				ctx,
				projectID,
				[]uuid.UUID{fieldDeviceID},
				toSource,
			); err != nil {
				return nil, err
			}
		}
		fromSource := domainProject.AssignmentSource{
			Kind:           domainProject.AssignmentSourceControlCabinet,
			SourceEntityID: from.controlCabinetID,
		}
		for _, projectID := range fromProjects {
			if err := pruneProjectAssignmentSource(
				ctx,
				s.projectFieldDeviceRepo,
				projectID,
				fromSource,
			); err != nil {
				return nil, err
			}
		}
	}

	return sortedProjectIDSet(affected), nil
}

func (s *ProjectFacilityLinkService) listProjectIDsByExplicitSPSController(
	ctx context.Context,
	spsControllerID uuid.UUID,
) ([]uuid.UUID, error) {
	reader, ok := s.projectSPSControllerRepo.(interface {
		ListProjectIDsByAssignmentSource(
			context.Context,
			domainProject.AssignmentSource,
		) ([]uuid.UUID, error)
	})
	if !ok {
		return s.ListProjectIDsBySPSControllerID(ctx, spsControllerID)
	}
	return reader.ListProjectIDsByAssignmentSource(
		ctx,
		domainProject.AssignmentSource{
			Kind:           domainProject.AssignmentSourceExplicit,
			SourceEntityID: spsControllerID,
		},
	)
}

// ReconcileSPSControllerMove transfers cabinet-inherited claims to the new
// cabinet projects. Direct SPS and FieldDevice links remain untouched.
func (s *ProjectFacilityLinkService) ReconcileSPSControllerMove(
	ctx context.Context,
	spsControllerID uuid.UUID,
	fromControlCabinetID uuid.UUID,
	toControlCabinetID uuid.UUID,
) ([]uuid.UUID, error) {
	if spsControllerID == uuid.Nil ||
		fromControlCabinetID == uuid.Nil ||
		toControlCabinetID == uuid.Nil {
		return nil, domain.ErrInvalidArgument
	}
	if fromControlCabinetID == toControlCabinetID {
		return nil, nil
	}

	fromProjects, err := s.ListProjectIDsByControlCabinetID(
		ctx,
		fromControlCabinetID,
	)
	if err != nil {
		return nil, err
	}
	toProjects, err := s.ListProjectIDsByControlCabinetID(
		ctx,
		toControlCabinetID,
	)
	if err != nil {
		return nil, err
	}

	assignments := s.assignments()
	toSource := domainProject.AssignmentSource{
		Kind:           domainProject.AssignmentSourceControlCabinet,
		SourceEntityID: toControlCabinetID,
	}
	for _, projectID := range toProjects {
		if err := assignments.assignSPSControllerDescendants(
			ctx,
			projectID,
			[]uuid.UUID{spsControllerID},
			toSource,
			toSource,
		); err != nil {
			return nil, err
		}
	}

	fromSource := domainProject.AssignmentSource{
		Kind:           domainProject.AssignmentSourceControlCabinet,
		SourceEntityID: fromControlCabinetID,
	}
	for _, projectID := range fromProjects {
		if err := pruneProjectAssignmentSource(
			ctx,
			s.projectFieldDeviceRepo,
			projectID,
			fromSource,
		); err != nil {
			return nil, err
		}
		if err := pruneProjectAssignmentSource(
			ctx,
			s.projectSPSControllerRepo,
			projectID,
			fromSource,
		); err != nil {
			return nil, err
		}
	}

	affected := make(map[uuid.UUID]struct{}, len(fromProjects)+len(toProjects))
	addProjectIDs(affected, fromProjects)
	addProjectIDs(affected, toProjects)
	return sortedProjectIDSet(affected), nil
}

func (s *ProjectFacilityLinkService) loadFieldDeviceHierarchyPlacement(
	ctx context.Context,
	systemTypeID uuid.UUID,
) (fieldDeviceHierarchyPlacement, error) {
	systemType, err := domain.GetByID(ctx, s.spsControllerSystemRepo, systemTypeID)
	if err != nil {
		return fieldDeviceHierarchyPlacement{}, err
	}
	controller, err := domain.GetByID(
		ctx,
		s.spsControllerRepo,
		systemType.SPSControllerID,
	)
	if err != nil {
		return fieldDeviceHierarchyPlacement{}, err
	}
	return fieldDeviceHierarchyPlacement{
		spsControllerID:  controller.ID,
		controlCabinetID: controller.ControlCabinetID,
	}, nil
}

func addProjectIDs(target map[uuid.UUID]struct{}, projectIDs []uuid.UUID) {
	for _, projectID := range projectIDs {
		if projectID != uuid.Nil {
			target[projectID] = struct{}{}
		}
	}
}

func sortedProjectIDSet(projectIDs map[uuid.UUID]struct{}) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(projectIDs))
	for projectID := range projectIDs {
		out = append(out, projectID)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].String() < out[j].String()
	})
	return out
}
