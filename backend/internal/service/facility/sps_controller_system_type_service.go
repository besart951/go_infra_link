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
	fieldDeviceSvc    *FieldDeviceService
	specificationRepo domainFacility.SpecificationStore
	bacnetObjectRepo  domainFacility.BacnetObjectStore
	hierarchyCopier   *HierarchyCopier
}

func NewSPSControllerSystemTypeService(
	repo domainFacility.SPSControllerSystemTypeStore,
	systemTypeRepo domainFacility.SystemTypeRepository,
	fieldDeviceRepo domainFacility.FieldDeviceStore,
	fieldDeviceSvc *FieldDeviceService,
	specificationRepo domainFacility.SpecificationStore,
	bacnetObjectRepo domainFacility.BacnetObjectStore,
	hierarchyCopier *HierarchyCopier,
) *SPSControllerSystemTypeService {
	return &SPSControllerSystemTypeService{
		repo:              repo,
		systemTypeRepo:    systemTypeRepo,
		fieldDeviceRepo:   fieldDeviceRepo,
		fieldDeviceSvc:    fieldDeviceSvc,
		specificationRepo: specificationRepo,
		bacnetObjectRepo:  bacnetObjectRepo,
		hierarchyCopier:   hierarchyCopier,
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

func (s *SPSControllerSystemTypeService) GetByID(id uuid.UUID) (*domainFacility.SPSControllerSystemType, error) {
	return domain.GetByID(s.repo, id)
}

func (s *SPSControllerSystemTypeService) CopyByID(id uuid.UUID) (*domainFacility.SPSControllerSystemType, error) {
	return s.hierarchyCopier.CopySPSControllerSystemTypeByID(id)
}

func (s *SPSControllerSystemTypeService) DeleteByID(id uuid.UUID) error {
	if _, err := domain.GetByID(s.repo, id); err != nil {
		return err
	}

	fieldDeviceIDs, err := s.fieldDeviceRepo.GetIDsBySPSControllerSystemTypeIDs([]uuid.UUID{id})
	if err != nil {
		return err
	}
	if err := s.fieldDeviceSvc.DeleteByIDs(fieldDeviceIDs); err != nil {
		return err
	}

	return s.repo.DeleteByIds([]uuid.UUID{id})
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
