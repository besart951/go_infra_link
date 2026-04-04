package facility

import (
	"context"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type SystemPartService struct {
	baseService[domainFacility.SystemPart]
	extRepo domainFacility.SystemPartRepository
}

func NewSystemPartService(repo domainFacility.SystemPartRepository) *SystemPartService {
	return &SystemPartService{
		baseService: newBase[domainFacility.SystemPart](repo, 10),
		extRepo:     repo,
	}
}

func (s *SystemPartService) Create(ctx context.Context, systemPart *domainFacility.SystemPart) error {
	if err := s.validateRequiredFields(systemPart); err != nil {
		return err
	}
	if err := s.ensureUnique(ctx, systemPart, nil); err != nil {
		return err
	}
	return s.repo.Create(ctx, systemPart)
}

func (s *SystemPartService) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*domainFacility.SystemPart, error) {
	return s.extRepo.GetByIds(ctx, ids)
}

func (s *SystemPartService) GetApparatIDs(ctx context.Context, id uuid.UUID) ([]uuid.UUID, error) {
	systemPart, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return extractIDs(systemPart.Apparats, func(a *domainFacility.Apparat) uuid.UUID { return a.ID }), nil
}

func (s *SystemPartService) Update(ctx context.Context, systemPart *domainFacility.SystemPart) error {
	if err := s.validateRequiredFields(systemPart); err != nil {
		return err
	}
	if err := s.ensureUnique(ctx, systemPart, &systemPart.ID); err != nil {
		return err
	}
	return s.repo.Update(ctx, systemPart)
}

func (s *SystemPartService) validateRequiredFields(systemPart *domainFacility.SystemPart) error {
	ve := domain.NewValidationError()
	shortName := strings.TrimSpace(systemPart.ShortName)
	if shortName == "" {
		ve = ve.Add("system_part.short_name", "short_name is required")
	} else if len(shortName) != 3 {
		ve = ve.Add("system_part.short_name", "short_name must be exactly 3 characters")
	}
	if strings.TrimSpace(systemPart.Name) == "" {
		ve = ve.Add("system_part.name", "name is required")
	}
	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
}

func (s *SystemPartService) ensureUnique(ctx context.Context, systemPart *domainFacility.SystemPart, excludeID *uuid.UUID) error {
	ve := domain.NewValidationError()
	if strings.TrimSpace(systemPart.ShortName) != "" {
		items, err := s.repo.GetPaginatedList(ctx, domain.PaginationParams{Page: 1, Limit: 1000, Search: systemPart.ShortName})
		if err != nil {
			return err
		}
		for i := range items.Items {
			item := items.Items[i]
			if excludeID != nil && item.ID == *excludeID {
				continue
			}
			if strings.EqualFold(item.ShortName, systemPart.ShortName) {
				ve = ve.Add("system_part.short_name", "short_name must be unique")
				break
			}
		}
	}
	if strings.TrimSpace(systemPart.Name) != "" {
		items, err := s.repo.GetPaginatedList(ctx, domain.PaginationParams{Page: 1, Limit: 1000, Search: systemPart.Name})
		if err != nil {
			return err
		}
		for i := range items.Items {
			item := items.Items[i]
			if excludeID != nil && item.ID == *excludeID {
				continue
			}
			if strings.EqualFold(item.Name, systemPart.Name) {
				ve = ve.Add("system_part.name", "name must be unique")
				break
			}
		}
	}
	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
}
