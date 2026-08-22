package objectdata

import (
	"context"
	"time"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

// BacnetObjectTemplate is owned by exactly one ObjectData aggregate. It is a
// prototype and can never be attached directly to a FieldDevice.
type BacnetObjectTemplate struct {
	ID                  uuid.UUID
	CreatedAt           time.Time
	UpdatedAt           time.Time
	Version             uint64
	ObjectDataID        uuid.UUID
	TextFix             string
	Description         *string
	GMSVisible          bool
	Optional            bool
	TextIndividual      *string
	SoftwareType        domainFacility.BacnetSoftwareType
	SoftwareNumber      uint16
	HardwareType        domainFacility.BacnetHardwareType
	HardwareQuantity    uint8
	SoftwareReferenceID *uuid.UUID
	StateTextID         *uuid.UUID
	NotificationClassID *uuid.UUID
	AlarmTypeID         *uuid.UUID
}

type BacnetObjectTemplateAlarmValue struct {
	ID               uuid.UUID
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Version          uint64
	TemplateID       uuid.UUID
	AlarmTypeFieldID uuid.UUID
	ValueNumber      *float64
	ValueInteger     *int64
	ValueBoolean     *bool
	ValueString      *string
	ValueJSON        *string
	UnitID           *uuid.UUID
	Source           string
}

type BacnetObjectTemplateStore interface {
	ListByObjectDataID(context.Context, uuid.UUID) ([]BacnetObjectTemplate, error)
	Replace(context.Context, uuid.UUID, []BacnetObjectTemplate) error
}
