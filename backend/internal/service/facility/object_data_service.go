package facility

import (
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
		baseService: newBase[domainFacility.ObjectData](repo, 10),
		extRepo:     repo,
	}
}

func (s *ObjectDataService) Create(objectData *domainFacility.ObjectData) error {
	if err := s.ensureUnique(objectData, nil); err != nil {
		return err
	}
	return s.repo.Create(objectData)
}

func (s *ObjectDataService) Update(objectData *domainFacility.ObjectData) error {
	if err := s.ensureUnique(objectData, &objectData.ID); err != nil {
		return err
	}
	return s.repo.Update(objectData)
}

func (s *ObjectDataService) ListByApparatID(page, limit int, search string, apparatID uuid.UUID) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	page, limit = domain.NormalizePagination(page, limit, s.defaultLimit)
	return s.extRepo.GetPaginatedListByApparatID(apparatID, domain.PaginationParams{Page: page, Limit: limit, Search: search})
}

func (s *ObjectDataService) ListBySystemPartID(page, limit int, search string, systemPartID uuid.UUID) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	page, limit = domain.NormalizePagination(page, limit, s.defaultLimit)
	return s.extRepo.GetPaginatedListBySystemPartID(systemPartID, domain.PaginationParams{Page: page, Limit: limit, Search: search})
}

func (s *ObjectDataService) ListByApparatAndSystemPartID(page, limit int, search string, apparatID, systemPartID uuid.UUID) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	page, limit = domain.NormalizePagination(page, limit, s.defaultLimit)
	return s.extRepo.GetPaginatedListByApparatAndSystemPartID(apparatID, systemPartID, domain.PaginationParams{Page: page, Limit: limit, Search: search})
}

func (s *ObjectDataService) GetBacnetObjectIDs(id uuid.UUID) ([]uuid.UUID, error) {
	return s.extRepo.GetBacnetObjectIDs(id)
}

func (s *ObjectDataService) GetApparatIDs(id uuid.UUID) ([]uuid.UUID, error) {
	objectData, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}
	return extractIDs(objectData.Apparats, func(a *domainFacility.Apparat) uuid.UUID { return a.ID }), nil
}

func (s *ObjectDataService) ExistsByDescription(projectID *uuid.UUID, description string, excludeID *uuid.UUID) (bool, error) {
	return s.extRepo.ExistsByDescription(projectID, description, excludeID)
}

func (s *ObjectDataService) ensureUnique(objectData *domainFacility.ObjectData, excludeID *uuid.UUID) error {
	description := strings.TrimSpace(objectData.Description)
	if description == "" {
		return nil
	}
	exists, err := s.extRepo.ExistsByDescription(objectData.ProjectID, description, excludeID)
	if err != nil {
		return err
	}
	if exists {
		return domain.NewValidationError().Add("objectdata.description", "description must be unique")
	}
	return nil
}
