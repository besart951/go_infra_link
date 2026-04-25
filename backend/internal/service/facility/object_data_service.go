package facility

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type ObjectDataService struct {
	baseService[domainFacility.ObjectData]
	extRepo               domainFacility.ObjectDataStore
	bacnetObjectRepo      domainFacility.BacnetObjectStore
	objectDataBacnetStore domainFacility.ObjectDataBacnetObjectStore
	apparatRepo           domainFacility.ApparatRepository
	alarmDefinitionRepo   domainFacility.AlarmDefinitionRepository
	alarmTypeRepo         domainFacility.AlarmTypeRepository
	tx                    txCoordinator
}

func NewObjectDataService(
	repo domainFacility.ObjectDataStore,
	bacnetObjectRepo domainFacility.BacnetObjectStore,
	objectDataBacnetStore domainFacility.ObjectDataBacnetObjectStore,
	apparatRepo domainFacility.ApparatRepository,
	alarmDefinitionRepo domainFacility.AlarmDefinitionRepository,
	alarmTypeRepo domainFacility.AlarmTypeRepository,
) *ObjectDataService {
	return &ObjectDataService{
		baseService:           newBase(repo, 10),
		extRepo:               repo,
		bacnetObjectRepo:      bacnetObjectRepo,
		objectDataBacnetStore: objectDataBacnetStore,
		apparatRepo:           apparatRepo,
		alarmDefinitionRepo:   alarmDefinitionRepo,
		alarmTypeRepo:         alarmTypeRepo,
	}
}

func (s *ObjectDataService) bindTransactions(tx txCoordinator) {
	s.tx = tx
}

func (s *ObjectDataService) transaction() facilityTx[*ObjectDataService] {
	return newFacilityTx(s.tx, s, func(services *Services) *ObjectDataService {
		return services.ObjectData
	})
}

func (s *ObjectDataService) template() objectDataTemplate {
	return objectDataTemplate{
		objectDataRepo:        s.extRepo,
		bacnetObjectRepo:      s.bacnetObjectRepo,
		objectDataBacnetStore: s.objectDataBacnetStore,
		apparatRepo:           s.apparatRepo,
		alarmDefinitionRepo:   s.alarmDefinitionRepo,
		alarmTypeRepo:         s.alarmTypeRepo,
	}
}

func (s *ObjectDataService) Create(ctx context.Context, objectData *domainFacility.ObjectData) error {
	if err := s.template().ensureDescriptionUnique(ctx, objectData, nil); err != nil {
		return err
	}
	return s.repo.Create(ctx, objectData)
}

func (s *ObjectDataService) Update(ctx context.Context, objectData *domainFacility.ObjectData) error {
	if err := s.template().ensureDescriptionUnique(ctx, objectData, &objectData.ID); err != nil {
		return err
	}
	return s.repo.Update(ctx, objectData)
}

func (s *ObjectDataService) CreateTemplate(ctx context.Context, input domainFacility.ObjectDataTemplateCreate) (*domainFacility.ObjectData, error) {
	return runWithFacilityTxResult(s.transaction(), func(txService *ObjectDataService) (*domainFacility.ObjectData, error) {
		return txService.template().create(ctx, input)
	})
}

func (s *ObjectDataService) UpdateTemplate(ctx context.Context, id uuid.UUID, input domainFacility.ObjectDataTemplateUpdate) (*domainFacility.ObjectData, error) {
	return runWithFacilityTxResult(s.transaction(), func(txService *ObjectDataService) (*domainFacility.ObjectData, error) {
		return txService.template().update(ctx, id, input)
	})
}

func (s *ObjectDataService) ListByApparatID(ctx context.Context, page, limit int, search string, apparatID uuid.UUID) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	page, limit = domain.NormalizePagination(page, limit, s.defaultLimit)
	return s.extRepo.GetPaginatedListByApparatID(ctx, apparatID, domain.PaginationParams{Page: page, Limit: limit, Search: search})
}

func (s *ObjectDataService) ListBySystemPartID(ctx context.Context, page, limit int, search string, systemPartID uuid.UUID) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	page, limit = domain.NormalizePagination(page, limit, s.defaultLimit)
	return s.extRepo.GetPaginatedListBySystemPartID(ctx, systemPartID, domain.PaginationParams{Page: page, Limit: limit, Search: search})
}

func (s *ObjectDataService) ListByApparatAndSystemPartID(ctx context.Context, page, limit int, search string, apparatID, systemPartID uuid.UUID) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	page, limit = domain.NormalizePagination(page, limit, s.defaultLimit)
	return s.extRepo.GetPaginatedListByApparatAndSystemPartID(ctx, apparatID, systemPartID, domain.PaginationParams{Page: page, Limit: limit, Search: search})
}

func (s *ObjectDataService) GetBacnetObjectIDs(ctx context.Context, id uuid.UUID) ([]uuid.UUID, error) {
	return s.extRepo.GetBacnetObjectIDs(ctx, id)
}

func (s *ObjectDataService) GetApparatIDs(ctx context.Context, id uuid.UUID) ([]uuid.UUID, error) {
	objectData, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return extractIDs(objectData.Apparats, func(a *domainFacility.Apparat) uuid.UUID { return a.ID }), nil
}

func (s *ObjectDataService) ExistsByDescription(ctx context.Context, projectID *uuid.UUID, description string, excludeID *uuid.UUID) (bool, error) {
	return s.extRepo.ExistsByDescription(ctx, projectID, description, excludeID)
}
