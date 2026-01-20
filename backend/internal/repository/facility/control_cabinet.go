package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type controlCabinetRepo struct {
	db *gorm.DB
}

func NewControlCabinetRepository(db *gorm.DB) facility.ControlCabinetRepository {
	return &controlCabinetRepo{db: db}
}

func (r *controlCabinetRepo) GetByIds(ids []uuid.UUID) ([]*facility.ControlCabinet, error) {
	var items []*facility.ControlCabinet
	err := r.db.
		Preload("Building").
		Preload("SPSControllers").
		Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *controlCabinetRepo) Create(entity *facility.ControlCabinet) error {
	return r.db.Create(entity).Error
}

func (r *controlCabinetRepo) Update(entity *facility.ControlCabinet) error {
	return r.db.Save(entity).Error
}

func (r *controlCabinetRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.db.Where("id IN ?", ids).Delete(&facility.ControlCabinet{}).Error
}

func (r *controlCabinetRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[facility.ControlCabinet], error) {
	return repository.Paginate[facility.ControlCabinet](r.db, params, []string{"control_cabinet_nr"})
}
