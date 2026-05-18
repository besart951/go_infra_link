package facility

import (
	"context"
	"fmt"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type BacnetReferenceUsageService struct {
	repo domainFacility.BacnetReferenceUsageRepository
}

func NewBacnetReferenceUsageService(repo domainFacility.BacnetReferenceUsageRepository) *BacnetReferenceUsageService {
	return &BacnetReferenceUsageService{repo: repo}
}

func (s *BacnetReferenceUsageService) CountByResource(ctx context.Context, resource domainFacility.BacnetReferenceResource, ids []uuid.UUID) ([]domainFacility.BacnetReferenceUsage, error) {
	if s == nil || s.repo == nil {
		return []domainFacility.BacnetReferenceUsage{}, nil
	}

	uniqueIDs := uniqueUUIDs(ids)
	counts, err := s.repo.CountByResource(ctx, resource, uniqueIDs)
	if err != nil {
		return nil, err
	}

	items := make([]domainFacility.BacnetReferenceUsage, 0, len(uniqueIDs))
	for _, id := range uniqueIDs {
		items = append(items, domainFacility.BacnetReferenceUsage{
			Resource:          resource,
			ID:                id,
			BacnetObjectCount: counts[id],
		})
	}
	return items, nil
}

type bacnetReferenceDeleteGuard struct {
	resource domainFacility.BacnetReferenceResource
	usage    *BacnetReferenceUsageService
}

func newBacnetReferenceDeleteGuard(resource domainFacility.BacnetReferenceResource, repos ...domainFacility.BacnetReferenceUsageRepository) bacnetReferenceDeleteGuard {
	if len(repos) == 0 || repos[0] == nil {
		return bacnetReferenceDeleteGuard{resource: resource}
	}
	return bacnetReferenceDeleteGuard{
		resource: resource,
		usage:    NewBacnetReferenceUsageService(repos[0]),
	}
}

func (g bacnetReferenceDeleteGuard) ensureDeleteAllowed(ctx context.Context, id uuid.UUID) error {
	if g.usage == nil {
		return nil
	}
	if id == uuid.Nil {
		return domain.ErrInvalidArgument
	}

	items, err := g.usage.CountByResource(ctx, g.resource, []uuid.UUID{id})
	if err != nil {
		return err
	}
	if len(items) == 0 || items[0].BacnetObjectCount == 0 {
		return nil
	}

	return fmt.Errorf("%w: %s %s is used by %d bacnet objects",
		domainFacility.ErrBacnetReferenceInUse,
		g.resource,
		id,
		items[0].BacnetObjectCount,
	)
}
