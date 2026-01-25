package facilitysql

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type alarmDefinitionRepo struct {
	db *gorm.DB
}

func NewAlarmDefinitionRepository(db *gorm.DB) domainFacility.AlarmDefinitionRepository {
	return &alarmDefinitionRepo{db: db}
}

func (r *alarmDefinitionRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.AlarmDefinition, error) {
	var items []*domainFacility.AlarmDefinition
	err := r.db.Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *alarmDefinitionRepo) Create(entity *domainFacility.AlarmDefinition) error {
	return r.db.Create(entity).Error
}

func (r *alarmDefinitionRepo) Update(entity *domainFacility.AlarmDefinition) error {
	return r.db.Save(entity).Error
}

func (r *alarmDefinitionRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.db.Delete(&domainFacility.AlarmDefinition{}, ids).Error
}

func (r *alarmDefinitionRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.AlarmDefinition], error) {
	var items []domainFacility.AlarmDefinition
	var total int64

	db := r.db.Model(&domainFacility.AlarmDefinition{})

	if params.Search != "" {
		db = db.Where("name ILIKE ?", "%"+params.Search+"%")
	}

	err := db.Count(&total).Error
	if err != nil {
		return nil, err
	}

	offset := (params.Page - 1) * params.Limit
	err = db.Offset(offset).Limit(params.Limit).Find(&items).Error
	if err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.AlarmDefinition]{
		Items:      items,
		Total:      total,
		Page:       params.Page,
		TotalPages: domain.CalculateTotalPages(total, params.Limit),
	}, nil
}
