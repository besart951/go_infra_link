package facility

import (
	"context"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type ApparatService struct {
	baseService[domainFacility.Apparat]
	extRepo domainFacility.ApparatRepository
}

func NewApparatService(repo domainFacility.ApparatRepository) *ApparatService {
	return &ApparatService{
		baseService: newBase[domainFacility.Apparat](repo, 10),
		extRepo:     repo,
	}
}

func (s *ApparatService) Create(ctx context.Context, apparat *domainFacility.Apparat) error {
	if err := s.validateRequiredFields(apparat); err != nil {
		return err
	}
	if err := s.ensureUnique(ctx, apparat, nil); err != nil {
		return err
	}
	return s.repo.Create(ctx, apparat)
}

func (s *ApparatService) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*domainFacility.Apparat, error) {
	return s.extRepo.GetByIds(ctx, ids)
}

func (s *ApparatService) Update(ctx context.Context, apparat *domainFacility.Apparat) error {
	if err := s.validateRequiredFields(apparat); err != nil {
		return err
	}
	if err := s.ensureUnique(ctx, apparat, &apparat.ID); err != nil {
		return err
	}
	return s.repo.Update(ctx, apparat)
}

func (s *ApparatService) GetSystemPartIDs(ctx context.Context, id uuid.UUID) ([]uuid.UUID, error) {
	apparat, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return extractIDs(apparat.SystemParts, func(sp *domainFacility.SystemPart) uuid.UUID { return sp.ID }), nil
}

func (s *ApparatService) validateRequiredFields(apparat *domainFacility.Apparat) error {
	ve := domain.NewValidationError()
	shortName := strings.TrimSpace(apparat.ShortName)
	if shortName == "" {
		ve = ve.Add("apparat.short_name", "short_name is required")
	} else if len(shortName) != 3 {
		ve = ve.Add("apparat.short_name", "short_name must be exactly 3 characters")
	}
	if strings.TrimSpace(apparat.Name) == "" {
		ve = ve.Add("apparat.name", "name is required")
	}
	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
}

func (s *ApparatService) ensureUnique(ctx context.Context, apparat *domainFacility.Apparat, excludeID *uuid.UUID) error {
	ve := domain.NewValidationError()
	if strings.TrimSpace(apparat.ShortName) != "" {
		items, err := s.repo.GetPaginatedList(ctx, domain.PaginationParams{Page: 1, Limit: 1000, Search: apparat.ShortName})
		if err != nil {
			return err
		}
		for i := range items.Items {
			item := items.Items[i]
			if excludeID != nil && item.ID == *excludeID {
				continue
			}
			if strings.EqualFold(item.ShortName, apparat.ShortName) {
				ve = ve.Add("apparat.short_name", "short_name must be unique")
				break
			}
		}
	}
	if strings.TrimSpace(apparat.Name) != "" {
		items, err := s.repo.GetPaginatedList(ctx, domain.PaginationParams{Page: 1, Limit: 1000, Search: apparat.Name})
		if err != nil {
			return err
		}
		for i := range items.Items {
			item := items.Items[i]
			if excludeID != nil && item.ID == *excludeID {
				continue
			}
			if strings.EqualFold(item.Name, apparat.Name) {
				ve = ve.Add("apparat.name", "name must be unique")
				break
			}
		}
	}
	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
}
