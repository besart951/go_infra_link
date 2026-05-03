package facility

import (
	"context"

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
	hierarchyCopier         *HierarchyCopier
	tx                      txCoordinator
}

const controlCabinetSPSControllerPageLimit = 500

func NewControlCabinetService(
	repo domainFacility.ControlCabinetRepository,
	buildingRepo domainFacility.BuildingRepository,
	spsControllerRepo domainFacility.SPSControllerRepository,
	spsControllerSystemRepo domainFacility.SPSControllerSystemTypeStore,
	fieldDeviceRepo domainFacility.FieldDeviceStore,
	bacnetObjectRepo domainFacility.BacnetObjectStore,
	specificationRepo domainFacility.SpecificationStore,
	hierarchyCopier *HierarchyCopier,
) *ControlCabinetService {
	return &ControlCabinetService{
		repo:                    repo,
		buildingRepo:            buildingRepo,
		spsControllerRepo:       spsControllerRepo,
		spsControllerSystemRepo: spsControllerSystemRepo,
		fieldDeviceRepo:         fieldDeviceRepo,
		bacnetObjectRepo:        bacnetObjectRepo,
		specificationRepo:       specificationRepo,
		hierarchyCopier:         hierarchyCopier,
	}
}

func (s *ControlCabinetService) bindTransactions(tx txCoordinator) {
	s.tx = tx
}

func (s *ControlCabinetService) transaction() facilityTx[*ControlCabinetService] {
	return newFacilityTx(s.tx, s, func(services *Services) *ControlCabinetService {
		return services.ControlCabinet
	})
}

func (s *ControlCabinetService) Create(ctx context.Context, controlCabinet *domainFacility.ControlCabinet) error {
	if err := s.Validate(ctx, controlCabinet, nil); err != nil {
		return err
	}
	return s.repo.Create(ctx, controlCabinet)
}

func (s *ControlCabinetService) GetByID(ctx context.Context, id uuid.UUID) (*domainFacility.ControlCabinet, error) {
	return domain.GetByID(ctx, s.repo, id)
}

func (s *ControlCabinetService) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]domainFacility.ControlCabinet, error) {
	controlCabinets, err := s.repo.GetByIds(ctx, ids)
	if err != nil {
		return nil, err
	}
	return derefSlice(controlCabinets), nil
}

func (s *ControlCabinetService) CopyByID(ctx context.Context, id uuid.UUID) (*domainFacility.ControlCabinet, error) {
	return s.hierarchyCopier.CopyControlCabinetByID(ctx, id)
}

func (s *ControlCabinetService) GetDeleteImpact(ctx context.Context, id uuid.UUID) (*domainFacility.ControlCabinetDeleteImpact, error) {
	// Ensure cabinet exists
	if _, err := s.GetByID(ctx, id); err != nil {
		return nil, err
	}

	spsControllerIDs, err := s.spsControllerRepo.GetIDsByControlCabinetID(ctx, id)
	if err != nil {
		return nil, err
	}

	spsControllerSystemTypeIDs, err := s.spsControllerSystemRepo.GetIDsBySPSControllerIDs(ctx, spsControllerIDs)
	if err != nil {
		return nil, err
	}

	fieldDeviceIDs, err := s.fieldDeviceRepo.GetIDsBySPSControllerSystemTypeIDs(ctx, spsControllerSystemTypeIDs)
	if err != nil {
		return nil, err
	}

	bos, err := s.bacnetObjectRepo.GetByFieldDeviceIDs(ctx, fieldDeviceIDs)
	if err != nil {
		return nil, err
	}

	specs, err := s.specificationRepo.GetByFieldDeviceIDs(ctx, fieldDeviceIDs)
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

func (s *ControlCabinetService) List(ctx context.Context, page, limit int, search string) (*domain.PaginatedList[domainFacility.ControlCabinet], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.repo.GetPaginatedList(ctx, domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *ControlCabinetService) ListByBuildingID(ctx context.Context, buildingID uuid.UUID, page, limit int, search string) (*domain.PaginatedList[domainFacility.ControlCabinet], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.repo.GetPaginatedListByBuildingID(ctx, buildingID, domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *ControlCabinetService) Update(ctx context.Context, controlCabinet *domainFacility.ControlCabinet) error {
	return s.transaction().run(func(txService *ControlCabinetService) error {
		if err := txService.Validate(ctx, controlCabinet, &controlCabinet.ID); err != nil {
			return err
		}
		if err := txService.repo.Update(ctx, controlCabinet); err != nil {
			return err
		}
		return txService.regenerateSPSControllerDeviceNames(ctx, controlCabinet)
	})
}

func (s *ControlCabinetService) Validate(ctx context.Context, controlCabinet *domainFacility.ControlCabinet, excludeID *uuid.UUID) error {
	if err := s.validateRequiredFields(controlCabinet); err != nil {
		return err
	}
	if err := s.ensureBuildingExists(ctx, controlCabinet.BuildingID); err != nil {
		return err
	}
	if err := s.ensureUnique(ctx, controlCabinet, excludeID); err != nil {
		return err
	}
	return nil
}

func (s *ControlCabinetService) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteByIds(ctx, []uuid.UUID{id})
}

func (s *ControlCabinetService) regenerateSPSControllerDeviceNames(ctx context.Context, controlCabinet *domainFacility.ControlCabinet) error {
	building, err := domain.GetByID(ctx, s.buildingRepo, controlCabinet.BuildingID)
	if err != nil {
		return err
	}

	for page := 1; ; page++ {
		result, err := s.spsControllerRepo.GetPaginatedListByControlCabinetID(ctx, controlCabinet.ID, domain.PaginationParams{
			Page:  page,
			Limit: controlCabinetSPSControllerPageLimit,
		})
		if err != nil {
			return err
		}

		for i := range result.Items {
			controller := result.Items[i]
			deviceName, ok := generatedSPSControllerDeviceName(controlCabinet, building, controller.GADevice)
			if !ok || controller.DeviceName == deviceName {
				continue
			}
			controller.DeviceName = deviceName
			if err := s.spsControllerRepo.Update(ctx, &controller); err != nil {
				return err
			}
		}

		if page >= result.TotalPages || len(result.Items) == 0 {
			break
		}
	}

	return nil
}

func (s *ControlCabinetService) ensureBuildingExists(ctx context.Context, buildingID uuid.UUID) error {
	return validateChecks(referenceExists(ctx, s.buildingRepo, buildingID))
}

func (s *ControlCabinetService) validateRequiredFields(controlCabinet *domainFacility.ControlCabinet) error {
	return validateRules(
		requiredUUID(controlCabinetBuildingIDField, controlCabinet.BuildingID),
		requiredTrimmedPtrMax(controlCabinetNumberField, controlCabinet.ControlCabinetNr, 11),
	)
}

func (s *ControlCabinetService) ensureUnique(ctx context.Context, controlCabinet *domainFacility.ControlCabinet, excludeID *uuid.UUID) error {
	if controlCabinet.ControlCabinetNr == nil || controlCabinet.BuildingID == uuid.Nil {
		return nil
	}
	return validateChecks(
		uniqueWithinIfPresent(controlCabinetNumberField, buildingScope, *controlCabinet.ControlCabinetNr, func() (bool, error) {
			return s.repo.ExistsControlCabinetNr(ctx, controlCabinet.BuildingID, *controlCabinet.ControlCabinetNr, excludeID)
		}),
	)
}
