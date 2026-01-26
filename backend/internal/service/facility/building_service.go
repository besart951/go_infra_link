package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type BuildingService struct {
	repo domainFacility.BuildingRepository
}

func NewBuildingService(repo domainFacility.BuildingRepository) *BuildingService {
	return &BuildingService{repo: repo}
}

func (s *BuildingService) Create(building *domainFacility.Building) error {
	if err := s.validateRequiredFields(building); err != nil {
		return err
	}
	return s.repo.Create(building)
}

func (s *BuildingService) GetByID(id uuid.UUID) (*domainFacility.Building, error) {
	buildings, err := s.repo.GetByIds([]uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	if len(buildings) == 0 {
		return nil, domain.ErrNotFound
	}
	return buildings[0], nil
}

func (s *BuildingService) List(page, limit int, search string) (*domain.PaginatedList[domainFacility.Building], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *BuildingService) Update(building *domainFacility.Building) error {
	if err := s.validateRequiredFields(building); err != nil {
		return err
	}
	return s.repo.Update(building)
}

func (s *BuildingService) DeleteByIds(ids []uuid.UUID) error {
	return s.repo.DeleteByIds(ids)
}

func (s *BuildingService) validateRequiredFields(building *domainFacility.Building) error {
	ve := domain.NewValidationError()
	if building.BuildingGroup == 0 {
		ve = ve.Add("building.building_group", "building_group is required")
	}
	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
}
