package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type buildingRepo struct {
	db *gorm.DB
}

func NewBuildingRepository(db *gorm.DB) facility.BuildingRepository {
	return &buildingRepo{db: db}
}

func (r *buildingRepo) GetByIds(ids []uuid.UUID) ([]*facility.Building, error) {
	var buildings []*facility.Building
	err := r.db.
		Preload("ControlCabinets").
		Preload("ControlCabinets.SPSControllers").
		Where("id IN ?", ids).
		Find(&buildings).Error
	return buildings, err
}

func (r *buildingRepo) Create(entity *facility.Building) error {
	return r.db.Create(entity).Error
}

func (r *buildingRepo) Update(entity *facility.Building) error {
	return r.db.Save(entity).Error
}

func (r *buildingRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.db.Where("id IN ?", ids).Delete(&facility.Building{}).Error
}

func (r *buildingRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[facility.Building], error) {
	return repository.Paginate[facility.Building](r.db, params, []string{"iws_code"})
}
