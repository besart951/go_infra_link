package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

// AlarmDefinitionFieldOverride allows fine-tuning type defaults per definition
type AlarmDefinitionFieldOverride struct {
	domain.Base
	AlarmDefinitionID        uuid.UUID        `gorm:"type:uuid;not null;uniqueIndex:idx_alarm_def_field_override"`
	AlarmDefinition          *AlarmDefinition `gorm:"foreignKey:AlarmDefinitionID"`
	AlarmTypeFieldID         uuid.UUID        `gorm:"type:uuid;not null;uniqueIndex:idx_alarm_def_field_override"`
	AlarmTypeField           *AlarmTypeField  `gorm:"foreignKey:AlarmTypeFieldID"`
	IsRequiredOverride       *bool
	DefaultValueOverrideJSON *string    `gorm:"type:text"`
	ValidationOverrideJSON   *string    `gorm:"type:text"`
	UnitOverrideID           *uuid.UUID `gorm:"type:uuid"`
	UnitOverride             *Unit      `gorm:"foreignKey:UnitOverrideID"`
}

// BacnetObjectAlarmValue stores concrete alarm values per BacnetObject instance
type BacnetObjectAlarmValue struct {
	domain.Base
	BacnetObjectID   uuid.UUID       `gorm:"type:uuid;not null;uniqueIndex:idx_bacnet_alarm_value"`
	BacnetObject     *BacnetObject   `gorm:"foreignKey:BacnetObjectID"`
	AlarmTypeFieldID uuid.UUID       `gorm:"type:uuid;not null;uniqueIndex:idx_bacnet_alarm_value"`
	AlarmTypeField   *AlarmTypeField `gorm:"foreignKey:AlarmTypeFieldID"`
	ValueNumber      *float64        `gorm:"type:numeric(18,6)"`
	ValueInteger     *int64
	ValueBoolean     *bool
	ValueString      *string    `gorm:"type:text"`
	ValueJSON        *string    `gorm:"type:text"`
	UnitID           *uuid.UUID `gorm:"type:uuid"`
	Unit             *Unit      `gorm:"foreignKey:UnitID"`
	Source           string     `gorm:"not null;default:'user';size:20"`
}

// AlarmValueSource constants
const (
	AlarmValueSourceDefault = "default"
	AlarmValueSourceUser    = "user"
	AlarmValueSourceImport  = "import"
)
