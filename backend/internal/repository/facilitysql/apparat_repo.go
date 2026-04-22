package facilitysql

import (
	"context"
	"errors"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type apparatRepo struct {
	*gormbase.BaseRepository[*domainFacility.Apparat]
}

func orderApparatsForFieldDeviceOptions(query *gorm.DB) *gorm.DB {
	return query.
		Order("LOWER(short_name) ASC").
		Order("LOWER(name) ASC").
		Order("id ASC")
}

func orderSystemPartsForFieldDeviceOptions(query *gorm.DB) *gorm.DB {
	return query.
		Order("LOWER(system_parts.short_name) ASC").
		Order("LOWER(system_parts.name) ASC").
		Order("system_parts.id ASC")
}

func NewApparatRepository(db *gorm.DB) domainFacility.ApparatRepository {
	searchCallback := func(query *gorm.DB, search string) *gorm.DB {
		pattern := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		return query.Where("LOWER(short_name) LIKE ? OR LOWER(name) LIKE ?", pattern, pattern)
	}

	baseRepo := gormbase.NewBaseRepository[*domainFacility.Apparat](db, searchCallback)
	return &apparatRepo{BaseRepository: baseRepo}
}

func (r *apparatRepo) GetByIds(ctx context.Context, ids []uuid.UUID) ([]*domainFacility.Apparat, error) {
	if len(ids) == 0 {
		return []*domainFacility.Apparat{}, nil
	}

	var result []*domainFacility.Apparat
	err := orderApparatsForFieldDeviceOptions(
		r.DB().WithContext(ctx).Where("id IN ?", ids),
	).Find(&result).Error
	if err != nil {
		return nil, err
	}

	// Preload SystemParts for each apparat
	for _, apparat := range result {
		if err := r.loadSortedSystemParts(ctx, apparat); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (r *apparatRepo) Create(ctx context.Context, entity *domainFacility.Apparat) error {
	return mapApparatWriteError(r.BaseRepository.Create(ctx, entity))
}

func (r *apparatRepo) Update(ctx context.Context, entity *domainFacility.Apparat) error {
	// Use GORM's Association API to replace SystemParts
	err := r.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update the entity itself
		if err := tx.Model(entity).Updates(entity).Error; err != nil {
			return err
		}

		// Replace SystemParts association
		if err := tx.Model(entity).Association("SystemParts").Replace(entity.SystemParts); err != nil {
			return err
		}

		return nil
	})
	return mapApparatWriteError(err)
}

func (r *apparatRepo) ExistsShortName(ctx context.Context, shortName string, excludeID *uuid.UUID) (bool, error) {
	return r.existsCaseInsensitive(ctx, "short_name", shortName, excludeID)
}

func (r *apparatRepo) ExistsName(ctx context.Context, name string, excludeID *uuid.UUID) (bool, error) {
	return r.existsCaseInsensitive(ctx, "name", name, excludeID)
}

func (r *apparatRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.Apparat], error) {
	result, err := r.BaseRepository.GetPaginatedList(ctx, params, 10)
	if err != nil {
		return nil, err
	}

	// Preload SystemParts for each apparat
	for _, apparat := range result.Items {
		if err := r.loadSortedSystemParts(ctx, apparat); err != nil {
			return nil, err
		}
	}

	return gormbase.DerefPaginatedList(result), nil
}

func (r *apparatRepo) loadSortedSystemParts(ctx context.Context, apparat *domainFacility.Apparat) error {
	var systemParts []*domainFacility.SystemPart
	err := orderSystemPartsForFieldDeviceOptions(
		r.DB().WithContext(ctx).
			Model(&domainFacility.SystemPart{}).
			Joins("JOIN system_part_apparats ON system_part_apparats.system_part_id = system_parts.id").
			Where("system_part_apparats.apparat_id = ?", apparat.ID),
	).Find(&systemParts).Error
	if err != nil {
		return err
	}
	apparat.SystemParts = systemParts
	return nil
}

func (r *apparatRepo) existsCaseInsensitive(ctx context.Context, column string, value string, excludeID *uuid.UUID) (bool, error) {
	query := r.DB().WithContext(ctx).
		Model(&domainFacility.Apparat{}).
		Where("LOWER("+column+") = ?", strings.ToLower(value))

	if excludeID != nil {
		query = query.Where("id <> ?", *excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func mapApparatWriteError(err error) error {
	if err == nil {
		return nil
	}
	if isDuplicateApparatWriteError(err) {
		return domain.ErrConflict
	}
	return err
}

func isDuplicateApparatWriteError(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	return false
}
