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
	BacnetObjects []*BacnetObject `gorm:"many2many:object_data_bacnet_objects;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
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
	Description   *string
	Version       *string
	IsActive      *bool
	ProjectID     *uuid.UUID
	ApparatIDs    *[]uuid.UUID
	BacnetObjects *[]BacnetObject
}

// ActivateForProject applies the local ownership and activation rules for an
// ObjectData template. Project existence and actor authorization remain
// application concerns because they require external state.
func (o *ObjectData) ActivateForProject(projectID uuid.UUID) error {
	if o == nil || projectID == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	if o.ProjectID != nil && *o.ProjectID != projectID {
		return domain.ErrConflict
	}
	projectIDCopy := projectID
	o.ProjectID = &projectIDCopy
	o.IsActive = true
	return nil
}

// DeactivateForProject preserves the established HTTP semantics: removing an
// ObjectData template from a project marks it inactive but retains ProjectID.
func (o *ObjectData) DeactivateForProject(projectID uuid.UUID) error {
	if o == nil || projectID == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	if o.ProjectID == nil || *o.ProjectID != projectID {
		return domain.ErrNotFound
	}
	o.IsActive = false
	return nil
}
