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

type spsControllerSystemTypeRepo struct {
	*gormbase.BaseRepository[*domainFacility.SPSControllerSystemType]
	db *gorm.DB
}

func (r *spsControllerSystemTypeRepo) querySystemTypes(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).
		Model(&domainFacility.SPSControllerSystemType{}).
		Omit("FieldDevicesCount")
}

func applySPSControllerSystemTypeSearch(query *gorm.DB, search string) *gorm.DB {
	columns := searchspec.SPSControllerSystemTypes.SearchColumns("sps_controller_system_types.")
	columns = append(columns, searchspec.SPSControllers.NamedSearchColumns("sps_controllers.", "device_name")...)
	columns = append(columns, searchspec.SystemTypes.SearchColumns("system_types.")...)

	return query.Joins("LEFT JOIN sps_controllers ON sps_controllers.id = sps_controller_system_types.sps_controller_id").
		Joins("LEFT JOIN system_types ON system_types.id = sps_controller_system_types.system_type_id").
		Scopes(func(db *gorm.DB) *gorm.DB {
			return gormbase.ApplyTrigramSearch(db, search, columns...)
		})
}

func NewSPSControllerSystemTypeRepository(db *gorm.DB) domainFacility.SPSControllerSystemTypeStore {
	searchCallback := func(query *gorm.DB, search string) *gorm.DB {
		return applySPSControllerSystemTypeSearch(query, search)
	}

	baseRepo := gormbase.NewBaseRepository[*domainFacility.SPSControllerSystemType](db, searchCallback)
	return &spsControllerSystemTypeRepo{
		BaseRepository: baseRepo,
		db:             db,
	}
}

func (r *spsControllerSystemTypeRepo) GetByIds(ctx context.Context, ids []uuid.UUID) ([]*domainFacility.SPSControllerSystemType, error) {
	if len(ids) == 0 {
		return []*domainFacility.SPSControllerSystemType{}, nil
	}
	var items []*domainFacility.SPSControllerSystemType
	err := r.querySystemTypes(ctx).
		Where("sps_controller_system_types.id IN ?", ids).
		Preload("SPSController").
		Preload("SystemType").
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, r.attachFieldDeviceCounts(ctx, items)
}

func (r *spsControllerSystemTypeRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SPSControllerSystemType], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.querySystemTypes(ctx)

	if strings.TrimSpace(params.Search) != "" {
		query = applySPSControllerSystemTypeSearch(query, params.Search)
	}

	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, err
	}

	var items []domainFacility.SPSControllerSystemType
	if err := query.Session(&gorm.Session{}).Preload("SPSController").Preload("SystemType").
		Order("sps_controller_system_types.created_at DESC").
		Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}
	if err := r.attachFieldDeviceCountsToValues(ctx, items); err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.SPSControllerSystemType]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

func (r *spsControllerSystemTypeRepo) GetPaginatedListBySPSControllerID(ctx context.Context, spsControllerID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SPSControllerSystemType], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.querySystemTypes(ctx).
		Where("sps_controller_system_types.sps_controller_id = ?", spsControllerID)

	if strings.TrimSpace(params.Search) != "" {
		query = applySPSControllerSystemTypeSearch(query, params.Search)
	}

	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, err
	}

	var items []domainFacility.SPSControllerSystemType
	if err := query.Session(&gorm.Session{}).Preload("SPSController").Preload("SystemType").
		Order("sps_controller_system_types.created_at DESC").
		Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}
	if err := r.attachFieldDeviceCountsToValues(ctx, items); err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.SPSControllerSystemType]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

func (r *spsControllerSystemTypeRepo) GetPaginatedListByProjectID(ctx context.Context, projectID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SPSControllerSystemType], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.querySystemTypes(ctx).
		Joins("INNER JOIN project_sps_controllers ON project_sps_controllers.sps_controller_id = sps_controller_system_types.sps_controller_id").
		Where("project_sps_controllers.project_id = ?", projectID)

	if strings.TrimSpace(params.Search) != "" {
		query = applySPSControllerSystemTypeSearch(query, params.Search)
	}

	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, err
	}

	var items []domainFacility.SPSControllerSystemType
	if err := query.Session(&gorm.Session{}).Preload("SPSController").Preload("SystemType").
		Order("sps_controller_system_types.created_at DESC").
		Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}
	if err := r.attachFieldDeviceCountsToValues(ctx, items); err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.SPSControllerSystemType]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

type systemTypeFieldDeviceCountRow struct {
	SPSControllerSystemTypeID uuid.UUID `gorm:"column:sps_controller_system_type_id"`
	FieldDevicesCount         int64     `gorm:"column:field_devices_count"`
}

func (r *spsControllerSystemTypeRepo) attachFieldDeviceCountsToValues(ctx context.Context, items []domainFacility.SPSControllerSystemType) error {
	ptrs := make([]*domainFacility.SPSControllerSystemType, 0, len(items))
	for i := range items {
		ptrs = append(ptrs, &items[i])
	}
	return r.attachFieldDeviceCounts(ctx, ptrs)
}

func (r *spsControllerSystemTypeRepo) attachFieldDeviceCounts(ctx context.Context, items []*domainFacility.SPSControllerSystemType) error {
	if len(items) == 0 {
		return nil
	}

	ids := make([]uuid.UUID, 0, len(items))
	seen := make(map[uuid.UUID]struct{}, len(items))
	for _, item := range items {
		if item == nil || item.ID == uuid.Nil {
			continue
		}
		item.FieldDevicesCount = 0
		if _, ok := seen[item.ID]; ok {
			continue
		}
		seen[item.ID] = struct{}{}
		ids = append(ids, item.ID)
	}
	if len(ids) == 0 {
		return nil
	}

	var rows []systemTypeFieldDeviceCountRow
	if err := r.db.WithContext(ctx).
		Model(&FieldDeviceRecord{}).
		Select("sps_controller_system_type_id, COUNT(*) AS field_devices_count").
		Where("sps_controller_system_type_id IN ?", ids).
		Group("sps_controller_system_type_id").
		Scan(&rows).Error; err != nil {
		return err
	}

	counts := make(map[uuid.UUID]int, len(rows))
	for _, row := range rows {
		counts[row.SPSControllerSystemTypeID] = int(row.FieldDevicesCount)
	}
	for _, item := range items {
		if item != nil {
			item.FieldDevicesCount = counts[item.ID]
		}
	}
	return nil
}

func (r *spsControllerSystemTypeRepo) ListBySPSControllerID(ctx context.Context, spsControllerID uuid.UUID) ([]*domainFacility.SPSControllerSystemType, error) {
	var items []*domainFacility.SPSControllerSystemType
	err := r.db.WithContext(ctx).
		Where("sps_controller_id = ?", spsControllerID).
		Find(&items).Error
	return items, err
}

func (r *spsControllerSystemTypeRepo) GetIDsBySPSControllerIDs(ctx context.Context, ids []uuid.UUID) ([]uuid.UUID, error) {
	if len(ids) == 0 {
		return []uuid.UUID{}, nil
	}
	var out []uuid.UUID
	err := r.db.WithContext(ctx).Model(&domainFacility.SPSControllerSystemType{}).
		Where("sps_controller_id IN ?", ids).
		Pluck("id", &out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *spsControllerSystemTypeRepo) DeleteBySPSControllerIDs(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).
		Where("sps_controller_id IN ?", ids).
		Delete(&domainFacility.SPSControllerSystemType{}).Error
}
