package collaboration

import (
	"time"

	"github.com/google/uuid"
)

const SchemaVersionV1 uint16 = 1

type Envelope struct {
	SchemaVersion uint16
	EventID       uuid.UUID
	OperationID   uuid.UUID
	CorrelationID uuid.UUID
	ProjectID     uuid.UUID
	ActorID       *uuid.UUID
	OccurredAt    time.Time
	Sequence      *uint64
}

type FacilityScope string

const (
	FacilityScopeControlCabinet FacilityScope = "control_cabinet"
	FacilityScopeSPSController  FacilityScope = "sps_controller"
	FacilityScopeFieldDevice    FacilityScope = "field_device"
	FacilityScopeProject        FacilityScope = "project"
)

type Command interface {
	isCollaborationCommand()
}

// FacilityHierarchyRefreshRequired asks clients to reconcile a committed
// facility scope with PostgreSQL. FullRefresh intentionally omits EntityIDs on
// the version-1 wire contract.
type FacilityHierarchyRefreshRequired struct {
	Envelope
	Scope       FacilityScope
	EntityIDs   []uuid.UUID
	FullRefresh bool
}

func (FacilityHierarchyRefreshRequired) isCollaborationCommand() {}

// FieldDeviceUpdated represents one committed FieldDevice update. The current
// realtime adapter translates it to a version-1 targeted refresh; later wire
// versions may carry a typed delta without changing mutation handlers.
type FieldDeviceUpdated struct {
	Envelope
	FieldDeviceID uuid.UUID
}

func (FieldDeviceUpdated) isCollaborationCommand() {}

// FieldDeviceMoved represents a committed parent move. Version-one clients
// still receive a targeted refresh, while the typed command preserves the old
// and new parents for later protocol versions and independent tests.
type FieldDeviceMoved struct {
	Envelope
	FieldDeviceID                 uuid.UUID
	FromSPSControllerSystemTypeID uuid.UUID
	ToSPSControllerSystemTypeID   uuid.UUID
}

func (FieldDeviceMoved) isCollaborationCommand() {}

// FieldDeviceDeleted carries the pre-delete parent identity for consumers that
// need hierarchy context. The version-one adapter uses the FieldDevice ID for
// an authoritative targeted refresh.
type FieldDeviceDeleted struct {
	Envelope
	FieldDeviceID             uuid.UUID
	SPSControllerSystemTypeID uuid.UUID
}

func (FieldDeviceDeleted) isCollaborationCommand() {}

// FieldDeviceState contains exactly the committed FieldDevice fields consumed
// by the existing version-one entity delta. It deliberately excludes loaded
// associations and persistence-only state.
type FieldDeviceState struct {
	ID                        uuid.UUID
	BMK                       *string
	Description               *string
	TextFix                   *string
	ApparatNumber             int
	SPSControllerSystemTypeID uuid.UUID
	SystemPartID              uuid.UUID
	SpecificationID           *uuid.UUID
	ApparatID                 uuid.UUID
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}

// FieldDevicesCreated represents the successful portion of one committed
// project-scoped multi-create operation. Partial failures remain in the HTTP
// result and are intentionally absent from this committed delta.
type FieldDevicesCreated struct {
	Envelope
	FieldDevices []FieldDeviceState
}

func (FieldDevicesCreated) isCollaborationCommand() {}

// BacnetObjectUpdated represents a committed BACnet-object mutation. The
// FieldDevice IDs are already filtered to the envelope's project and let the
// version-one adapter request authoritative parent refreshes without exposing
// an unrelated project link.
type BacnetObjectUpdated struct {
	Envelope
	BacnetObjectID uuid.UUID
	FieldDeviceIDs []uuid.UUID
}

func (BacnetObjectUpdated) isCollaborationCommand() {}

// BacnetObjectCreated represents a committed BACnet object created for one
// FieldDevice. The version-one adapter asks clients to refresh that parent so
// PostgreSQL remains the authoritative source for the new child collection.
type BacnetObjectCreated struct {
	Envelope
	BacnetObjectID uuid.UUID
	FieldDeviceID  uuid.UUID
}

func (BacnetObjectCreated) isCollaborationCommand() {}

// SPSControllerUpdated represents one committed SPSController content or
// system-type update.
type SPSControllerUpdated struct {
	Envelope
	SPSControllerID uuid.UUID
}

func (SPSControllerUpdated) isCollaborationCommand() {}

type SPSControllerState struct {
	ID                uuid.UUID
	ControlCabinetID  uuid.UUID
	GADevice          *string
	DeviceName        string
	DeviceDescription *string
	DeviceLocation    *string
	IPAddress         *string
	Subnet            *string
	Gateway           *string
	VLAN              *string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// SPSControllerCreated carries exactly the committed fields used by the
// existing version-one SPS entity delta.
type SPSControllerCreated struct {
	Envelope
	SPSController SPSControllerState
}

func (SPSControllerCreated) isCollaborationCommand() {}

// SPSControllerCloned identifies the source while carrying only the committed
// copy projection needed by the existing version-one SPS entity delta.
type SPSControllerCloned struct {
	Envelope
	SourceSPSControllerID uuid.UUID
	SPSController         SPSControllerState
}

func (SPSControllerCloned) isCollaborationCommand() {}

// SPSControllerSystemTypeCloned identifies the copied child collection root
// and its owning SPSController. Version-one clients reconcile the owning
// controller because their wire contract has no system-type delta.
type SPSControllerSystemTypeCloned struct {
	Envelope
	SourceSPSControllerSystemTypeID uuid.UUID
	SPSControllerSystemTypeID       uuid.UUID
	SPSControllerID                 uuid.UUID
}

func (SPSControllerSystemTypeCloned) isCollaborationCommand() {}

// SPSControllerMoved retains explicit old/new cabinet intent even while v1
// clients receive the same authoritative targeted refresh as normal updates.
type SPSControllerMoved struct {
	Envelope
	SPSControllerID      uuid.UUID
	FromControlCabinetID uuid.UUID
	ToControlCabinetID   uuid.UUID
}

func (SPSControllerMoved) isCollaborationCommand() {}

// SPSControllerDeleted carries the pre-delete cabinet identity for future wire
// versions. The version-one Adapter requests authoritative reconciliation of
// the deleted controller ID.
type SPSControllerDeleted struct {
	Envelope
	SPSControllerID  uuid.UUID
	ControlCabinetID uuid.UUID
}

func (SPSControllerDeleted) isCollaborationCommand() {}

// ControlCabinetState contains exactly the committed cabinet fields consumed
// by the version-one collaboration delta.
type ControlCabinetState struct {
	ID               uuid.UUID
	BuildingID       uuid.UUID
	ControlCabinetNr *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// ControlCabinetUpdated represents a committed cabinet update. The v1 adapter
// emits the existing entity_delta so dependent SPS/FieldDevice views retain
// their established refresh behavior.
type ControlCabinetUpdated struct {
	Envelope
	ControlCabinet ControlCabinetState
}

func (ControlCabinetUpdated) isCollaborationCommand() {}

// ControlCabinetCreated carries the committed projection used by the existing
// version-one cabinet delta without exposing the persistence model.
type ControlCabinetCreated struct {
	Envelope
	ControlCabinet ControlCabinetState
}

func (ControlCabinetCreated) isCollaborationCommand() {}

// ControlCabinetCloned identifies the source while carrying only the committed
// copy projection needed by the existing version-one cabinet delta.
type ControlCabinetCloned struct {
	Envelope
	SourceControlCabinetID uuid.UUID
	ControlCabinet         ControlCabinetState
}

func (ControlCabinetCloned) isCollaborationCommand() {}

// ControlCabinetDeleted carries the deleted identity and former parent needed
// for committed hierarchy invalidation without exposing a persistence model.
type ControlCabinetDeleted struct {
	Envelope
	ControlCabinetID uuid.UUID
	BuildingID       uuid.UUID
}

func (ControlCabinetDeleted) isCollaborationCommand() {}

// ControlCabinetMoved preserves old/new building intent while carrying the
// same committed cabinet projection required by version-one clients.
type ControlCabinetMoved struct {
	Envelope
	ControlCabinet ControlCabinetState
	FromBuildingID uuid.UUID
	ToBuildingID   uuid.UUID
}

func (ControlCabinetMoved) isCollaborationCommand() {}
