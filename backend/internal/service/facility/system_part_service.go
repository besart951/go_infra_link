package facility

import (
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type SystemPartService struct {
	repo domainFacility.SystemPartRepository
}

func NewSystemPartService(repo domainFacility.SystemPartRepository) *SystemPartService {
	return &SystemPartService{repo: repo}
}

func (s *SystemPartService) Create(systemPart *domainFacility.SystemPart) error {
	if err := s.validateRequiredFields(systemPart); err != nil {
		return err
	}
	if err := s.ensureUnique(systemPart, nil); err != nil {
		return err
	}
	return s.repo.Create(systemPart)
}

func (s *SystemPartService) GetByID(id uuid.UUID) (*domainFacility.SystemPart, error) {
	systemParts, err := s.repo.GetByIds([]uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	if len(systemParts) == 0 {
		return nil, domain.ErrNotFound
	}
	return systemParts[0], nil
}

func (s *SystemPartService) GetByIDs(ids []uuid.UUID) ([]*domainFacility.SystemPart, error) {
	return s.repo.GetByIds(ids)
}

func (s *SystemPartService) GetApparatIDs(id uuid.UUID) ([]uuid.UUID, error) {
	systemPart, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	apparatIDs := make([]uuid.UUID, 0, len(systemPart.Apparats))
	for _, apparat := range systemPart.Apparats {
		if apparat != nil {
			apparatIDs = append(apparatIDs, apparat.ID)
		}
	}
	return apparatIDs, nil
}

func (s *SystemPartService) List(page, limit int, search string) (*domain.PaginatedList[domainFacility.SystemPart], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *SystemPartService) Update(systemPart *domainFacility.SystemPart) error {
	if err := s.validateRequiredFields(systemPart); err != nil {
		return err
	}
	if err := s.ensureUnique(systemPart, &systemPart.ID); err != nil {
		return err
	}
	return s.repo.Update(systemPart)
}

func (s *SystemPartService) DeleteByID(id uuid.UUID) error {
	return s.repo.DeleteByIds([]uuid.UUID{id})
}

func (s *SystemPartService) validateRequiredFields(systemPart *domainFacility.SystemPart) error {
	ve := domain.NewValidationError()
	if strings.TrimSpace(systemPart.ShortName) == "" {
		ve = ve.Add("system_part.short_name", "short_name is required")
	}
	if strings.TrimSpace(systemPart.Name) == "" {
		ve = ve.Add("system_part.name", "name is required")
	}
	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
}

func (s *SystemPartService) ensureUnique(systemPart *domainFacility.SystemPart, excludeID *uuid.UUID) error {
	ve := domain.NewValidationError()

	if strings.TrimSpace(systemPart.ShortName) != "" {
		items, err := s.repo.GetPaginatedList(domain.PaginationParams{Page: 1, Limit: 1000, Search: systemPart.ShortName})
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
		items, err := s.repo.GetPaginatedList(domain.PaginationParams{Page: 1, Limit: 1000, Search: systemPart.Name})
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
