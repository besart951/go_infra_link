package facilitysql

import (
	"time"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainObjectData "github.com/besart951/go_infra_link/backend/internal/domain/facility/objectdata"
	"github.com/google/uuid"
)

type BacnetObjectTemplateRecord struct {
	ID                  uuid.UUID `gorm:"type:uuid;primaryKey;uniqueIndex:idx_bacnet_template_owner_id"`
	CreatedAt           time.Time `gorm:"not null"`
	UpdatedAt           time.Time `gorm:"not null"`
	Version             uint64    `gorm:"not null;default:1"`
	ObjectDataID        uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_bacnet_template_owner_id;uniqueIndex:idx_bacnet_template_software;index"`
	TextFix             string    `gorm:"not null"`
	Description         *string
	GMSVisible          bool `gorm:"not null;default:false"`
	Optional            bool `gorm:"not null;default:false"`
	TextIndividual      *string
	SoftwareType        domainFacility.BacnetSoftwareType `gorm:"type:varchar(50);not null;uniqueIndex:idx_bacnet_template_software"`
	SoftwareNumber      uint16                            `gorm:"not null;uniqueIndex:idx_bacnet_template_software"`
	HardwareType        domainFacility.BacnetHardwareType `gorm:"type:varchar(50)"`
	HardwareQuantity    uint8
	SoftwareReferenceID *uuid.UUID `gorm:"type:uuid"`
	StateTextID         *uuid.UUID `gorm:"type:uuid;index"`
	NotificationClassID *uuid.UUID `gorm:"type:uuid;index"`
	AlarmTypeID         *uuid.UUID `gorm:"type:uuid;index"`
}

func (BacnetObjectTemplateRecord) TableName() string { return "bacnet_object_templates" }

type BacnetObjectTemplateAlarmValueRecord struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt        time.Time `gorm:"not null"`
	UpdatedAt        time.Time `gorm:"not null"`
	Version          uint64    `gorm:"not null;default:1"`
	TemplateID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_bacnet_template_alarm_field"`
	AlarmTypeFieldID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_bacnet_template_alarm_field"`
	ValueNumber      *float64  `gorm:"type:numeric(18,6)"`
	ValueInteger     *int64
	ValueBoolean     *bool
	ValueString      *string `gorm:"type:text"`
	ValueJSON        *string `gorm:"type:text"`
	UnitID           *uuid.UUID
	Source           string `gorm:"not null;default:'user';size:20"`
}

func (BacnetObjectTemplateAlarmValueRecord) TableName() string {
	return "bacnet_object_template_alarm_values"
}

func templateRecordToDomain(record BacnetObjectTemplateRecord) domainObjectData.BacnetObjectTemplate {
	return domainObjectData.BacnetObjectTemplate{
		ID: record.ID, CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt, Version: record.Version,
		ObjectDataID: record.ObjectDataID, TextFix: record.TextFix, Description: record.Description,
		GMSVisible: record.GMSVisible, Optional: record.Optional, TextIndividual: record.TextIndividual,
		SoftwareType: record.SoftwareType, SoftwareNumber: record.SoftwareNumber,
		HardwareType: record.HardwareType, HardwareQuantity: record.HardwareQuantity,
		SoftwareReferenceID: record.SoftwareReferenceID, StateTextID: record.StateTextID,
		NotificationClassID: record.NotificationClassID, AlarmTypeID: record.AlarmTypeID,
	}
}

func templateRecordFromDomain(template domainObjectData.BacnetObjectTemplate) BacnetObjectTemplateRecord {
	return BacnetObjectTemplateRecord{
		ID: template.ID, CreatedAt: template.CreatedAt, UpdatedAt: template.UpdatedAt, Version: template.Version,
		ObjectDataID: template.ObjectDataID, TextFix: template.TextFix, Description: template.Description,
		GMSVisible: template.GMSVisible, Optional: template.Optional, TextIndividual: template.TextIndividual,
		SoftwareType: template.SoftwareType, SoftwareNumber: template.SoftwareNumber,
		HardwareType: template.HardwareType, HardwareQuantity: template.HardwareQuantity,
		SoftwareReferenceID: template.SoftwareReferenceID, StateTextID: template.StateTextID,
		NotificationClassID: template.NotificationClassID, AlarmTypeID: template.AlarmTypeID,
	}
}
