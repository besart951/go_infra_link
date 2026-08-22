package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type ObjectData struct {
	domain.Base
	Description string      `gorm:"not null;uniqueIndex:idx_object_data_project_description"`
	Version     string      `gorm:"not null;column:obj_version"`
	IsActive    bool        `gorm:"default:true"`
	ProjectID   *uuid.UUID  `gorm:"type:uuid;index;uniqueIndex:idx_object_data_project_description"`
	Project     *ProjectRef `gorm:"foreignKey:ProjectID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	// BacnetObjects is a transport view populated from bacnet_object_templates.
	// Templates are persisted by the ObjectData capability, never as a GORM
	// many-to-many association.
	BacnetObjects []*BacnetObject `gorm:"-:all"`
	Apparats      []*Apparat      `gorm:"many2many:object_data_apparats;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type ObjectDataTemplateCreate struct {
	Description   string
	Version       string
	IsActive      *bool
	ProjectID     *uuid.UUID
	ApparatIDs    []uuid.UUID
	BacnetObjects []BacnetObject
}

type ObjectDataTemplateUpdate struct {
	BaseVersion   uint64
	Description   *string
	Version       *string
	IsActive      *bool
	ProjectID     *uuid.UUID
	ApparatIDs    *[]uuid.UUID
	BacnetObjects *[]BacnetObject
}
