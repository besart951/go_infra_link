package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository"
	"github.com/google/uuid"
)

func (r *facilityRepo) GetAlarmDefinitionByIds(ids []uuid.UUID) ([]*facility.AlarmDefinition, error) {
	var items []*facility.AlarmDefinition
	err := r.db.Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *facilityRepo) CreateAlarmDefinition(entity *facility.AlarmDefinition) error {
	return r.db.Create(entity).Error
}

func (r *facilityRepo) UpdateAlarmDefinition(entity *facility.AlarmDefinition) error {
	return r.db.Save(entity).Error
}

func (r *facilityRepo) DeleteAlarmDefinitionByIds(ids []uuid.UUID) error {
	return r.db.Where("id IN ?", ids).Delete(&facility.AlarmDefinition{}).Error
}

func (r *facilityRepo) GetPaginatedAlarmDefinitions(params domain.PaginationParams) (*domain.PaginatedList[facility.AlarmDefinition], error) {
	return repository.Paginate[facility.AlarmDefinition](r.db, params, []string{"name"})
}
