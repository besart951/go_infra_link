package facility

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Building struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	IwsCode       string    `gorm:"type:varchar(4);index"`
	BuildingGroup int
	Cabinets      []ControlCabinet `gorm:"foreignKey:BuildingID"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (b *Building) BeforeCreate(tx *gorm.DB) (err error) {
	if b.ID == uuid.Nil {
		b.ID, err = uuid.NewV7()
	}
	return
}

type BuildingRepository interface {
	Create(ctx context.Context, b *Building) error
	GetByID(ctx context.Context, id uuid.UUID) (*Building, error)
	GetPaginated(ctx context.Context) ([]Building, error)
	Update(ctx context.Context, b *Building) error
	Delete(ctx context.Context, id uuid.UUID) error
}
