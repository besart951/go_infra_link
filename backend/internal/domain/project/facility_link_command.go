package project

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

// FacilityLinkUpdate identifies one version-guarded project link mutation.
type FacilityLinkUpdate struct {
	LinkID      uuid.UUID
	ProjectID   uuid.UUID
	TargetID    uuid.UUID
	BaseVersion domain.AggregateVersion
}

func (command FacilityLinkUpdate) Validate() error {
	if command.LinkID == uuid.Nil || command.ProjectID == uuid.Nil || command.TargetID == uuid.Nil || command.BaseVersion == 0 {
		return domain.ErrInvalidArgument
	}
	return nil
}

// FacilityLinkDelete identifies one version-guarded project link deletion.
type FacilityLinkDelete struct {
	LinkID      uuid.UUID
	ProjectID   uuid.UUID
	BaseVersion domain.AggregateVersion
}

func (command FacilityLinkDelete) Validate() error {
	if command.LinkID == uuid.Nil || command.ProjectID == uuid.Nil || command.BaseVersion == 0 {
		return domain.ErrInvalidArgument
	}
	return nil
}
