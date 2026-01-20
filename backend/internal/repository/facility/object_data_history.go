package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type objectDataHistoryRepo struct {
	db *gorm.DB
}

func NewObjectDataHistoryRepository(db *gorm.DB) facility.ObjectDataHistoryRepository {
	return &objectDataHistoryRepo{db: db}
}

func (r *objectDataHistoryRepo) GetByIds(ids []uuid.UUID) ([]*facility.ObjectDataHistory, error) {
	var items []*facility.ObjectDataHistory
	err := r.db.
		Preload("User").
		Preload("ObjectData").
		Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *objectDataHistoryRepo) Create(entity *facility.ObjectDataHistory) error {
	return r.db.Create(entity).Error
}

func (r *objectDataHistoryRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[facility.ObjectDataHistory], error) {
	return repository.Paginate[facility.ObjectDataHistory](r.db, params, []string{"action"})
}
