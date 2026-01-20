package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type specificationRepo struct {
	db *gorm.DB
}

func NewSpecificationRepository(db *gorm.DB) facility.SpecificationRepository {
	return &specificationRepo{db: db}
}

func (r *specificationRepo) GetByIds(ids []uuid.UUID) ([]*facility.Specification, error) {
	var items []*facility.Specification
	err := r.db.Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *specificationRepo) Create(entity *facility.Specification) error {
	return r.db.Create(entity).Error
}

func (r *specificationRepo) Update(entity *facility.Specification) error {
	return r.db.Save(entity).Error
}

func (r *specificationRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.db.Where("id IN ?", ids).Delete(&facility.Specification{}).Error
}

func (r *specificationRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[facility.Specification], error) {
	return repository.Paginate[facility.Specification](r.db, params, []string{"specification_supplier", "specification_brand"})
}
