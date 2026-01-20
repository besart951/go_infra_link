package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type systemPartRepo struct {
	db *gorm.DB
}

func NewSystemPartRepository(db *gorm.DB) facility.SystemPartRepository {
	return &systemPartRepo{db: db}
}

func (r *systemPartRepo) GetByIds(ids []uuid.UUID) ([]*facility.SystemPart, error) {
	var items []*facility.SystemPart
	err := r.db.Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *systemPartRepo) Create(entity *facility.SystemPart) error {
	return r.db.Create(entity).Error
}

func (r *systemPartRepo) Update(entity *facility.SystemPart) error {
	return r.db.Save(entity).Error
}

func (r *systemPartRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.db.Where("id IN ?", ids).Delete(&facility.SystemPart{}).Error
}

func (r *systemPartRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[facility.SystemPart], error) {
	return repository.Paginate[facility.SystemPart](r.db, params, []string{"name", "short_name"})
}
