package facilitysql

import (
	"fmt"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type alarmDefinitionRepo struct {
	*gormbase.BaseRepository[*domainFacility.AlarmDefinition]
	db *gorm.DB
}

func NewAlarmDefinitionRepository(db *gorm.DB) domainFacility.AlarmDefinitionRepository {
	searchCallback := func(query *gorm.DB, search string) *gorm.DB {
		pattern := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		return query.Where("LOWER(name) LIKE ?", pattern)
	}

	baseRepo := gormbase.NewBaseRepository[*domainFacility.AlarmDefinition](db, searchCallback)
	return &alarmDefinitionRepo{BaseRepository: baseRepo, db: db}
}

func (r *alarmDefinitionRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.AlarmDefinition], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 10)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}

func (r *alarmDefinitionRepo) FindOrCreateTemplateByAlarmTypeID(alarmTypeID uuid.UUID) (*domainFacility.AlarmDefinition, error) {
	var existing domainFacility.AlarmDefinition
	err := r.db.
		Where("alarm_type_id = ?", alarmTypeID).
		Where("scope = ?", "template").
		Where("is_active = ?", true).
		Order("created_at ASC").
		First(&existing).Error
	if err == nil {
		return &existing, nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	created := &domainFacility.AlarmDefinition{
		Name:        fmt.Sprintf("AUTO_TEMPLATE_%s", strings.ToUpper(alarmTypeID.String()[:8])),
		AlarmTypeID: &alarmTypeID,
		IsActive:    true,
		Scope:       "template",
	}

	if err := r.BaseRepository.Create(created); err != nil {
		return nil, err
	}

	return created, nil
}
