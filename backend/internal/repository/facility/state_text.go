package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type stateTextRepo struct {
	db *gorm.DB
}

func NewStateTextRepository(db *gorm.DB) facility.StateTextRepository {
	return &stateTextRepo{db: db}
}

func (r *stateTextRepo) GetByIds(ids []uuid.UUID) ([]*facility.StateText, error) {
	var items []*facility.StateText
	err := r.db.Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *stateTextRepo) Create(entity *facility.StateText) error {
	return r.db.Create(entity).Error
}

func (r *stateTextRepo) Update(entity *facility.StateText) error {
	return r.db.Save(entity).Error
}

func (r *stateTextRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.db.Where("id IN ?", ids).Delete(&facility.StateText{}).Error
}

func (r *stateTextRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[facility.StateText], error) {
	return repository.Paginate[facility.StateText](r.db, params, []string{"state_text_1", "state_text_2"})
}
