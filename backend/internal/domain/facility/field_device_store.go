package facility

import "github.com/google/uuid"

// FieldDeviceStore extends the base repository with helper methods
// needed for high-volume uniqueness checks.
type FieldDeviceStore interface {
	FieldDeviceRepository

	// GetIDsBySPSControllerSystemTypeIDs returns IDs of non-deleted field devices
	// that belong to the given SPS controller system type IDs.
	GetIDsBySPSControllerSystemTypeIDs(ids []uuid.UUID) ([]uuid.UUID, error)

	// ExistsApparatNrConflict reports whether apparat_nr is already taken
	// for the given (sps_controller_system_type_id, system_part_id, apparat_id) tuple.
	// excludeID can be set for updates.
	ExistsApparatNrConflict(spsControllerSystemTypeID uuid.UUID, systemPartID *uuid.UUID, apparatID uuid.UUID, apparatNr int, excludeID *uuid.UUID) (bool, error)

	// GetUsedApparatNumbers returns a list of used apparat_nr values for the given scope.
	GetUsedApparatNumbers(spsControllerSystemTypeID uuid.UUID, systemPartID *uuid.UUID, apparatID uuid.UUID) ([]int, error)
}
