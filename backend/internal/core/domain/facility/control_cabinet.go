package facility

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ControlCabinet struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey;"`
	BuildingID       uuid.UUID  `gorm:"type:uuid;not null;index"`
	Building         Building   `gorm:"constraint:OnDelete:CASCADE;"`
	ProjectID        *uuid.UUID `gorm:"type:uuid;index"`
	ControlCabinetNr *string    `gorm:"type:varchar(11);index"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (c *ControlCabinet) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID, err = uuid.NewV7()
	}
	return
}

type CabinetRepository interface {
	Create(ctx context.Context, c *ControlCabinet) error
	GetByID(ctx context.Context, id uuid.UUID) (*ControlCabinet, error)
	FindAllByBuildingID(ctx context.Context, buildingID uuid.UUID) ([]ControlCabinet, error)
}
