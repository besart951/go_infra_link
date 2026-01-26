package facilitysql

import (
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type alarmDefinitionRepo struct {
	*gormbase.BaseRepository[*domainFacility.AlarmDefinition]
}

func NewAlarmDefinitionRepository(db *gorm.DB) domainFacility.AlarmDefinitionRepository {
	searchCallback := func(query *gorm.DB, search string) *gorm.DB {
		pattern := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		return query.Where("LOWER(name) LIKE ?", pattern)
	}

	baseRepo := gormbase.NewBaseRepository[*domainFacility.AlarmDefinition](db, searchCallback)
	return &alarmDefinitionRepo{BaseRepository: baseRepo}
}

func (r *alarmDefinitionRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.AlarmDefinition, error) {
	return r.BaseRepository.GetByIds(ids)
}

func (r *alarmDefinitionRepo) Create(entity *domainFacility.AlarmDefinition) error {
	return r.BaseRepository.Create(entity)
}

func (r *alarmDefinitionRepo) Update(entity *domainFacility.AlarmDefinition) error {
	return r.BaseRepository.Update(entity)
}

func (r *alarmDefinitionRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.BaseRepository.DeleteByIds(ids)
}

func (r *alarmDefinitionRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.AlarmDefinition], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 10)
	if err != nil {
		return nil, err
	}

	// Convert []*AlarmDefinition to []AlarmDefinition for the interface
	items := make([]domainFacility.AlarmDefinition, len(result.Items))
	for i, item := range result.Items {
		items[i] = *item
	}

	return &domain.PaginatedList[domainFacility.AlarmDefinition]{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}, nil
}
