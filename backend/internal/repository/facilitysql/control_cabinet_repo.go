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
	"gorm.io/gorm/clause"
)

type controlCabinetRepo struct {
	*gormbase.BaseRepository[*domainFacility.ControlCabinet]
	db *gorm.DB
}

func NewControlCabinetRepository(db *gorm.DB) domainFacility.ControlCabinetRepository {
	baseRepo := gormbase.NewBaseRepository[*domainFacility.ControlCabinet](db, nil)
	return &controlCabinetRepo{BaseRepository: baseRepo, db: db}
}

func (r *controlCabinetRepo) applySearch(ctx context.Context, query *gorm.DB, search string) *gorm.DB {
	for token := range strings.FieldsSeq(strings.TrimSpace(search)) {
		query = r.applySearchToken(ctx, query, token)
	}
	return query
}

func (r *controlCabinetRepo) applySearchToken(ctx context.Context, query *gorm.DB, token string) *gorm.DB {
	token = strings.ToLower(strings.TrimSpace(token))
	if query == nil || token == "" {
		return query
	}

	pattern := gormbase.SearchLikePattern(query, token)
	buildingIDs, err := r.findSearchBuildingIDs(ctx, pattern)
	if err != nil {
		query.AddError(err)
		return query
	}

	conditions := []string{"LOWER(control_cabinets.control_cabinet_nr) LIKE ?"}
	args := []any{pattern}
	if len(buildingIDs) > 0 {
		conditions = append(conditions, "control_cabinets.building_id IN ?")
		args = append(args, buildingIDs)
	}

	return query.Where("("+strings.Join(conditions, " OR ")+")", args...)
}

func (r *controlCabinetRepo) findSearchBuildingIDs(ctx context.Context, pattern string) ([]uuid.UUID, error) {
	columns := searchspec.Buildings.SearchColumns("")
	conditions := make([]string, 0, len(columns))
	args := make([]any, 0, len(columns))

	for _, column := range columns {
		if strings.TrimSpace(column.Expression) == "" {
			continue
		}
		conditions = append(conditions, "LOWER("+column.Expression+") LIKE ?")
		args = append(args, pattern)
	}
	if len(conditions) == 0 {
		return nil, nil
	}

	var ids []uuid.UUID
	err := r.db.WithContext(ctx).
		Model(&domainFacility.Building{}).
		Where("("+strings.Join(conditions, " OR ")+")", args...).
		Pluck("id", &ids).Error
	return ids, err
}

func (r *controlCabinetRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ControlCabinet], error) {
	query := r.db.WithContext(ctx).Model(&domainFacility.ControlCabinet{})
	if strings.TrimSpace(params.Search) == "" {
		return r.getUnfilteredPaginatedList(ctx, query, params)
	}

	query = r.applySearch(ctx, query, params.Search)

	result, err := gormbase.ExactOffsetPage[*domainFacility.ControlCabinet](
		query,
		gormbase.NormalizeOffsetPage(params, 10),
		controlCabinetCreatedAtOrder(),
	)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}

func (r *controlCabinetRepo) getUnfilteredPaginatedList(ctx context.Context, query *gorm.DB, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ControlCabinet], error) {
	page := gormbase.NormalizeOffsetPage(params, 10)

	total, ok, err := gormbase.EstimatedTableCount(ctx, r.db, "public.control_cabinets")
	if err != nil {
		return nil, err
	}
	if !ok {
		if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
			return nil, err
		}
	}

	var items []domainFacility.ControlCabinet
	if err := query.Session(&gorm.Session{}).
		Order(controlCabinetCreatedAtOrder()).
		Limit(page.Limit).
		Offset(page.Offset).
		Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.ControlCabinet]{
		Items:      items,
		Total:      total,
		Page:       page.Page,
		TotalPages: domain.CalculateTotalPages(total, page.Limit),
	}, nil
}

func (r *controlCabinetRepo) GetPaginatedListByBuildingID(ctx context.Context, buildingID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ControlCabinet], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&domainFacility.ControlCabinet{}).
		Where("building_id = ?", buildingID)

	if strings.TrimSpace(params.Search) != "" {
		query = r.applySearch(ctx, query, params.Search)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []domainFacility.ControlCabinet
	if err := query.Order(controlCabinetCreatedAtOrder()).Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.ControlCabinet]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

func controlCabinetCreatedAtOrder() clause.OrderByColumn {
	return clause.OrderByColumn{
		Column: clause.Column{Table: clause.CurrentTable, Name: "created_at"},
		Desc:   true,
	}
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
