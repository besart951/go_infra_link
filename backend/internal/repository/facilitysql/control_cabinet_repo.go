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

type controlCabinetRepo struct {
	*gormbase.BaseRepository[*domainFacility.ControlCabinet]
	db *gorm.DB
}

func NewControlCabinetRepository(db *gorm.DB) domainFacility.ControlCabinetRepository {
	baseRepo := gormbase.NewBaseRepository[*domainFacility.ControlCabinet](db, applyControlCabinetSearch)
	return &controlCabinetRepo{BaseRepository: baseRepo, db: db}
}

func applyControlCabinetSearch(query *gorm.DB, search string) *gorm.DB {
	tokens := strings.Fields(strings.ToLower(strings.TrimSpace(search)))
	for _, token := range tokens {
		pattern := "%" + token + "%"
		query = query.Where(`
			LOWER(control_cabinet_nr) LIKE ?
			OR building_id IN (
				SELECT id
				FROM buildings
				WHERE LOWER(iws_code) LIKE ?
					OR CAST(building_group AS TEXT) LIKE ?
					OR LOWER(iws_code || '-' || CAST(building_group AS TEXT)) LIKE ?
			)
		`, pattern, pattern, pattern, pattern)
	}
	return query
}

func (r *controlCabinetRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ControlCabinet], error) {
	result, err := r.BaseRepository.GetPaginatedList(ctx, params, 10)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}

func (r *controlCabinetRepo) GetPaginatedListByBuildingID(ctx context.Context, buildingID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ControlCabinet], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&domainFacility.ControlCabinet{}).
		Where("building_id = ?", buildingID)

	if strings.TrimSpace(params.Search) != "" {
		query = applyControlCabinetSearch(query, params.Search)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []domainFacility.ControlCabinet
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.ControlCabinet]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

func (r *controlCabinetRepo) ExistsControlCabinetNr(ctx context.Context, buildingID uuid.UUID, controlCabinetNr string, excludeID *uuid.UUID) (bool, error) {
	query := r.db.WithContext(ctx).Model(&domainFacility.ControlCabinet{}).
		Where("building_id = ?", buildingID).
		Where("LOWER(control_cabinet_nr) = ?", strings.ToLower(strings.TrimSpace(controlCabinetNr)))

	if excludeID != nil {
		query = query.Where("id <> ?", *excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *controlCabinetRepo) GetIDsByBuildingID(ctx context.Context, buildingID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).Model(&domainFacility.ControlCabinet{}).
		Where("building_id = ?", buildingID).
		Pluck("id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *controlCabinetRepo) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Delete(&domainFacility.ControlCabinet{}).Error
}
