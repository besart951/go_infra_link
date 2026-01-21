package facilitysql

import (
	"database/sql"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
)

// Repositories groups per-entity SQL repositories for convenient wiring.
// This is the SQL-based counterpart to internal/repository/facility.
type Repositories struct {
	Buildings       domainFacility.BuildingRepository
	ControlCabinets domainFacility.ControlCabinetRepository
	FieldDevices    domainFacility.FieldDeviceRepository
}

func NewRepositories(db *sql.DB) Repositories {
	return Repositories{
		Buildings:       NewBuildingRepository(db),
		ControlCabinets: NewControlCabinetRepository(db),
		FieldDevices:    NewFieldDeviceRepository(db),
	}
}
