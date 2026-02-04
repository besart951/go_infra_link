package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

// ObjectDataStore extends the base repository with helper methods
// needed to apply ObjectData templates.
type ObjectDataStore interface {
	ObjectDataRepository

	GetBacnetObjectIDs(objectDataID uuid.UUID) ([]uuid.UUID, error)
	GetTemplates() ([]*ObjectData, error)
	GetForProject(projectID uuid.UUID) ([]*ObjectData, error)
	GetPaginatedListForProject(projectID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[ObjectData], error)
	GetPaginatedListByApparatID(apparatID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[ObjectData], error)
	GetPaginatedListBySystemPartID(systemPartID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[ObjectData], error)
	GetPaginatedListByApparatAndSystemPartID(apparatID, systemPartID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[ObjectData], error)
	GetPaginatedListForProjectByApparatID(projectID, apparatID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[ObjectData], error)
	GetPaginatedListForProjectBySystemPartID(projectID, systemPartID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[ObjectData], error)
	GetPaginatedListForProjectByApparatAndSystemPartID(projectID, apparatID, systemPartID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[ObjectData], error)
}
