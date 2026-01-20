package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type objectDataRepo struct {
	db *gorm.DB
}

func NewObjectDataRepository(db *gorm.DB) facility.ObjectDataRepository {
	return &objectDataRepo{db: db}
}

func (r *objectDataRepo) GetByIds(ids []uuid.UUID) ([]*facility.ObjectData, error) {
	var items []*facility.ObjectData
	err := r.db.
		Preload("Project").
		Preload("BacnetObjects").
		Preload("Apparats").
		Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *objectDataRepo) Create(entity *facility.ObjectData) error {
	return r.db.Create(entity).Error
}

func (r *objectDataRepo) Update(entity *facility.ObjectData) error {
	return r.db.Save(entity).Error
}

func (r *objectDataRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.db.Where("id IN ?", ids).Delete(&facility.ObjectData{}).Error
}

func (r *objectDataRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[facility.ObjectData], error) {
	return repository.Paginate[facility.ObjectData](r.db, params, []string{"description", "version"})
}
