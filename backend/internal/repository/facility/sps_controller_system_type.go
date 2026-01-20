package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type spsControllerSystemTypeRepo struct {
	db *gorm.DB
}

func NewSPSControllerSystemTypeRepository(db *gorm.DB) facility.SPSControllerSystemTypeRepository {
	return &spsControllerSystemTypeRepo{db: db}
}

func (r *spsControllerSystemTypeRepo) GetByIds(ids []uuid.UUID) ([]*facility.SPSControllerSystemType, error) {
	var items []*facility.SPSControllerSystemType
	err := r.db.
		Preload("SPSController").
		Preload("SystemType").
		Preload("FieldDevices").
		Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *spsControllerSystemTypeRepo) Create(entity *facility.SPSControllerSystemType) error {
	return r.db.Create(entity).Error
}

func (r *spsControllerSystemTypeRepo) Update(entity *facility.SPSControllerSystemType) error {
	return r.db.Save(entity).Error
}

func (r *spsControllerSystemTypeRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.db.Where("id IN ?", ids).Delete(&facility.SPSControllerSystemType{}).Error
}

func (r *spsControllerSystemTypeRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[facility.SPSControllerSystemType], error) {
	return repository.Paginate[facility.SPSControllerSystemType](r.db, params, []string{"document_name"})
}
