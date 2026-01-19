package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type ObjectData struct {
	domain.Base
	Description string     `gorm:"size:250;uniqueIndex:idx_obj_data_proj_desc"`
	Version     string     `gorm:"size:50;default:'1.0.0'"`
	IsActive    bool       `gorm:"default:true;index"`
	ProjectID   *uuid.UUID `gorm:"uniqueIndex:idx_obj_data_proj_desc"`

	Project *domain.Project `gorm:"foreignKey:ProjectID"`

	BacnetObjects []*BacnetObject `gorm:"many2many:object_data_Bacnet_objects;"`
	Apparats      []*Apparat      `gorm:"many2many:object_data_apparats;"`
}
