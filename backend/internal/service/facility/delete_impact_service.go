package facility

import (
	"context"
	"fmt"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type DeleteImpactService struct {
	repo domainFacility.DeleteImpactRepository
}

func NewDeleteImpactService(repo domainFacility.DeleteImpactRepository) *DeleteImpactService {
	return &DeleteImpactService{repo: repo}
}

func (s *DeleteImpactService) List(ctx context.Context, resource domainFacility.DeleteImpactResource, ids []uuid.UUID) ([]domainFacility.DeleteImpact, error) {
	if s == nil || s.repo == nil {
		return []domainFacility.DeleteImpact{}, nil
	}
	return s.repo.DeleteImpacts(ctx, resource, uniqueUUIDs(ids))
}

func (s *DeleteImpactService) EnsureDeleteAllowed(ctx context.Context, resource domainFacility.DeleteImpactResource, id uuid.UUID) error {
	if id == uuid.Nil {
		return domainFacility.ErrReferenceInUse
	}
	impacts, err := s.List(ctx, resource, []uuid.UUID{id})
	if err != nil || len(impacts) == 0 || len(impacts[0].Blockers) == 0 {
		return err
	}
	return fmt.Errorf("%w: %s %s has %d blocking reference kinds", domainFacility.ErrReferenceInUse, resource, id, len(impacts[0].Blockers))
}
