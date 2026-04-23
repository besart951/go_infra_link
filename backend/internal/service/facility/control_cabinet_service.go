package facility

import (
	"context"
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
	hierarchyCopier         *HierarchyCopier
}

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
	if err := s.Validate(ctx, controlCabinet, &controlCabinet.ID); err != nil {
		return err
	}
	return s.repo.Update(ctx, controlCabinet)
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

func (s *ControlCabinetService) ensureBuildingExists(ctx context.Context, buildingID uuid.UUID) error {
	return domain.EnsureReferenceExists(ctx, s.buildingRepo, buildingID)
}

func (s *ControlCabinetService) validateRequiredFields(controlCabinet *domainFacility.ControlCabinet) error {
	builder := domain.NewValidationBuilder()
	controlCabinetBuildingIDField.RequireUUID(builder, controlCabinet.BuildingID)
	controlCabinetNr := controlCabinetNumberField.RequireTrimmedPtr(builder, controlCabinet.ControlCabinetNr)
	controlCabinetNumberField.MaxLength(builder, controlCabinetNr, 11)
	return builder.Err()
}

func (s *ControlCabinetService) ensureUnique(ctx context.Context, controlCabinet *domainFacility.ControlCabinet, excludeID *uuid.UUID) error {
	if controlCabinet.ControlCabinetNr == nil || strings.TrimSpace(*controlCabinet.ControlCabinetNr) == "" || controlCabinet.BuildingID == uuid.Nil {
		return nil
	}
	exists, err := s.repo.ExistsControlCabinetNr(ctx, controlCabinet.BuildingID, *controlCabinet.ControlCabinetNr, excludeID)
	if err != nil {
		return err
	}
	if exists {
		return controlCabinetNumberField.UniqueWithinError(buildingScope)
	}
	return nil
}
