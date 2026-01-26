package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type ControlCabinetService struct {
	repo         domainFacility.ControlCabinetRepository
	buildingRepo domainFacility.BuildingRepository
}

func NewControlCabinetService(repo domainFacility.ControlCabinetRepository, buildingRepo domainFacility.BuildingRepository) *ControlCabinetService {
	return &ControlCabinetService{repo: repo, buildingRepo: buildingRepo}
}

func (s *ControlCabinetService) Create(controlCabinet *domainFacility.ControlCabinet) error {
	if err := s.validateRequiredFields(controlCabinet); err != nil {
		return err
	}
	if err := s.ensureBuildingExists(controlCabinet.BuildingID); err != nil {
		return err
	}
	return s.repo.Create(controlCabinet)
}

func (s *ControlCabinetService) GetByID(id uuid.UUID) (*domainFacility.ControlCabinet, error) {
	controlCabinets, err := s.repo.GetByIds([]uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	if len(controlCabinets) == 0 {
		return nil, domain.ErrNotFound
	}
	return controlCabinets[0], nil
}

func (s *ControlCabinetService) List(page, limit int, search string) (*domain.PaginatedList[domainFacility.ControlCabinet], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *ControlCabinetService) Update(controlCabinet *domainFacility.ControlCabinet) error {
	if err := s.validateRequiredFields(controlCabinet); err != nil {
		return err
	}
	if err := s.ensureBuildingExists(controlCabinet.BuildingID); err != nil {
		return err
	}
	return s.repo.Update(controlCabinet)
}

func (s *ControlCabinetService) DeleteByIds(ids []uuid.UUID) error {
	return s.repo.DeleteByIds(ids)
}

func (s *ControlCabinetService) ensureBuildingExists(buildingID uuid.UUID) error {
	buildings, err := s.buildingRepo.GetByIds([]uuid.UUID{buildingID})
	if err != nil {
		return err
	}
	if len(buildings) == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *ControlCabinetService) validateRequiredFields(controlCabinet *domainFacility.ControlCabinet) error {
	ve := domain.NewValidationError()
	if controlCabinet.BuildingID == uuid.Nil {
		ve = ve.Add("controlcabinet.building_id", "building_id is required")
	}
	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
}
