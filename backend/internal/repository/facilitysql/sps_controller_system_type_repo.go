package facilitysql

import (
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type spsControllerSystemTypeRepo struct {
	db *gorm.DB
}

func NewSPSControllerSystemTypeRepository(db *gorm.DB) domainFacility.SPSControllerSystemTypeStore {
	return &spsControllerSystemTypeRepo{db: db}
}

func (r *spsControllerSystemTypeRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.SPSControllerSystemType, error) {
	if len(ids) == 0 {
		return []*domainFacility.SPSControllerSystemType{}, nil
	}
	var items []*domainFacility.SPSControllerSystemType
	err := r.db.
		Where("sps_controller_system_types.deleted_at IS NULL").
		Where("id IN ?", ids).
		Preload("SPSController").
		Preload("SystemType").
		Find(&items).Error
	return items, err
}

func (r *spsControllerSystemTypeRepo) Create(entity *domainFacility.SPSControllerSystemType) error {
	now := time.Now().UTC()
	if err := entity.Base.InitForCreate(now); err != nil {
		return err
	}
	return r.db.Create(entity).Error
}

func (r *spsControllerSystemTypeRepo) Update(entity *domainFacility.SPSControllerSystemType) error {
	entity.Base.TouchForUpdate(time.Now().UTC())
	return r.db.Save(entity).Error
}

func (r *spsControllerSystemTypeRepo) DeleteByIds(ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	now := time.Now().UTC()
	return r.db.Model(&domainFacility.SPSControllerSystemType{}).
		Where("id IN ?", ids).
		Updates(map[string]any{"deleted_at": now, "updated_at": now}).Error
}

func (r *spsControllerSystemTypeRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SPSControllerSystemType], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.Model(&domainFacility.SPSControllerSystemType{}).
		Where("sps_controller_system_types.deleted_at IS NULL")

	if strings.TrimSpace(params.Search) != "" {
		pattern := "%" + strings.TrimSpace(params.Search) + "%"
		query = query.Joins("LEFT JOIN sps_controllers ON sps_controllers.id = sps_controller_system_types.sps_controller_id").
			Joins("LEFT JOIN system_types ON system_types.id = sps_controller_system_types.system_type_id").
			Where("sps_controller_system_types.document_name ILIKE ? OR sps_controllers.device_name ILIKE ? OR system_types.name ILIKE ?", pattern, pattern, pattern)
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

func (r *spsControllerSystemTypeRepo) SoftDeleteBySPSControllerIDs(ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	now := time.Now().UTC()
	return r.db.Model(&domainFacility.SPSControllerSystemType{}).
		Where("sps_controller_id IN ?", ids).
		Updates(map[string]any{"deleted_at": now, "updated_at": now}).Error
}
