package facility

import (
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type BuildingService struct {
	repo                        domainFacility.BuildingRepository
	controlCabinetRepo          domainFacility.ControlCabinetRepository
	spsControllerRepo           domainFacility.SPSControllerRepository
	spsControllerSystemTypeRepo domainFacility.SPSControllerSystemTypeStore
	fieldDeviceRepo             domainFacility.FieldDeviceStore
	specificationRepo           domainFacility.SpecificationStore
	bacnetObjectRepo            domainFacility.BacnetObjectStore
}

func NewBuildingService(
	repo domainFacility.BuildingRepository,
	controlCabinetRepo domainFacility.ControlCabinetRepository,
	spsControllerRepo domainFacility.SPSControllerRepository,
	spsControllerSystemTypeRepo domainFacility.SPSControllerSystemTypeStore,
	fieldDeviceRepo domainFacility.FieldDeviceStore,
	specificationRepo domainFacility.SpecificationStore,
	bacnetObjectRepo domainFacility.BacnetObjectStore,
) *BuildingService {
	return &BuildingService{
		repo:                        repo,
		controlCabinetRepo:          controlCabinetRepo,
		spsControllerRepo:           spsControllerRepo,
		spsControllerSystemTypeRepo: spsControllerSystemTypeRepo,
		fieldDeviceRepo:             fieldDeviceRepo,
		specificationRepo:           specificationRepo,
		bacnetObjectRepo:            bacnetObjectRepo,
	}
}

func (s *BuildingService) Create(building *domainFacility.Building) error {
	if err := s.Validate(building, nil); err != nil {
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

func (s *BuildingService) GetByIDs(ids []uuid.UUID) ([]domainFacility.Building, error) {
	buildings, err := s.repo.GetByIds(ids)
	if err != nil {
		return nil, err
	}
	items := make([]domainFacility.Building, len(buildings))
	for i, item := range buildings {
		items[i] = *item
	}
	return items, nil
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
	if err := s.Validate(building, &building.ID); err != nil {
		return err
	}
	return s.repo.Update(building)
}

func (s *BuildingService) Validate(building *domainFacility.Building, excludeID *uuid.UUID) error {
	if err := s.validateRequiredFields(building); err != nil {
		return err
	}
	if err := s.ensureUnique(building, excludeID); err != nil {
		return err
	}
	return nil
}

func (s *BuildingService) DeleteByID(id uuid.UUID) error {
	// Cascade delete: Building → ControlCabinets → SPSControllers → SPSControllerSystemTypes → FieldDevices
	controlCabinetIDs, err := s.controlCabinetRepo.GetIDsByBuildingID(id)
	if err != nil {
		return err
	}

	if len(controlCabinetIDs) == 0 {
		// No children, delete directly
		return s.repo.DeleteByIds([]uuid.UUID{id})
	}

	// Collect all dependent IDs
	spsControllerIDs, err := s.spsControllerRepo.GetIDsByControlCabinetIDs(controlCabinetIDs)
	if err != nil {
		return err
	}

	spsControllerSystemTypeIDs, err := s.spsControllerSystemTypeRepo.GetIDsBySPSControllerIDs(spsControllerIDs)
	if err != nil {
		return err
	}

	fieldDeviceIDs, err := s.fieldDeviceRepo.GetIDsBySPSControllerSystemTypeIDs(spsControllerSystemTypeIDs)
	if err != nil {
		return err
	}

	// Delete in correct order (bottom-up)
	if err := s.bacnetObjectRepo.SoftDeleteByFieldDeviceIDs(fieldDeviceIDs); err != nil {
		return err
	}
	if err := s.specificationRepo.SoftDeleteByFieldDeviceIDs(fieldDeviceIDs); err != nil {
		return err
	}
	if err := s.fieldDeviceRepo.DeleteByIds(fieldDeviceIDs); err != nil {
		return err
	}

	if err := s.spsControllerSystemTypeRepo.SoftDeleteBySPSControllerIDs(spsControllerIDs); err != nil {
		return err
	}

	if err := s.spsControllerRepo.DeleteByIds(spsControllerIDs); err != nil {
		return err
	}

	if err := s.controlCabinetRepo.DeleteByIds(controlCabinetIDs); err != nil {
		return err
	}

	return s.repo.DeleteByIds([]uuid.UUID{id})
}

func (s *BuildingService) validateRequiredFields(building *domainFacility.Building) error {
	ve := domain.NewValidationError()
	if strings.TrimSpace(building.IWSCode) == "" {
		ve = ve.Add("building.iws_code", "iws_code is required")
	} else if len(strings.TrimSpace(building.IWSCode)) != 4 {
		ve = ve.Add("building.iws_code", "iws_code must be exactly 4 characters")
	}
	if building.BuildingGroup == 0 {
		ve = ve.Add("building.building_group", "building_group is required")
	}
	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
}

func (s *BuildingService) ensureUnique(building *domainFacility.Building, excludeID *uuid.UUID) error {
	if strings.TrimSpace(building.IWSCode) == "" || building.BuildingGroup == 0 {
		return nil
	}
	exists, err := s.repo.ExistsIWSCodeGroup(building.IWSCode, building.BuildingGroup, excludeID)
	if err != nil {
		return err
	}
	if exists {
		return domain.NewValidationError().Add("building.iws_code", "iws_code must be unique within the building group")
	}
	return nil
}
