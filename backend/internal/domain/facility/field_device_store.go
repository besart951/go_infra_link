package facility

import "github.com/google/uuid"

// FieldDeviceStore extends the base repository with helper methods
// needed for high-volume uniqueness checks.
type FieldDeviceStore interface {
	FieldDeviceRepository

	// ExistsApparatNrConflict reports whether apparat_nr is already taken
	// for the given (sps_controller_system_type_id, system_part_id, apparat_id) tuple.
	// excludeID can be set for updates.
	// Note: systemPartID is now required (NOT NULL)
	ExistsApparatNrConflict(spsControllerSystemTypeID uuid.UUID, systemPartID uuid.UUID, apparatID uuid.UUID, apparatNr int, excludeID *uuid.UUID) (bool, error)
}
