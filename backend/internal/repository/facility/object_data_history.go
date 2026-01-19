package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository"
	"github.com/google/uuid"
)

func (r *facilityRepo) GetObjectDataHistoryByIds(ids []uuid.UUID) ([]*facility.ObjectDataHistory, error) {
	var items []*facility.ObjectDataHistory
	err := r.db.
		Preload("User").
		Preload("ObjectData").
		Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *facilityRepo) CreateObjectDataHistory(entity *facility.ObjectDataHistory) error {
	return r.db.Create(entity).Error
}

func (r *facilityRepo) GetPaginatedObjectDataHistory(params domain.PaginationParams) (*domain.PaginatedList[facility.ObjectDataHistory], error) {
	return repository.Paginate[facility.ObjectDataHistory](r.db, params, []string{"action"})
}
