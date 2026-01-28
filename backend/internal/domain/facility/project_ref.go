package facility

import "github.com/google/uuid"

// ProjectRef is a minimal reference to the projects table for FK constraints.
type ProjectRef struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`
}

func (ProjectRef) TableName() string {
	return "projects"
}
