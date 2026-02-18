package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type SPSControllerSystemTypeService struct {
	repo              domainFacility.SPSControllerSystemTypeStore
	systemTypeRepo    domainFacility.SystemTypeRepository
	fieldDeviceRepo   domainFacility.FieldDeviceStore
	specificationRepo domainFacility.SpecificationStore
	bacnetObjectRepo  domainFacility.BacnetObjectStore
}

func NewSPSControllerSystemTypeService(
	repo domainFacility.SPSControllerSystemTypeStore,
	systemTypeRepo domainFacility.SystemTypeRepository,
	fieldDeviceRepo domainFacility.FieldDeviceStore,
	specificationRepo domainFacility.SpecificationStore,
	bacnetObjectRepo domainFacility.BacnetObjectStore,
) *SPSControllerSystemTypeService {
	return &SPSControllerSystemTypeService{
		repo:              repo,
		systemTypeRepo:    systemTypeRepo,
		fieldDeviceRepo:   fieldDeviceRepo,
		specificationRepo: specificationRepo,
		bacnetObjectRepo:  bacnetObjectRepo,
	}
}

func (s *SPSControllerSystemTypeService) List(page, limit int, search string) (*domain.PaginatedList[domainFacility.SPSControllerSystemType], error) {
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *SPSControllerSystemTypeService) ListBySPSControllerID(spsControllerID uuid.UUID, page, limit int, search string) (*domain.PaginatedList[domainFacility.SPSControllerSystemType], error) {
	return s.repo.GetPaginatedListBySPSControllerID(spsControllerID, domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *SPSControllerSystemTypeService) CopyByID(id uuid.UUID) (*domainFacility.SPSControllerSystemType, error) {
	original, err := domain.GetByID(s.repo, id)
	if err != nil {
		return nil, err
	}

	systemType, err := domain.GetByID(s.systemTypeRepo, original.SystemTypeID)
	if err != nil {
		return nil, err
	}

	existing, err := s.repo.ListBySPSControllerID(original.SPSControllerID)
	if err != nil {
		return nil, err
	}

	usedNumbers := make(map[int]struct{}, len(existing))
	for _, item := range existing {
		if item.SystemTypeID != original.SystemTypeID || item.Number == nil {
			continue
		}
		usedNumbers[*item.Number] = struct{}{}
	}

	nextNumber, ok := findLowestAvailableNumber(systemType.NumberMin, systemType.NumberMax, usedNumbers)
	if !ok {
		return nil, domain.NewValidationError().Add("spscontroller.system_types", "no available number in the system type range")
	}

	copyNumber := nextNumber
	copyEntity := &domainFacility.SPSControllerSystemType{
		Number:          &copyNumber,
		DocumentName:    original.DocumentName,
		SPSControllerID: original.SPSControllerID,
		SystemTypeID:    original.SystemTypeID,
	}
	if err := s.repo.Create(copyEntity); err != nil {
		return nil, err
	}

	if err := s.copyFieldDevicesForSystemType(original.ID, copyEntity.ID); err != nil {
		return nil, err
	}

	return domain.GetByID(s.repo, copyEntity.ID)
}

func (s *SPSControllerSystemTypeService) copyFieldDevicesForSystemType(originalSystemTypeID, newSystemTypeID uuid.UUID) error {
	page := 1
	filters := domainFacility.FieldDeviceFilterParams{SPSControllerSystemTypeID: &originalSystemTypeID}

	for {
		result, err := s.fieldDeviceRepo.GetPaginatedListWithFilters(domain.PaginationParams{Page: page, Limit: 500}, filters)
		if err != nil {
			return err
		}

		for _, originalFieldDevice := range result.Items {
			fieldDeviceCopy := &domainFacility.FieldDevice{
				BMK:                       originalFieldDevice.BMK,
				Description:               originalFieldDevice.Description,
				ApparatNr:                 originalFieldDevice.ApparatNr,
				SPSControllerSystemTypeID: newSystemTypeID,
				SystemPartID:              originalFieldDevice.SystemPartID,
				ApparatID:                 originalFieldDevice.ApparatID,
			}
			if err := s.fieldDeviceRepo.Create(fieldDeviceCopy); err != nil {
				return err
			}

			if err := copySpecificationForFieldDevice(s.specificationRepo, originalFieldDevice.ID, fieldDeviceCopy.ID); err != nil {
				return err
			}
			if err := copyBacnetObjectsForFieldDevice(s.bacnetObjectRepo, originalFieldDevice.ID, fieldDeviceCopy.ID); err != nil {
				return err
			}
		}

		if page >= result.TotalPages || len(result.Items) == 0 {
			break
		}
		page++
	}

	return nil
}
