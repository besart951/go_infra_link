package fielddevice

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

// FieldDeviceStore extends the base repository with helper methods
// needed for high-volume uniqueness checks.
type FieldDeviceStore interface {
	domainFacility.FieldDeviceRepository

	// GetIDsBySPSControllerSystemTypeIDs returns IDs of non-deleted field devices
	// that belong to the given SPS controller system type IDs.
	GetIDsBySPSControllerSystemTypeIDs(ctx context.Context, ids []uuid.UUID) ([]uuid.UUID, error)

	// ListIDsBySPSControllerSystemTypeIDsAfter returns one deterministic ID page
	// ordered ascending. afterID is an exclusive cursor; nil starts the scan.
	ListIDsBySPSControllerSystemTypeIDsAfter(
		ctx context.Context,
		ids []uuid.UUID,
		afterID *uuid.UUID,
		limit int,
	) ([]uuid.UUID, error)

	// ExistsApparatNrConflict reports whether apparat_nr is already taken
	// for the given (sps_controller_system_type_id, system_part_id, apparat_id) tuple.
	// excludeIDs allows excluding multiple IDs (e.g. for batch updates).
	ExistsApparatNrConflict(ctx context.Context, spsControllerSystemTypeID uuid.UUID, systemPartID uuid.UUID, apparatID uuid.UUID, apparatNr int, excludeIDs []uuid.UUID) (bool, error)

	// GetUsedApparatNumbers returns a list of used apparat_nr values for the given scope.
	GetUsedApparatNumbers(ctx context.Context, spsControllerSystemTypeID uuid.UUID, systemPartID uuid.UUID, apparatID uuid.UUID) ([]int, error)

	// GetPaginatedListWithFilters returns paginated field devices with optional filtering
	GetPaginatedListWithFilters(ctx context.Context, params domain.PaginationParams, filters domainFacility.FieldDeviceFilterParams) (*domain.PaginatedList[domainFacility.FieldDevice], error)

	// BulkCreate creates multiple field devices in batches.
	BulkCreate(ctx context.Context, entities []*domainFacility.FieldDevice, batchSize int) error

	// AssignSpecificationIDs completes the cyclic FieldDevice/Specification
	// relationship after both sides have been created. The map key is the field
	// device ID and the value is its specification ID.
	AssignSpecificationIDs(ctx context.Context, assignments map[uuid.UUID]uuid.UUID) error
}
