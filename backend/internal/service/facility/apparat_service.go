package facility

import (
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type ApparatService struct {
	repo domainFacility.ApparatRepository
}

func NewApparatService(repo domainFacility.ApparatRepository) *ApparatService {
	return &ApparatService{repo: repo}
}

func (s *ApparatService) Create(apparat *domainFacility.Apparat) error {
	if err := s.validateRequiredFields(apparat); err != nil {
		return err
	}
	if err := s.ensureUnique(apparat, nil); err != nil {
		return err
	}
	return s.repo.Create(apparat)
}

func (s *ApparatService) GetByID(id uuid.UUID) (*domainFacility.Apparat, error) {
	apparats, err := s.repo.GetByIds([]uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	if len(apparats) == 0 {
		return nil, domain.ErrNotFound
	}
	return apparats[0], nil
}

func (s *ApparatService) List(page, limit int, search string) (*domain.PaginatedList[domainFacility.Apparat], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *ApparatService) Update(apparat *domainFacility.Apparat) error {
	if err := s.validateRequiredFields(apparat); err != nil {
		return err
	}
	if err := s.ensureUnique(apparat, &apparat.ID); err != nil {
		return err
	}
	return s.repo.Update(apparat)
}

func (s *ApparatService) DeleteByIds(ids []uuid.UUID) error {
	return s.repo.DeleteByIds(ids)
}

func (s *ApparatService) validateRequiredFields(apparat *domainFacility.Apparat) error {
	ve := domain.NewValidationError()
	if strings.TrimSpace(apparat.ShortName) == "" {
		ve.Add("apparat.short_name", "short_name is required")
	}
	if strings.TrimSpace(apparat.Name) == "" {
		ve.Add("apparat.name", "name is required")
	}
	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
}

func (s *ApparatService) ensureUnique(apparat *domainFacility.Apparat, excludeID *uuid.UUID) error {
	ve := domain.NewValidationError()

	if strings.TrimSpace(apparat.ShortName) != "" {
		items, err := s.repo.GetPaginatedList(domain.PaginationParams{Page: 1, Limit: 1000, Search: apparat.ShortName})
		if err != nil {
			return err
		}
		for i := range items.Items {
			item := items.Items[i]
			if excludeID != nil && item.ID == *excludeID {
				continue
			}
			if strings.EqualFold(item.ShortName, apparat.ShortName) {
				ve.Add("apparat.short_name", "short_name must be unique")
				break
			}
		}
	}

	if strings.TrimSpace(apparat.Name) != "" {
		items, err := s.repo.GetPaginatedList(domain.PaginationParams{Page: 1, Limit: 1000, Search: apparat.Name})
		if err != nil {
			return err
		}
		for i := range items.Items {
			item := items.Items[i]
			if excludeID != nil && item.ID == *excludeID {
				continue
			}
			if strings.EqualFold(item.Name, apparat.Name) {
				ve.Add("apparat.name", "name must be unique")
				break
			}
		}
	}

	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
}
