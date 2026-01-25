package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type ObjectData struct {
	domain.Base
	Description   string           `gorm:"not null"`
	Version       string           `gorm:"not null;column:obj_version"`
	IsActive      bool             `gorm:"default:true"`
	ProjectID     *uuid.UUID       `gorm:"type:uuid;index"`
	Project       *project.Project `gorm:"foreignKey:ProjectID"`
	BacnetObjects []*BacnetObject  `gorm:"many2many:object_data_bacnet_objects;"`
	Apparats      []*Apparat       `gorm:"many2many:object_data_apparats;"`
}
