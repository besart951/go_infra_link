package repository

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type facilityRepo struct {
	db *gorm.DB
}

func NewFacilityRepository(db *gorm.DB) domain.FacilityRepository {
	return &facilityRepo{db: db}
}

func (r *facilityRepo) GetBuildingByIds(ids []uuid.UUID) ([]*domain.Building, error) {
	var buildings []*domain.Building
	// Eager Loading der tief verschachtelten Struktur
	err := r.db.
		Preload("ControlCabinets").
		Preload("ControlCabinets.SPSControllers").
		Where("id IN ?", ids).
		Find(&buildings).Error
	return buildings, err
}

func (r *facilityRepo) CreateBuilding(entity *domain.Building) error {
	return r.db.Create(entity).Error
}

func (r *facilityRepo) UpdateBuilding(entity *domain.Building) error {
	return r.db.Save(entity).Error
}

func (r *facilityRepo) DeleteBuildingByIds(ids []uuid.UUID) error {
	return r.db.Where("id IN ?", ids).Delete(&domain.Building{}).Error
}

func (r *facilityRepo) GetPaginatedBuildings(params domain.PaginationParams) (*domain.PaginatedList[domain.Building], error) {
	return paginate[domain.Building](r.db, params, []string{"iws_code"})
}
