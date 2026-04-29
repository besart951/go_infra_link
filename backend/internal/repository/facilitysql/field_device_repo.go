package facilitysql

import (
	"context"
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type fieldDeviceRepo struct {
	db *gorm.DB
}

const (
	fieldDeviceListDefaultLimit = 300
	fieldDeviceListMaxLimit     = 300
	fieldDeviceListSelect       = `
		field_devices.id,
		field_devices.created_at,
		field_devices.updated_at,
		field_devices.bmk,
		field_devices.description,
		field_devices.apparat_nr,
		field_devices.text_individuell,
		field_devices.sps_controller_system_type_id,
		field_devices.system_part_id,
		field_devices.specification_id,
		field_devices.apparat_id,
		scts_list.id AS sps_system_type_id,
		scts_list.created_at AS sps_system_type_created_at,
		scts_list.updated_at AS sps_system_type_updated_at,
		scts_list.number AS sps_system_type_number,
		scts_list.document_name AS sps_system_type_document_name,
		sc_list.id AS sps_controller_id,
		sc_list.device_name AS sps_controller_device_name,
		st_list.id AS system_type_id,
		st_list.name AS system_type_name,
		apparats_list.id AS apparat_list_id,
		apparats_list.created_at AS apparat_created_at,
		apparats_list.updated_at AS apparat_updated_at,
		apparats_list.short_name AS apparat_short_name,
		apparats_list.name AS apparat_name,
		apparats_list.description AS apparat_description,
		system_parts_list.id AS system_part_list_id,
		system_parts_list.created_at AS system_part_created_at,
		system_parts_list.updated_at AS system_part_updated_at,
		system_parts_list.short_name AS system_part_short_name,
		system_parts_list.name AS system_part_name,
		system_parts_list.description AS system_part_description`
)

func NewFieldDeviceRepository(db *gorm.DB) domainFacility.FieldDeviceStore {
	return &fieldDeviceRepo{
		db: db,
	}
}

func normalizeFieldDeviceListPagination(page, limit int) (int, int) {
	page, limit = domain.NormalizePagination(page, limit, fieldDeviceListDefaultLimit)
	if limit > fieldDeviceListMaxLimit {
		limit = fieldDeviceListMaxLimit
	}
	return page, limit
}

func withFieldDeviceListJoins(query *gorm.DB) *gorm.DB {
	return query.
		Joins("LEFT JOIN sps_controller_system_types scts_list ON scts_list.id = field_devices.sps_controller_system_type_id").
		Joins("LEFT JOIN sps_controllers sc_list ON sc_list.id = scts_list.sps_controller_id").
		Joins("LEFT JOIN system_types st_list ON st_list.id = scts_list.system_type_id").
		Joins("LEFT JOIN apparats apparats_list ON apparats_list.id = field_devices.apparat_id").
		Joins("LEFT JOIN system_parts system_parts_list ON system_parts_list.id = field_devices.system_part_id")
}

func scanFieldDeviceListRows(query *gorm.DB, limit, offset int) ([]domainFacility.FieldDevice, error) {
	var rows []fieldDeviceListRow
	if err := withFieldDeviceListJoins(query).
		Select(fieldDeviceListSelect).
		Limit(limit).
		Offset(offset).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	return fieldDeviceListRowDomainValues(rows), nil
}

func applyFieldDeviceSorting(query *gorm.DB, params domain.PaginationParams) *gorm.DB {
	orderBy := strings.TrimSpace(params.OrderBy)
	if orderBy == "" {
		return query.Order("field_devices.created_at DESC")
	}

	order := strings.ToLower(strings.TrimSpace(params.Order))
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	switch orderBy {
	case "sps_system_type":
		query = query.Joins("LEFT JOIN sps_controller_system_types scts_sort ON scts_sort.id = field_devices.sps_controller_system_type_id")
		return query.Order("scts_sort.number " + order).Order("scts_sort.document_name " + order)
	case "bmk":
		return query.Order("field_devices.bmk " + order)
	case "description":
		return query.Order("field_devices.description " + order)
	case "apparat_nr":
		return query.Order("field_devices.apparat_nr " + order)
	case "apparat":
		query = query.Joins("LEFT JOIN apparats apparats_sort ON apparats_sort.id = field_devices.apparat_id")
		return query.Order("apparats_sort.name " + order)
	case "system_part":
		query = query.Joins("LEFT JOIN system_parts system_parts_sort ON system_parts_sort.id = field_devices.system_part_id")
		return query.Order("system_parts_sort.name " + order)
	case "spec_supplier":
		query = query.Joins("LEFT JOIN specifications specs_sort ON specs_sort.id = field_devices.specification_id")
		return query.Order("specs_sort.specification_supplier " + order)
	case "spec_brand":
		query = query.Joins("LEFT JOIN specifications specs_sort ON specs_sort.id = field_devices.specification_id")
		return query.Order("specs_sort.specification_brand " + order)
	case "spec_type":
		query = query.Joins("LEFT JOIN specifications specs_sort ON specs_sort.id = field_devices.specification_id")
		return query.Order("specs_sort.specification_type " + order)
	case "spec_motor_valve":
		query = query.Joins("LEFT JOIN specifications specs_sort ON specs_sort.id = field_devices.specification_id")
		return query.Order("specs_sort.additional_info_motor_valve " + order)
	case "spec_size":
		query = query.Joins("LEFT JOIN specifications specs_sort ON specs_sort.id = field_devices.specification_id")
		return query.Order("specs_sort.additional_info_size " + order)
	case "spec_install_loc":
		query = query.Joins("LEFT JOIN specifications specs_sort ON specs_sort.id = field_devices.specification_id")
		return query.Order("specs_sort.additional_information_installation_location " + order)
	case "spec_ph":
		query = query.Joins("LEFT JOIN specifications specs_sort ON specs_sort.id = field_devices.specification_id")
		return query.Order("specs_sort.electrical_connection_ph " + order)
	case "spec_acdc":
		query = query.Joins("LEFT JOIN specifications specs_sort ON specs_sort.id = field_devices.specification_id")
		return query.Order("specs_sort.electrical_connection_acdc " + order)
	case "spec_amperage":
		query = query.Joins("LEFT JOIN specifications specs_sort ON specs_sort.id = field_devices.specification_id")
		return query.Order("specs_sort.electrical_connection_amperage " + order)
	case "spec_power":
		query = query.Joins("LEFT JOIN specifications specs_sort ON specs_sort.id = field_devices.specification_id")
		return query.Order("specs_sort.electrical_connection_power " + order)
	case "spec_rotation":
		query = query.Joins("LEFT JOIN specifications specs_sort ON specs_sort.id = field_devices.specification_id")
		return query.Order("specs_sort.electrical_connection_rotation " + order)
	case "created_at":
		return query.Order("field_devices.created_at " + order)
	default:
		return query.Order("field_devices.created_at DESC")
	}
}

func (r *fieldDeviceRepo) GetByIds(ctx context.Context, ids []uuid.UUID) ([]*domainFacility.FieldDevice, error) {
	// FieldDevice needs to preload Specification
	if len(ids) == 0 {
		return []*domainFacility.FieldDevice{}, nil
	}
	var records []*FieldDeviceRecord
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Preload("Specification").Find(&records).Error
	return toFieldDeviceDomains(records), err
}

func (r *fieldDeviceRepo) Create(ctx context.Context, entity *domainFacility.FieldDevice) error {
	if err := entity.Base.InitForCreate(time.Now().UTC()); err != nil {
		return err
	}

	return r.db.WithContext(ctx).
		Omit(clause.Associations).
		Create(toFieldDeviceRecord(entity)).Error
}

func (r *fieldDeviceRepo) BulkCreate(ctx context.Context, entities []*domainFacility.FieldDevice, batchSize int) error {
	if len(entities) == 0 {
		return nil
	}

	now := time.Now().UTC()
	records := make([]*FieldDeviceRecord, len(entities))
	for i, entity := range entities {
		if err := entity.Base.InitForCreate(now); err != nil {
			return err
		}
		records[i] = toFieldDeviceRecord(entity)
	}

	if batchSize <= 0 {
		batchSize = gormbase.DefaultBatchSize
	}

	return r.db.WithContext(ctx).
		Omit(clause.Associations).
		CreateInBatches(records, batchSize).Error
}

func (r *fieldDeviceRepo) Update(ctx context.Context, entity *domainFacility.FieldDevice) error {
	entity.Base.TouchForUpdate(time.Now().UTC())
	return r.db.WithContext(ctx).Model(&FieldDeviceRecord{}).
		Where("id = ?", entity.ID).
		Updates(map[string]any{
			"updated_at":                    entity.UpdatedAt,
			"bmk":                           entity.BMK,
			"description":                   entity.Description,
			"apparat_nr":                    entity.ApparatNr,
			"text_individuell":              entity.TextIndividuell,
			"sps_controller_system_type_id": entity.SPSControllerSystemTypeID,
			"system_part_id":                entity.SystemPartID,
			"specification_id":              entity.SpecificationID,
			"apparat_id":                    entity.ApparatID,
		}).Error
}

func (r *fieldDeviceRepo) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&FieldDeviceRecord{}).Error
}

func (r *fieldDeviceRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.FieldDevice], error) {
	page, limit := normalizeFieldDeviceListPagination(params.Page, params.Limit)
	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&FieldDeviceRecord{})

	// Apply search
	if strings.TrimSpace(params.Search) != "" {
		pattern := "%" + strings.ToLower(strings.TrimSpace(params.Search)) + "%"
		query = query.Where("LOWER(bmk) LIKE ? OR LOWER(description) LIKE ?", pattern, pattern)
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	orderedQuery := applyFieldDeviceSorting(query.Session(&gorm.Session{}), params)
	items, err := scanFieldDeviceListRows(orderedQuery, limit, offset)
	if err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.FieldDevice]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

func (r *fieldDeviceRepo) GetIDsBySPSControllerSystemTypeIDs(ctx context.Context, ids []uuid.UUID) ([]uuid.UUID, error) {
	if len(ids) == 0 {
		return []uuid.UUID{}, nil
	}
	var out []uuid.UUID
	err := r.db.WithContext(ctx).Model(&FieldDeviceRecord{}).
		Where("sps_controller_system_type_id IN ?", ids).
		Pluck("id", &out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *fieldDeviceRepo) ExistsApparatNrConflict(ctx context.Context, spsControllerSystemTypeID uuid.UUID, systemPartID uuid.UUID, apparatID uuid.UUID, apparatNr int, excludeIDs []uuid.UUID) (bool, error) {
	db := r.db.WithContext(ctx).Model(&FieldDeviceRecord{}).
		Where("sps_controller_system_type_id = ?", spsControllerSystemTypeID).
		Where("system_part_id = ?", systemPartID).
		Where("apparat_id = ?", apparatID).
		Where("apparat_nr = ?", apparatNr)

	if len(excludeIDs) > 0 {
		db = db.Where("id NOT IN ?", excludeIDs)
	}

	var count int64
	err := db.Count(&count).Error
	return count > 0, err
}

func (r *fieldDeviceRepo) GetUsedApparatNumbers(ctx context.Context, spsControllerSystemTypeID uuid.UUID, systemPartID uuid.UUID, apparatID uuid.UUID) ([]int, error) {
	query := r.db.WithContext(ctx).Model(&FieldDeviceRecord{}).
		Where("sps_controller_system_type_id = ?", spsControllerSystemTypeID).
		Where("system_part_id = ?", systemPartID).
		Where("apparat_id = ?", apparatID)

	var nums []int
	if err := query.Pluck("apparat_nr", &nums).Error; err != nil {
		return nil, err
	}
	return nums, nil
}

func (r *fieldDeviceRepo) GetPaginatedListWithFilters(ctx context.Context, params domain.PaginationParams, filters domainFacility.FieldDeviceFilterParams) (*domain.PaginatedList[domainFacility.FieldDevice], error) {
	page, limit := normalizeFieldDeviceListPagination(params.Page, params.Limit)
	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&FieldDeviceRecord{})
	hasPotentialDuplicateRows := false

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

	if filters.ProjectID != nil {
		query = query.Joins("JOIN project_field_devices pfd ON pfd.field_device_id = field_devices.id").
			Where("pfd.project_id = ?", *filters.ProjectID)
		hasPotentialDuplicateRows = true
	} else if len(filters.ProjectIDs) > 0 {
		query = query.Joins("JOIN project_field_devices pfd ON pfd.field_device_id = field_devices.id").
			Where("pfd.project_id IN ?", filters.ProjectIDs)
		hasPotentialDuplicateRows = true
	}

	// Apply search
	if strings.TrimSpace(params.Search) != "" {
		pattern := "%" + strings.ToLower(strings.TrimSpace(params.Search)) + "%"
		query = query.Where("LOWER(field_devices.bmk) LIKE ? OR LOWER(field_devices.description) LIKE ?", pattern, pattern)
	}

	// Count total (count before DISTINCT to get accurate total)
	var total int64
	totalQuery := query.Session(&gorm.Session{})
	if hasPotentialDuplicateRows {
		totalQuery = totalQuery.Distinct("field_devices.id")
	}
	if err := totalQuery.Count(&total).Error; err != nil {
		return nil, err
	}

	orderedQuery := query.Session(&gorm.Session{})
	if hasPotentialDuplicateRows {
		distinctIDs := query.Session(&gorm.Session{}).Select("DISTINCT field_devices.id")
		orderedQuery = r.db.WithContext(ctx).
			Model(&FieldDeviceRecord{}).
			Joins("JOIN (?) ids ON ids.id = field_devices.id", distinctIDs)
	}
	orderedQuery = applyFieldDeviceSorting(orderedQuery, params)

	items, err := scanFieldDeviceListRows(orderedQuery, limit, offset)
	if err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.FieldDevice]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}
