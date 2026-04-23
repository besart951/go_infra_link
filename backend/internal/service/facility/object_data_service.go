package facility

import (
	"context"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type ObjectDataService struct {
	baseService[domainFacility.ObjectData]
	extRepo domainFacility.ObjectDataStore
}

func NewObjectDataService(repo domainFacility.ObjectDataStore) *ObjectDataService {
	return &ObjectDataService{
		baseService: newBase(repo, 10),
		extRepo:     repo,
	}
}

func (s *ObjectDataService) Create(ctx context.Context, objectData *domainFacility.ObjectData) error {
	if err := s.ensureUnique(ctx, objectData, nil); err != nil {
		return err
	}
	return s.repo.Create(ctx, objectData)
}

func (s *ObjectDataService) Update(ctx context.Context, objectData *domainFacility.ObjectData) error {
	if err := s.ensureUnique(ctx, objectData, &objectData.ID); err != nil {
		return err
	}
	return s.repo.Update(ctx, objectData)
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

func (s *ObjectDataService) ensureUnique(ctx context.Context, objectData *domainFacility.ObjectData, excludeID *uuid.UUID) error {
	description := strings.TrimSpace(objectData.Description)
	if description == "" {
		return nil
	}
	exists, err := s.extRepo.ExistsByDescription(ctx, objectData.ProjectID, description, excludeID)
	if err != nil {
		return err
	}
	if exists {
		return domain.NewValidationError().Add("objectdata.description", "description must be unique")
	}
	return nil
}
