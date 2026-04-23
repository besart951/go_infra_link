package facilitysql

import (
	"context"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type systemPartRepo struct {
	*gormbase.BaseRepository[*domainFacility.SystemPart]
}

func NewSystemPartRepository(db *gorm.DB) domainFacility.SystemPartRepository {
	searchCallback := func(query *gorm.DB, search string) *gorm.DB {
		pattern := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		return query.Where("LOWER(short_name) LIKE ? OR LOWER(name) LIKE ?", pattern, pattern)
	}

	baseRepo := gormbase.NewBaseRepository[*domainFacility.SystemPart](db, searchCallback)
	return &systemPartRepo{BaseRepository: baseRepo}
}

func (r *systemPartRepo) GetByIds(ctx context.Context, ids []uuid.UUID) ([]*domainFacility.SystemPart, error) {
	result, err := r.BaseRepository.GetByIds(ctx, ids)
	if err != nil {
		return nil, err
	}

	// Preload Apparats for each system part (many2many)
	for _, systemPart := range result {
		if err := r.DB().WithContext(ctx).Model(systemPart).Association("Apparats").Find(&systemPart.Apparats); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (r *systemPartRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SystemPart], error) {
	result, err := r.BaseRepository.GetPaginatedList(ctx, params, 10)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}

func (r *systemPartRepo) ExistsShortName(ctx context.Context, shortName string, excludeID *uuid.UUID) (bool, error) {
	return r.existsCaseInsensitive(ctx, "short_name", shortName, excludeID)
}

func (r *systemPartRepo) ExistsName(ctx context.Context, name string, excludeID *uuid.UUID) (bool, error) {
	return r.existsCaseInsensitive(ctx, "name", name, excludeID)
}

func (r *systemPartRepo) existsCaseInsensitive(ctx context.Context, column string, value string, excludeID *uuid.UUID) (bool, error) {
	query := r.DB().WithContext(ctx).
		Model(&domainFacility.SystemPart{}).
		Where("LOWER("+column+") = ?", strings.ToLower(strings.TrimSpace(value)))

	if excludeID != nil {
		query = query.Where("id <> ?", *excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}
