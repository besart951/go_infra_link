package facilitysql

import (
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type fieldDeviceRepo struct {
	*gormbase.BaseRepository[*domainFacility.FieldDevice]
	db *gorm.DB
}

func NewFieldDeviceRepository(db *gorm.DB) domainFacility.FieldDeviceStore {
	searchCallback := func(query *gorm.DB, search string) *gorm.DB {
		pattern := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		return query.Where("LOWER(bmk) LIKE ? OR LOWER(description) LIKE ?", pattern, pattern)
	}

	baseRepo := gormbase.NewBaseRepository[*domainFacility.FieldDevice](db, searchCallback)
	return &fieldDeviceRepo{
		BaseRepository: baseRepo,
		db:             db,
	}
}

func (r *fieldDeviceRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.FieldDevice, error) {
	// FieldDevice needs to preload Specification
	if len(ids) == 0 {
		return []*domainFacility.FieldDevice{}, nil
	}
	var items []*domainFacility.FieldDevice
	err := r.db.Where("deleted_at IS NULL").Where("id IN ?", ids).Preload("Specification").Find(&items).Error
	return items, err
}

func (r *fieldDeviceRepo) Create(entity *domainFacility.FieldDevice) error {
	return r.BaseRepository.Create(entity)
}

func (r *fieldDeviceRepo) Update(entity *domainFacility.FieldDevice) error {
	return r.BaseRepository.Update(entity)
}

func (r *fieldDeviceRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.BaseRepository.DeleteByIds(ids)
}

func (r *fieldDeviceRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.FieldDevice], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 10)
	if err != nil {
		return nil, err
	}

	// Convert []*FieldDevice to []FieldDevice for the interface
	items := make([]domainFacility.FieldDevice, len(result.Items))
	for i, item := range result.Items {
		items[i] = *item
	}

	return &domain.PaginatedList[domainFacility.FieldDevice]{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}, nil
}

func (r *fieldDeviceRepo) GetIDsBySPSControllerSystemTypeIDs(ids []uuid.UUID) ([]uuid.UUID, error) {
	if len(ids) == 0 {
		return []uuid.UUID{}, nil
	}
	var out []uuid.UUID
	err := r.db.Model(&domainFacility.FieldDevice{}).
		Where("deleted_at IS NULL").
		Where("sps_controller_system_type_id IN ?", ids).
		Pluck("id", &out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *fieldDeviceRepo) ExistsApparatNrConflict(spsControllerSystemTypeID uuid.UUID, systemPartID *uuid.UUID, apparatID uuid.UUID, apparatNr int, excludeID *uuid.UUID) (bool, error) {
	db := r.db.Model(&domainFacility.FieldDevice{}).
		Where("deleted_at IS NULL").
		Where("sps_controller_system_type_id = ?", spsControllerSystemTypeID).
		Where("apparat_id = ?", apparatID).
		Where("apparat_nr = ?", apparatNr)

	if systemPartID != nil {
		db = db.Where("system_part_id = ?", *systemPartID)
	} else {
		db = db.Where("system_part_id IS NULL")
	}

	if excludeID != nil {
		db = db.Where("id != ?", *excludeID)
	}

	var count int64
	err := db.Count(&count).Error
	return count > 0, err
}

func (r *fieldDeviceRepo) GetUsedApparatNumbers(spsControllerSystemTypeID uuid.UUID, systemPartID *uuid.UUID, apparatID uuid.UUID) ([]int, error) {
	query := r.db.Model(&domainFacility.FieldDevice{}).
		Where("deleted_at IS NULL").
		Where("sps_controller_system_type_id = ?", spsControllerSystemTypeID).
		Where("apparat_id = ?", apparatID)

	if systemPartID != nil {
		query = query.Where("system_part_id = ?", *systemPartID)
	} else {
		query = query.Where("system_part_id IS NULL")
	}

	var nums []int
	if err := query.Pluck("apparat_nr", &nums).Error; err != nil {
		return nil, err
	}
	return nums, nil
}

func (r *fieldDeviceRepo) GetPaginatedListWithFilters(params domain.PaginationParams, filters domainFacility.FieldDeviceFilterParams) (*domain.PaginatedList[domainFacility.FieldDevice], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 1000)
	offset := (page - 1) * limit

	query := r.db.Model(&domainFacility.FieldDevice{}).Where("deleted_at IS NULL")

	// Apply filters by joining through the hierarchy
	if filters.SPSControllerSystemTypeID != nil {
		query = query.Where("sps_controller_system_type_id = ?", *filters.SPSControllerSystemTypeID)
	}

	if filters.SPSControllerID != nil {
		// Join through sps_controller_system_types to filter by sps_controller_id
		query = query.Joins("JOIN sps_controller_system_types ON sps_controller_system_types.id = field_devices.sps_controller_system_type_id").
			Where("sps_controller_system_types.sps_controller_id = ?", *filters.SPSControllerID)
	}

	if filters.ControlCabinetID != nil {
		// Join through sps_controller_system_types and sps_controllers to filter by control_cabinet_id
		query = query.Joins("JOIN sps_controller_system_types scts ON scts.id = field_devices.sps_controller_system_type_id").
			Joins("JOIN sps_controllers sc ON sc.id = scts.sps_controller_id").
			Where("sc.control_cabinet_id = ?", *filters.ControlCabinetID)
	}

	if filters.BuildingID != nil {
		// Join through the full hierarchy to filter by building_id
		query = query.Joins("JOIN sps_controller_system_types scts2 ON scts2.id = field_devices.sps_controller_system_type_id").
			Joins("JOIN sps_controllers sc2 ON sc2.id = scts2.sps_controller_id").
			Joins("JOIN control_cabinets cc ON cc.id = sc2.control_cabinet_id").
			Where("cc.building_id = ?", *filters.BuildingID)
	}

	// Apply search
	if strings.TrimSpace(params.Search) != "" {
		pattern := "%" + strings.ToLower(strings.TrimSpace(params.Search)) + "%"
		query = query.Where("LOWER(field_devices.bmk) LIKE ? OR LOWER(field_devices.description) LIKE ?", pattern, pattern)
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Fetch items with preloads
	var items []domainFacility.FieldDevice
	if err := query.
		Select("DISTINCT field_devices.*").
		Order("field_devices.created_at DESC").
		Preload("Specification").
		Limit(limit).
		Offset(offset).
		Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.FieldDevice]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}
