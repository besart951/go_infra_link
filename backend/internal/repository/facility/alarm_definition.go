package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type alarmDefinitionRepo struct {
	db *gorm.DB
}

func NewAlarmDefinitionRepository(db *gorm.DB) facility.AlarmDefinitionRepository {
	return &alarmDefinitionRepo{db: db}
}

func (r *alarmDefinitionRepo) GetByIds(ids []uuid.UUID) ([]*facility.AlarmDefinition, error) {
	var items []*facility.AlarmDefinition
	err := r.db.Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *alarmDefinitionRepo) Create(entity *facility.AlarmDefinition) error {
	return r.db.Create(entity).Error
}

func (r *alarmDefinitionRepo) Update(entity *facility.AlarmDefinition) error {
	return r.db.Save(entity).Error
}

func (r *alarmDefinitionRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.db.Where("id IN ?", ids).Delete(&facility.AlarmDefinition{}).Error
}

func (r *alarmDefinitionRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[facility.AlarmDefinition], error) {
	return repository.Paginate[facility.AlarmDefinition](r.db, params, []string{"name"})
}
