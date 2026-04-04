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

type spsControllerSystemTypeRepo struct {
	*gormbase.BaseRepository[*domainFacility.SPSControllerSystemType]
	db *gorm.DB
}

func NewSPSControllerSystemTypeRepository(db *gorm.DB) domainFacility.SPSControllerSystemTypeStore {
	searchCallback := func(query *gorm.DB, search string) *gorm.DB {
		pattern := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		return query.Joins("LEFT JOIN sps_controllers ON sps_controllers.id = sps_controller_system_types.sps_controller_id").
			Joins("LEFT JOIN system_types ON system_types.id = sps_controller_system_types.system_type_id").
			Where("LOWER(sps_controller_system_types.document_name) LIKE ? OR LOWER(sps_controllers.device_name) LIKE ? OR LOWER(system_types.name) LIKE ?", pattern, pattern, pattern)
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
	err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Preload("SPSController").
		Preload("SystemType").
		Find(&items).Error
	return items, err
}

func (r *spsControllerSystemTypeRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SPSControllerSystemType], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&domainFacility.SPSControllerSystemType{})

	if strings.TrimSpace(params.Search) != "" {
		pattern := "%" + strings.ToLower(strings.TrimSpace(params.Search)) + "%"
		query = query.Joins("LEFT JOIN sps_controllers ON sps_controllers.id = sps_controller_system_types.sps_controller_id").
			Joins("LEFT JOIN system_types ON system_types.id = sps_controller_system_types.system_type_id").
			Where("LOWER(sps_controller_system_types.document_name) LIKE ? OR LOWER(sps_controllers.device_name) LIKE ? OR LOWER(system_types.name) LIKE ?", pattern, pattern, pattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []domainFacility.SPSControllerSystemType
	if err := query.Preload("SPSController").Preload("SystemType").
		Order("sps_controller_system_types.created_at DESC").
		Limit(limit).Offset(offset).Find(&items).Error; err != nil {
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

	query := r.db.WithContext(ctx).Model(&domainFacility.SPSControllerSystemType{}).
		Where("sps_controller_system_types.sps_controller_id = ?", spsControllerID)

	if strings.TrimSpace(params.Search) != "" {
		pattern := "%" + strings.ToLower(strings.TrimSpace(params.Search)) + "%"
		query = query.Joins("LEFT JOIN sps_controllers ON sps_controllers.id = sps_controller_system_types.sps_controller_id").
			Joins("LEFT JOIN system_types ON system_types.id = sps_controller_system_types.system_type_id").
			Where("LOWER(sps_controller_system_types.document_name) LIKE ? OR LOWER(sps_controllers.device_name) LIKE ? OR LOWER(system_types.name) LIKE ?", pattern, pattern, pattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []domainFacility.SPSControllerSystemType
	if err := query.Preload("SPSController").Preload("SystemType").
		Order("sps_controller_system_types.created_at DESC").
		Limit(limit).Offset(offset).Find(&items).Error; err != nil {
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

	query := r.db.WithContext(ctx).Model(&domainFacility.SPSControllerSystemType{}).
		Joins("INNER JOIN project_sps_controllers ON project_sps_controllers.sps_controller_id = sps_controller_system_types.sps_controller_id").
		Where("project_sps_controllers.project_id = ?", projectID)

	if strings.TrimSpace(params.Search) != "" {
		pattern := "%" + strings.ToLower(strings.TrimSpace(params.Search)) + "%"
		query = query.Joins("LEFT JOIN sps_controllers ON sps_controllers.id = sps_controller_system_types.sps_controller_id").
			Joins("LEFT JOIN system_types ON system_types.id = sps_controller_system_types.system_type_id").
			Where("LOWER(sps_controller_system_types.document_name) LIKE ? OR LOWER(sps_controllers.device_name) LIKE ? OR LOWER(system_types.name) LIKE ?", pattern, pattern, pattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []domainFacility.SPSControllerSystemType
	if err := query.Preload("SPSController").Preload("SystemType").
		Order("sps_controller_system_types.created_at DESC").
		Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.SPSControllerSystemType]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
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
