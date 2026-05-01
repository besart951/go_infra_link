package facilitysql

import (
	"context"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/besart951/go_infra_link/backend/internal/repository/searchspec"
	"gorm.io/gorm"
)

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

type fieldDeviceQuery struct {
	db *gorm.DB
}

func newFieldDeviceQuery(db *gorm.DB) fieldDeviceQuery {
	return fieldDeviceQuery{db: db}
}

func (q fieldDeviceQuery) List(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.FieldDevice], error) {
	page, limit := normalizeFieldDeviceListPagination(params.Page, params.Limit)
	query := q.db.WithContext(ctx).Model(&FieldDeviceRecord{})

	if strings.TrimSpace(params.Search) != "" {
		query = applyFieldDeviceSearch(query, params.Search)
	}

	return q.page(ctx, query, params, page, limit, false)
}

func (q fieldDeviceQuery) ListWithFilters(ctx context.Context, params domain.PaginationParams, filters domainFacility.FieldDeviceFilterParams) (*domain.PaginatedList[domainFacility.FieldDevice], error) {
	page, limit := normalizeFieldDeviceListPagination(params.Page, params.Limit)
	query := q.db.WithContext(ctx).Model(&FieldDeviceRecord{})
	query, hasPotentialDuplicateRows := applyFieldDeviceFilters(query, filters)

	if strings.TrimSpace(params.Search) != "" {
		query = applyFieldDeviceSearch(query, params.Search)
	}

	return q.page(ctx, query, params, page, limit, hasPotentialDuplicateRows)
}

func (q fieldDeviceQuery) page(ctx context.Context, query *gorm.DB, params domain.PaginationParams, page, limit int, hasPotentialDuplicateRows bool) (*domain.PaginatedList[domainFacility.FieldDevice], error) {
	total, err := countFieldDeviceRows(query, hasPotentialDuplicateRows)
	if err != nil {
		return nil, err
	}

	orderedQuery := query.Session(&gorm.Session{})
	if hasPotentialDuplicateRows {
		distinctIDs := query.Session(&gorm.Session{}).Select("DISTINCT field_devices.id")
		orderedQuery = q.db.WithContext(ctx).
			Model(&FieldDeviceRecord{}).
			Joins("JOIN (?) ids ON ids.id = field_devices.id", distinctIDs)
	}
	orderedQuery = applyFieldDeviceSorting(orderedQuery, params)

	items, err := scanFieldDeviceListRows(orderedQuery, limit, (page-1)*limit)
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

func normalizeFieldDeviceListPagination(page, limit int) (int, int) {
	page, limit = domain.NormalizePagination(page, limit, fieldDeviceListDefaultLimit)
	if limit > fieldDeviceListMaxLimit {
		limit = fieldDeviceListMaxLimit
	}
	return page, limit
}

func applyFieldDeviceFilters(query *gorm.DB, filters domainFacility.FieldDeviceFilterParams) (*gorm.DB, bool) {
	hasPotentialDuplicateRows := false

	if filters.SPSControllerSystemTypeID != nil {
		query = query.Where("field_devices.sps_controller_system_type_id = ?", *filters.SPSControllerSystemTypeID)
	}

	if filters.SPSControllerID != nil {
		query = query.Joins("JOIN sps_controller_system_types ON sps_controller_system_types.id = field_devices.sps_controller_system_type_id").
			Where("sps_controller_system_types.sps_controller_id = ?", *filters.SPSControllerID)
	}

	if filters.ControlCabinetID != nil {
		query = query.Joins("JOIN sps_controller_system_types scts ON scts.id = field_devices.sps_controller_system_type_id").
			Joins("JOIN sps_controllers sc ON sc.id = scts.sps_controller_id").
			Where("sc.control_cabinet_id = ?", *filters.ControlCabinetID)
	}

	if filters.BuildingID != nil {
		query = query.Joins("JOIN sps_controller_system_types scts2 ON scts2.id = field_devices.sps_controller_system_type_id").
			Joins("JOIN sps_controllers sc2 ON sc2.id = scts2.sps_controller_id").
			Joins("JOIN control_cabinets cc ON cc.id = sc2.control_cabinet_id").
			Where("cc.building_id = ?", *filters.BuildingID)
	}

	if filters.ProjectID != nil {
		query = query.Joins("JOIN project_field_devices pfd ON pfd.field_device_id = field_devices.id").
			Where("pfd.project_id = ?", *filters.ProjectID)
	} else if len(filters.ProjectIDs) > 0 {
		query = query.Joins("JOIN project_field_devices pfd ON pfd.field_device_id = field_devices.id").
			Where("pfd.project_id IN ?", filters.ProjectIDs)
		hasPotentialDuplicateRows = true
	}

	return query, hasPotentialDuplicateRows
}

func countFieldDeviceRows(query *gorm.DB, hasPotentialDuplicateRows bool) (int64, error) {
	var total int64
	totalQuery := query.Session(&gorm.Session{})
	if hasPotentialDuplicateRows {
		totalQuery = totalQuery.Distinct("field_devices.id")
	}
	if err := totalQuery.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
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

func applyFieldDeviceSearch(query *gorm.DB, search string) *gorm.DB {
	term := strings.ToLower(strings.TrimSpace(search))
	if term == "" {
		return query
	}
	if len([]rune(term)) < 3 {
		return query.Where(
			"(LOWER(field_devices.bmk) LIKE ? OR LOWER(field_devices.description) LIKE ?)",
			"%"+term+"%",
			term+"%",
		)
	}
	return gormbase.ApplyTrigramSearch(query, term, searchspec.FieldDevices.SearchColumns("field_devices.")...)
}
