package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type systemTypeRepo struct {
	db *gorm.DB
}

func NewSystemTypeRepository(db *gorm.DB) facility.SystemTypeRepository {
	return &systemTypeRepo{db: db}
}

func (r *systemTypeRepo) GetByIds(ids []uuid.UUID) ([]*facility.SystemType, error) {
	var items []*facility.SystemType
	err := r.db.Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *systemTypeRepo) Create(entity *facility.SystemType) error {
	return r.db.Create(entity).Error
}

func (r *systemTypeRepo) Update(entity *facility.SystemType) error {
	return r.db.Save(entity).Error
}

func (r *systemTypeRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.db.Where("id IN ?", ids).Delete(&facility.SystemType{}).Error
}

func (r *systemTypeRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[facility.SystemType], error) {
	return repository.Paginate[facility.SystemType](r.db, params, []string{"name"})
}
