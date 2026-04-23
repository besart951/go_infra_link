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
	TextIndividuell           *string
	SPSControllerSystemTypeID uuid.UUID               `gorm:"type:uuid;not null;index"`
	SPSControllerSystemType   SPSControllerSystemType `gorm:"foreignKey:SPSControllerSystemTypeID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	SystemPartID              uuid.UUID               `gorm:"type:uuid;not null;index"`
	SystemPart                SystemPart              `gorm:"foreignKey:SystemPartID"`
	SpecificationID           *uuid.UUID              `gorm:"type:uuid;index"`
	Specification             *Specification          `gorm:"foreignKey:SpecificationID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
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
	ProjectID                 *uuid.UUID
	ProjectIDs                []uuid.UUID
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

// BulkFieldDeviceUpdate represents a single field device update in a bulk operation
type BulkFieldDeviceUpdate struct {
	ID                 uuid.UUID
	BMK                *string
	HasBMK             bool
	Description        *string
	HasDescription     bool
	TextIndividuell    *string
	HasTextIndividuell bool
	ApparatNr          *int
	ApparatID          *uuid.UUID
	SystemPartID       *uuid.UUID
	Specification      *SpecificationPatch
	BacnetObjects      *[]BacnetObjectPatch
}

func (u BulkFieldDeviceUpdate) HasBMKUpdate() bool {
	return u.HasBMK || u.BMK != nil
}

func (u BulkFieldDeviceUpdate) HasDescriptionUpdate() bool {
	return u.HasDescription || u.Description != nil
}

func (u BulkFieldDeviceUpdate) HasTextIndividuellUpdate() bool {
	return u.HasTextIndividuell || u.TextIndividuell != nil
}

type SpecificationPatch struct {
	SpecificationSupplier                        *string
	HasSpecificationSupplier                     bool
	SpecificationBrand                           *string
	HasSpecificationBrand                        bool
	SpecificationType                            *string
	HasSpecificationType                         bool
	AdditionalInfoMotorValve                     *string
	HasAdditionalInfoMotorValve                  bool
	AdditionalInfoSize                           *int
	HasAdditionalInfoSize                        bool
	AdditionalInformationInstallationLocation    *string
	HasAdditionalInformationInstallationLocation bool
	ElectricalConnectionPH                       *int
	HasElectricalConnectionPH                    bool
	ElectricalConnectionACDC                     *string
	HasElectricalConnectionACDC                  bool
	ElectricalConnectionAmperage                 *float64
	HasElectricalConnectionAmperage              bool
	ElectricalConnectionPower                    *float64
	HasElectricalConnectionPower                 bool
	ElectricalConnectionRotation                 *int
	HasElectricalConnectionRotation              bool
}

func (p *SpecificationPatch) HasChanges() bool {
	if p == nil {
		return false
	}
	return p.HasSpecificationSupplier ||
		p.HasSpecificationBrand ||
		p.HasSpecificationType ||
		p.HasAdditionalInfoMotorValve ||
		p.HasAdditionalInfoSize ||
		p.HasAdditionalInformationInstallationLocation ||
		p.HasElectricalConnectionPH ||
		p.HasElectricalConnectionACDC ||
		p.HasElectricalConnectionAmperage ||
		p.HasElectricalConnectionPower ||
		p.HasElectricalConnectionRotation
}

func (p *SpecificationPatch) HasNonNilValues() bool {
	if p == nil {
		return false
	}
	return p.SpecificationSupplier != nil ||
		p.SpecificationBrand != nil ||
		p.SpecificationType != nil ||
		p.AdditionalInfoMotorValve != nil ||
		p.AdditionalInfoSize != nil ||
		p.AdditionalInformationInstallationLocation != nil ||
		p.ElectricalConnectionPH != nil ||
		p.ElectricalConnectionACDC != nil ||
		p.ElectricalConnectionAmperage != nil ||
		p.ElectricalConnectionPower != nil ||
		p.ElectricalConnectionRotation != nil
}

// BulkOperationResultItem represents the result of a single item in a bulk operation
type BulkOperationResultItem struct {
	ID      uuid.UUID
	Success bool
	Error   string
	Fields  map[string]string // Per-field validation errors (e.g., "bacnet_objects.0.text_fix" -> "message")
}

// BulkOperationResult represents the result of a bulk operation
type BulkOperationResult struct {
	Results      []BulkOperationResultItem
	TotalCount   int
	SuccessCount int
	FailureCount int
}
