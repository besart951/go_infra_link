package facilityrepo

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/adapters/storage"
	"github.com/besart951/go_infra_link/backend/internal/core/domain/facility"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CabinetStorage struct {
	storage.BaseRepository[facility.ControlCabinet]
}

func NewCabinetStorage(db *gorm.DB) facility.CabinetRepository {
	return &CabinetStorage{
		BaseRepository: storage.BaseRepository[facility.ControlCabinet]{DB: db},
	}
}

func (r *CabinetStorage) FindAllByBuildingID(ctx context.Context, buildingID uuid.UUID) ([]facility.ControlCabinet, error) {
	var cabinets []facility.ControlCabinet
	err := r.DB.WithContext(ctx).Where("building_id = ?", buildingID).Find(&cabinets).Error
	return cabinets, err
}
