package facility

import (
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type ObjectDataService struct {
	repo domainFacility.ObjectDataStore
}

func NewObjectDataService(repo domainFacility.ObjectDataStore) *ObjectDataService {
	return &ObjectDataService{repo: repo}
}

func (s *ObjectDataService) Create(objectData *domainFacility.ObjectData) error {
	if err := s.ensureUnique(objectData, nil); err != nil {
		return err
	}
	return s.repo.Create(objectData)
}

func (s *ObjectDataService) List(page, limit int, search string) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *ObjectDataService) ListByApparatID(page, limit int, search string, apparatID uuid.UUID) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.repo.GetPaginatedListByApparatID(apparatID, domain.PaginationParams{Page: page, Limit: limit, Search: search})
}

func (s *ObjectDataService) ListBySystemPartID(page, limit int, search string, systemPartID uuid.UUID) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.repo.GetPaginatedListBySystemPartID(systemPartID, domain.PaginationParams{Page: page, Limit: limit, Search: search})
}

func (s *ObjectDataService) ListByApparatAndSystemPartID(page, limit int, search string, apparatID, systemPartID uuid.UUID) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.repo.GetPaginatedListByApparatAndSystemPartID(apparatID, systemPartID, domain.PaginationParams{Page: page, Limit: limit, Search: search})
}

func (s *ObjectDataService) GetByID(id uuid.UUID) (*domainFacility.ObjectData, error) {
	items, err := s.repo.GetByIds([]uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, domain.ErrNotFound
	}
	return items[0], nil
}

func (s *ObjectDataService) Update(objectData *domainFacility.ObjectData) error {
	if err := s.ensureUnique(objectData, &objectData.ID); err != nil {
		return err
	}
	return s.repo.Update(objectData)
}

func (s *ObjectDataService) ensureUnique(objectData *domainFacility.ObjectData, excludeID *uuid.UUID) error {
	description := strings.TrimSpace(objectData.Description)
	if description == "" {
		return nil
	}

	exists, err := s.repo.ExistsByDescription(objectData.ProjectID, description, excludeID)
	if err != nil {
		return err
	}
	if exists {
		return domain.NewValidationError().Add("objectdata.description", "description must be unique")
	}
	return nil
}

func (s *ObjectDataService) DeleteByID(id uuid.UUID) error {
	return s.repo.DeleteByIds([]uuid.UUID{id})
}

func (s *ObjectDataService) GetBacnetObjectIDs(id uuid.UUID) ([]uuid.UUID, error) {
	return s.repo.GetBacnetObjectIDs(id)
}

func (s *ObjectDataService) GetApparatIDs(id uuid.UUID) ([]uuid.UUID, error) {
	objectData, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Get the IDs from the loaded apparats
	apparatIDs := make([]uuid.UUID, 0, len(objectData.Apparats))
	for _, apparat := range objectData.Apparats {
		if apparat != nil {
			apparatIDs = append(apparatIDs, apparat.ID)
		}
	}
	return apparatIDs, nil
}

func (s *ObjectDataService) ExistsByDescription(projectID *uuid.UUID, description string, excludeID *uuid.UUID) (bool, error) {
	return s.repo.ExistsByDescription(projectID, description, excludeID)
}
