package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type FieldDevice struct {
	domain.Base
	BMK                       *string
	Description               *string
	ApparatNr                 int
	SPSControllerSystemTypeID uuid.UUID               `gorm:"type:uuid;not null;index"`
	SPSControllerSystemType   SPSControllerSystemType `gorm:"foreignKey:SPSControllerSystemTypeID"`
	SystemPartID              uuid.UUID               `gorm:"type:uuid;index"`
	SystemPart                SystemPart              `gorm:"foreignKey:SystemPartID"`
	SpecificationID           *uuid.UUID              `gorm:"type:uuid;index"`
	Specification             *Specification          `gorm:"foreignKey:SpecificationID"`
	ApparatID                 uuid.UUID               `gorm:"type:uuid;not null;index"`
	Apparat                   Apparat                 `gorm:"foreignKey:ApparatID"`

	BacnetObjects []BacnetObject `gorm:"foreignKey:FieldDeviceID"`
}

// FieldDeviceOptions contains all metadata needed for creating/editing field devices
type FieldDeviceOptions struct {
	Apparats            []Apparat
	SystemParts         []SystemPart
	ObjectDatas         []ObjectData
	ApparatToSystemPart map[uuid.UUID][]uuid.UUID // apparat_id -> [system_part_ids]
	ObjectDataToApparat map[uuid.UUID][]uuid.UUID // object_data_id -> [apparat_ids]
}

// FieldDeviceFilterParams contains optional filter parameters for listing field devices
type FieldDeviceFilterParams struct {
	BuildingID                *uuid.UUID
	ControlCabinetID          *uuid.UUID
	SPSControllerID           *uuid.UUID
	SPSControllerSystemTypeID *uuid.UUID
}

// FieldDeviceCreateItem represents a single field device to create in a multi-create operation
type FieldDeviceCreateItem struct {
	FieldDevice   *FieldDevice
	ObjectDataID  *uuid.UUID
	BacnetObjects []BacnetObject
}

// FieldDeviceCreateResult represents the result of creating a single field device
type FieldDeviceCreateResult struct {
	Index       int          // Index in the original request array
	Success     bool         // Whether the creation succeeded
	FieldDevice *FieldDevice // The created field device (nil if failed)
	Error       string       // Error message if failed (empty if succeeded)
	ErrorField  string       // Specific field that caused the error (if applicable)
}

// FieldDeviceMultiCreateResult represents the result of a multi-create operation
type FieldDeviceMultiCreateResult struct {
	Results       []FieldDeviceCreateResult
	TotalRequests int
	SuccessCount  int
	FailureCount  int
}
