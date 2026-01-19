package facilityrepo

import (
	"github.com/besart951/go_infra_link/backend/internal/adapters/storage"
	"github.com/besart951/go_infra_link/backend/internal/core/domain/facility"
	"gorm.io/gorm"
)

type BuildingStorage struct {
	storage.BaseRepository[facility.Building]
}

func NewBuildingStorage(db *gorm.DB) facility.BuildingRepository {
	return &BuildingStorage{
		BaseRepository: storage.BaseRepository[facility.Building]{DB: db},
	}
}
