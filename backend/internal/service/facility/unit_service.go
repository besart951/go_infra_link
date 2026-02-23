package facility

import (
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type UnitService struct {
	repo domainFacility.UnitRepository
}

func NewUnitService(repo domainFacility.UnitRepository) *UnitService {
	return &UnitService{repo: repo}
}

func (s *UnitService) Create(unit *domainFacility.Unit) error {
	ve := domain.NewValidationError()
	if strings.TrimSpace(unit.Code) == "" {
		ve = ve.Add("code", "required")
	}
	if strings.TrimSpace(unit.Symbol) == "" {
		ve = ve.Add("symbol", "required")
	}
	if strings.TrimSpace(unit.Name) == "" {
		ve = ve.Add("name", "required")
	}
	if len(ve.Fields) > 0 {
		return ve
	}
	return s.repo.Create(unit)
}

func (s *UnitService) List(page, limit int, search string) (*domain.PaginatedList[domainFacility.Unit], error) {
	page, limit = domain.NormalizePagination(page, limit, 20)
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *UnitService) GetByID(id uuid.UUID) (*domainFacility.Unit, error) {
	return domain.GetByID(s.repo, id)
}

func (s *UnitService) Update(unit *domainFacility.Unit) error {
	ve := domain.NewValidationError()
	if strings.TrimSpace(unit.Code) == "" {
		ve = ve.Add("code", "required")
	}
	if strings.TrimSpace(unit.Symbol) == "" {
		ve = ve.Add("symbol", "required")
	}
	if strings.TrimSpace(unit.Name) == "" {
		ve = ve.Add("name", "required")
	}
	if len(ve.Fields) > 0 {
		return ve
	}
	return s.repo.Update(unit)
}

func (s *UnitService) DeleteByID(id uuid.UUID) error {
	return s.repo.DeleteByIds([]uuid.UUID{id})
}
