package facility

import (
	"strings"

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
	if err := s.Validate(controlCabinet, nil); err != nil {
		return err
	}
	if err := s.ensureBuildingExists(controlCabinet.BuildingID); err != nil {
		return err
	}
	return s.repo.Create(controlCabinet)
}

func (s *ControlCabinetService) GetByID(id uuid.UUID) (*domainFacility.ControlCabinet, error) {
	return domain.GetByID(s.repo, id)
}

func (s *ControlCabinetService) GetByIDs(ids []uuid.UUID) ([]domainFacility.ControlCabinet, error) {
	controlCabinets, err := s.repo.GetByIds(ids)
	if err != nil {
		return nil, err
	}
	items := make([]domainFacility.ControlCabinet, len(controlCabinets))
	for i, item := range controlCabinets {
		items[i] = *item
	}
	return items, nil
}

func (s *ControlCabinetService) CopyByID(id uuid.UUID) (*domainFacility.ControlCabinet, error) {
	original, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	baseNr := ""
	if original.ControlCabinetNr != nil {
		baseNr = strings.TrimSpace(*original.ControlCabinetNr)
	}

	nextNr, err := s.nextAvailableControlCabinetNr(original.BuildingID, baseNr)
	if err != nil {
		return nil, err
	}

	copyEntity := &domainFacility.ControlCabinet{
		BuildingID:       original.BuildingID,
		ControlCabinetNr: &nextNr,
	}

	if err := s.Create(copyEntity); err != nil {
		return nil, err
	}

	return copyEntity, nil
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
	if err := s.Validate(controlCabinet, &controlCabinet.ID); err != nil {
		return err
	}
	if err := s.ensureBuildingExists(controlCabinet.BuildingID); err != nil {
		return err
	}
	return s.repo.Update(controlCabinet)
}

func (s *ControlCabinetService) Validate(controlCabinet *domainFacility.ControlCabinet, excludeID *uuid.UUID) error {
	if err := s.validateRequiredFields(controlCabinet); err != nil {
		return err
	}
	if err := s.ensureUnique(controlCabinet, excludeID); err != nil {
		return err
	}
	return nil
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
	_, err := domain.GetByID(s.buildingRepo, buildingID)
	return err
}

func (s *ControlCabinetService) validateRequiredFields(controlCabinet *domainFacility.ControlCabinet) error {
	ve := domain.NewValidationError()
	if controlCabinet.BuildingID == uuid.Nil {
		ve = ve.Add("controlcabinet.building_id", "building_id is required")
	}
	if controlCabinet.ControlCabinetNr == nil || strings.TrimSpace(*controlCabinet.ControlCabinetNr) == "" {
		ve = ve.Add("controlcabinet.control_cabinet_nr", "control_cabinet_nr is required")
	} else if len(strings.TrimSpace(*controlCabinet.ControlCabinetNr)) > 11 {
		ve = ve.Add("controlcabinet.control_cabinet_nr", "control_cabinet_nr must be 11 characters or less")
	}
	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
}

func (s *ControlCabinetService) ensureUnique(controlCabinet *domainFacility.ControlCabinet, excludeID *uuid.UUID) error {
	if controlCabinet.ControlCabinetNr == nil || strings.TrimSpace(*controlCabinet.ControlCabinetNr) == "" || controlCabinet.BuildingID == uuid.Nil {
		return nil
	}
	exists, err := s.repo.ExistsControlCabinetNr(controlCabinet.BuildingID, *controlCabinet.ControlCabinetNr, excludeID)
	if err != nil {
		return err
	}
	if exists {
		return domain.NewValidationError().Add("controlcabinet.control_cabinet_nr", "control_cabinet_nr must be unique within the building")
	}
	return nil
}

func (s *ControlCabinetService) nextAvailableControlCabinetNr(buildingID uuid.UUID, base string) (string, error) {
	for i := 1; i <= 9999; i++ {
		candidate := nextIncrementedValue(base, i, 11)
		exists, err := s.repo.ExistsControlCabinetNr(buildingID, candidate, nil)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}

	return "", domain.ErrConflict
}
