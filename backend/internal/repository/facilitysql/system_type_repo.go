package facilitysql

import (
	"context"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/besart951/go_infra_link/backend/internal/repository/searchspec"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type systemTypeRepo struct {
	*gormbase.BaseRepository[*domainFacility.SystemType]
	db *gorm.DB
}

func NewSystemTypeRepository(db *gorm.DB) domainFacility.SystemTypeRepository {
	baseRepo := gormbase.NewBaseRepository(db,
		gormbase.TrigramSearchCallback[*domainFacility.SystemType](searchspec.SystemTypes.SearchColumns("")...),
	)
	return &systemTypeRepo{BaseRepository: baseRepo, db: db}
}

func (r *systemTypeRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SystemType], error) {
	result, err := r.BaseRepository.GetPaginatedList(ctx, params, 10)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}

func (r *systemTypeRepo) ExistsName(ctx context.Context, name string, excludeID *uuid.UUID) (bool, error) {
	query := r.db.WithContext(ctx).Model(&domainFacility.SystemType{}).
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

func (r *systemTypeRepo) ExistsOverlappingRange(ctx context.Context, numberMin, numberMax int, excludeID *uuid.UUID) (bool, error) {
	query := r.db.WithContext(ctx).Model(&domainFacility.SystemType{}).
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
