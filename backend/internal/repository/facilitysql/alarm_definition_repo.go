package facilitysql

import (
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
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

func (r *alarmDefinitionRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.AlarmDefinition], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 10)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}
