package facilitysql

import (
	"database/sql"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"gorm.io/gorm"
)

// Repositories groups per-entity SQL repositories for convenient wiring.
// This is the SQL-based counterpart to internal/repository/facility.
type Repositories struct {
	Buildings       domainFacility.BuildingRepository
	ControlCabinets domainFacility.ControlCabinetRepository
	FieldDevices    domainFacility.FieldDeviceStore
}

func NewRepositories(db *sql.DB, driver string, gormDB *gorm.DB) Repositories {
	return Repositories{
		Buildings:       NewBuildingRepository(db, driver),
		ControlCabinets: NewControlCabinetRepository(db, driver),
		FieldDevices:    NewFieldDeviceRepository(gormDB),
	}
}
