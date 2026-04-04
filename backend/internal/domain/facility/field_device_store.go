package facility

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

// FieldDeviceStore extends the base repository with helper methods
// needed for high-volume uniqueness checks.
type FieldDeviceStore interface {
	FieldDeviceRepository

	// GetIDsBySPSControllerSystemTypeIDs returns IDs of non-deleted field devices
	// that belong to the given SPS controller system type IDs.
	GetIDsBySPSControllerSystemTypeIDs(ctx context.Context, ids []uuid.UUID) ([]uuid.UUID, error)

	// ExistsApparatNrConflict reports whether apparat_nr is already taken
	// for the given (sps_controller_system_type_id, system_part_id, apparat_id) tuple.
	// excludeIDs allows excluding multiple IDs (e.g. for batch updates).
	ExistsApparatNrConflict(ctx context.Context, spsControllerSystemTypeID uuid.UUID, systemPartID *uuid.UUID, apparatID uuid.UUID, apparatNr int, excludeIDs []uuid.UUID) (bool, error)

	// GetUsedApparatNumbers returns a list of used apparat_nr values for the given scope.
	GetUsedApparatNumbers(ctx context.Context, spsControllerSystemTypeID uuid.UUID, systemPartID *uuid.UUID, apparatID uuid.UUID) ([]int, error)

	// GetPaginatedListWithFilters returns paginated field devices with optional filtering
	GetPaginatedListWithFilters(ctx context.Context, params domain.PaginationParams, filters FieldDeviceFilterParams) (*domain.PaginatedList[FieldDevice], error)

	// BulkCreate creates multiple field devices in batches.
	BulkCreate(ctx context.Context, entities []*FieldDevice, batchSize int) error
}
