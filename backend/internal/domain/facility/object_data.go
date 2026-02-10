package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type ObjectData struct {
	domain.Base
	Description   string          `gorm:"not null;uniqueIndex:idx_object_data_project_description"`
	Version       string          `gorm:"not null;column:obj_version"`
	IsActive      bool            `gorm:"default:true"`
	ProjectID     *uuid.UUID      `gorm:"type:uuid;index;uniqueIndex:idx_object_data_project_description"`
	Project       *ProjectRef     `gorm:"foreignKey:ProjectID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	BacnetObjects []*BacnetObject `gorm:"many2many:object_data_bacnet_objects;"`
	Apparats      []*Apparat      `gorm:"many2many:object_data_apparats;"`
}
