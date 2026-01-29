package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type ControlCabinetService struct {
	repo                    domainFacility.ControlCabinetRepository
	buildingRepo            domainFacility.BuildingRepository
	spsControllerRepo       domainFacility.SPSControllerRepository
	spsControllerSystemRepo domainFacility.SPSControllerSystemTypeStore
	fieldDeviceRepo         domainFacility.FieldDeviceStore
	bacnetObjectRepo        domainFacility.BacnetObjectStore
	specificationRepo       domainFacility.SpecificationStore
}

func NewControlCabinetService(
	repo domainFacility.ControlCabinetRepository,
	buildingRepo domainFacility.BuildingRepository,
	spsControllerRepo domainFacility.SPSControllerRepository,
	spsControllerSystemRepo domainFacility.SPSControllerSystemTypeStore,
	fieldDeviceRepo domainFacility.FieldDeviceStore,
	bacnetObjectRepo domainFacility.BacnetObjectStore,
	specificationRepo domainFacility.SpecificationStore,
) *ControlCabinetService {
	return &ControlCabinetService{
		repo:                    repo,
		buildingRepo:            buildingRepo,
		spsControllerRepo:       spsControllerRepo,
		spsControllerSystemRepo: spsControllerSystemRepo,
		fieldDeviceRepo:         fieldDeviceRepo,
		bacnetObjectRepo:        bacnetObjectRepo,
		specificationRepo:       specificationRepo,
	}
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

func (s *ControlCabinetService) GetDeleteImpact(id uuid.UUID) (*domainFacility.ControlCabinetDeleteImpact, error) {
	// Ensure cabinet exists
	if _, err := s.GetByID(id); err != nil {
		return nil, err
	}

	spsControllerIDs, err := s.spsControllerRepo.GetIDsByControlCabinetID(id)
	if err != nil {
		return nil, err
	}

	spsControllerSystemTypeIDs, err := s.spsControllerSystemRepo.GetIDsBySPSControllerIDs(spsControllerIDs)
	if err != nil {
		return nil, err
	}

	fieldDeviceIDs, err := s.fieldDeviceRepo.GetIDsBySPSControllerSystemTypeIDs(spsControllerSystemTypeIDs)
	if err != nil {
		return nil, err
	}

	bos, err := s.bacnetObjectRepo.GetByFieldDeviceIDs(fieldDeviceIDs)
	if err != nil {
		return nil, err
	}

	specs, err := s.specificationRepo.GetByFieldDeviceIDs(fieldDeviceIDs)
	if err != nil {
		return nil, err
	}

	return &domainFacility.ControlCabinetDeleteImpact{
		ControlCabinetID:              id,
		SPSControllersCount:           len(spsControllerIDs),
		SPSControllerSystemTypesCount: len(spsControllerSystemTypeIDs),
		FieldDevicesCount:             len(fieldDeviceIDs),
		BacnetObjectsCount:            len(bos),
		SpecificationsCount:           len(specs),
	}, nil
}

func (s *ControlCabinetService) List(page, limit int, search string) (*domain.PaginatedList[domainFacility.ControlCabinet], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *ControlCabinetService) ListByBuildingID(buildingID uuid.UUID, page, limit int, search string) (*domain.PaginatedList[domainFacility.ControlCabinet], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.repo.GetPaginatedListByBuildingID(buildingID, domain.PaginationParams{
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

func (s *ControlCabinetService) DeleteByID(id uuid.UUID) error {
	impact, err := s.GetDeleteImpact(id)
	if err != nil {
		return err
	}

	// Nothing references this cabinet => delete directly
	if impact.SPSControllersCount == 0 {
		return s.repo.DeleteByIds([]uuid.UUID{id})
	}

	spsControllerIDs, err := s.spsControllerRepo.GetIDsByControlCabinetID(id)
	if err != nil {
		return err
	}

	spsControllerSystemTypeIDs, err := s.spsControllerSystemRepo.GetIDsBySPSControllerIDs(spsControllerIDs)
	if err != nil {
		return err
	}

	fieldDeviceIDs, err := s.fieldDeviceRepo.GetIDsBySPSControllerSystemTypeIDs(spsControllerSystemTypeIDs)
	if err != nil {
		return err
	}

	// Delete dependents in safe order (soft-delete everywhere)
	if err := s.spsControllerSystemRepo.SoftDeleteBySPSControllerIDs(spsControllerIDs); err != nil {
		return err
	}

	if err := s.bacnetObjectRepo.SoftDeleteByFieldDeviceIDs(fieldDeviceIDs); err != nil {
		return err
	}
	if err := s.specificationRepo.SoftDeleteByFieldDeviceIDs(fieldDeviceIDs); err != nil {
		return err
	}
	if err := s.fieldDeviceRepo.DeleteByIds(fieldDeviceIDs); err != nil {
		return err
	}

	if err := s.spsControllerRepo.DeleteByIds(spsControllerIDs); err != nil {
		return err
	}

	return s.repo.DeleteByIds([]uuid.UUID{id})
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
