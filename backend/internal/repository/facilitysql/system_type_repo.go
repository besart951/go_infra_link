package facilitysql

import (
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type systemTypeRepo struct {
	*gormbase.BaseRepository[*domainFacility.SystemType]
	db *gorm.DB
}

func NewSystemTypeRepository(db *gorm.DB) domainFacility.SystemTypeRepository {
	searchCallback := func(query *gorm.DB, search string) *gorm.DB {
		pattern := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		return query.Where("LOWER(name) LIKE ?", pattern)
	}

	baseRepo := gormbase.NewBaseRepository[*domainFacility.SystemType](db, searchCallback)
	return &systemTypeRepo{BaseRepository: baseRepo, db: db}
}

func (r *systemTypeRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SystemType], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 10)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}

func (r *systemTypeRepo) ExistsName(name string, excludeID *uuid.UUID) (bool, error) {
	query := r.db.Model(&domainFacility.SystemType{}).
		Where("deleted_at IS NULL").
		Where("LOWER(name) = ?", strings.ToLower(strings.TrimSpace(name)))

	if excludeID != nil {
		query = query.Where("id <> ?", *excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *systemTypeRepo) ExistsOverlappingRange(numberMin, numberMax int, excludeID *uuid.UUID) (bool, error) {
	query := r.db.Model(&domainFacility.SystemType{}).
		Where("deleted_at IS NULL").
		Where("number_min <= ?", numberMax).
		Where("number_max >= ?", numberMin)

	if excludeID != nil {
		query = query.Where("id <> ?", *excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
