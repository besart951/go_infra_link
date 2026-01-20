package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type apparatRepo struct {
	db *gorm.DB
}

func NewApparatRepository(db *gorm.DB) facility.ApparatRepository {
	return &apparatRepo{db: db}
}

func (r *apparatRepo) GetByIds(ids []uuid.UUID) ([]*facility.Apparat, error) {
	var items []*facility.Apparat
	err := r.db.Preload("SystemParts").Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *apparatRepo) Create(entity *facility.Apparat) error {
	return r.db.Create(entity).Error
}

func (r *apparatRepo) Update(entity *facility.Apparat) error {
	return r.db.Save(entity).Error
}

func (r *apparatRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.db.Where("id IN ?", ids).Delete(&facility.Apparat{}).Error
}

func (r *apparatRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[facility.Apparat], error) {
	return repository.Paginate[facility.Apparat](r.db, params, []string{"name", "short_name"})
}
