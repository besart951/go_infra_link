package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

// Unit represents a measurement unit (e.g. °C, %, Pa)
type Unit struct {
	domain.Base
	Code   string `gorm:"uniqueIndex;not null;size:30"`
	Symbol string `gorm:"not null;size:20"`
	Name   string `gorm:"not null;size:100"`
}

// AlarmType represents a technical alarm type (e.g. limit_high_low, io_monitoring)
type AlarmType struct {
	domain.Base
	Code   string           `gorm:"uniqueIndex;not null;size:80"`
	Name   string           `gorm:"not null;size:120"`
	Fields []AlarmTypeField `gorm:"foreignKey:AlarmTypeID"`
}

// AlarmField is the global field catalog
type AlarmField struct {
	domain.Base
	Key             string  `gorm:"uniqueIndex;not null;size:100"`
	Label           string  `gorm:"not null;size:150"`
	DataType        string  `gorm:"not null;size:30"`
	DefaultUnitCode *string `gorm:"size:30"`
}

// AlarmTypeField maps fields to alarm types with constraints
type AlarmTypeField struct {
	domain.Base
	AlarmTypeID      uuid.UUID   `gorm:"type:uuid;not null;uniqueIndex:idx_alarm_type_field"`
	AlarmType        *AlarmType  `gorm:"foreignKey:AlarmTypeID"`
	AlarmFieldID     uuid.UUID   `gorm:"type:uuid;not null;uniqueIndex:idx_alarm_type_field"`
	AlarmField       *AlarmField `gorm:"foreignKey:AlarmFieldID"`
	DisplayOrder     int         `gorm:"not null;default:0"`
	IsRequired       bool        `gorm:"not null;default:false"`
	IsUserEditable   bool        `gorm:"not null;default:true"`
	DefaultValueJSON *string     `gorm:"type:text"`
	ValidationJSON   *string     `gorm:"type:text"`
	DefaultUnitID    *uuid.UUID  `gorm:"type:uuid"`
	DefaultUnit      *Unit       `gorm:"foreignKey:DefaultUnitID"`
	UIGroup          *string     `gorm:"size:80"`
}
