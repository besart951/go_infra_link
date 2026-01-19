package domain

import "github.com/google/uuid"

type Building struct {
	Base
	IWSCode       string `gorm:"size:4;index"`
	BuildingGroup int

	ControlCabinets []ControlCabinet `gorm:"foreignKey:BuildingID"`
}

type ControlCabinet struct {
	Base
	BuildingID       uuid.UUID
	ProjectID        *uuid.UUID
	ControlCabinetNr string `gorm:"size:11;index"`

	SPSControllers []SPSController `gorm:"foreignKey:ControlCabinetID"`
}

type SPSController struct {
	Base
	ControlCabinetID uuid.UUID
	DeviceName       string
	IPAddress        string
	// ... other fields
}

type FieldDevice struct {
	Base
	BMK             string `gorm:"size:10;index"`
	ApparatID       uuid.UUID
	SPSControllerID *uuid.UUID // via system type relation path simplified here
	ProjectID       *uuid.UUID

	Apparat Apparat `gorm:"foreignKey:ApparatID"`
}

type Apparat struct {
	Base
	ShortName string `gorm:"uniqueIndex:idx_apparat_name"`
	Name      string `gorm:"uniqueIndex:idx_apparat_name"`
}

type FacilityRepository interface {
	GetBuildingByIds(ids []uuid.UUID) ([]*Building, error)
	CreateBuilding(entity *Building) error
	UpdateBuilding(entity *Building) error
	DeleteBuildingByIds(ids []uuid.UUID) error
	GetPaginatedBuildings(params PaginationParams) (*PaginatedList[Building], error)
}
