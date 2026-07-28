package project

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

// AssignmentSourceKind describes why a materialized project hierarchy link
// exists. A link may have several sources at the same time.
type AssignmentSourceKind string

const (
	AssignmentSourceExplicit       AssignmentSourceKind = "explicit"
	AssignmentSourceControlCabinet AssignmentSourceKind = "control_cabinet"
	AssignmentSourceSPSController  AssignmentSourceKind = "sps_controller"
	AssignmentSourceSPSSystemType  AssignmentSourceKind = "sps_controller_system_type"
)

// AssignmentSource is persisted separately from the materialized project link.
// Explicit sources use the linked entity itself as SourceEntityID; callers may
// leave it nil when applying one source to a set of explicitly linked rows.
type AssignmentSource struct {
	Kind           AssignmentSourceKind
	SourceEntityID uuid.UUID
}

func ExplicitAssignmentSource() AssignmentSource {
	return AssignmentSource{Kind: AssignmentSourceExplicit}
}

func InheritedAssignmentSource(
	kind AssignmentSourceKind,
	sourceEntityID uuid.UUID,
) (AssignmentSource, error) {
	source := AssignmentSource{Kind: kind, SourceEntityID: sourceEntityID}
	if err := source.Validate(); err != nil {
		return AssignmentSource{}, err
	}
	return source, nil
}

func (source AssignmentSource) Validate() error {
	switch source.Kind {
	case AssignmentSourceExplicit:
		return nil
	case AssignmentSourceControlCabinet,
		AssignmentSourceSPSController,
		AssignmentSourceSPSSystemType:
		if source.SourceEntityID == uuid.Nil {
			return domain.ErrInvalidArgument
		}
		return nil
	default:
		return domain.ErrInvalidArgument
	}
}
