package facilitysql

import (
	"context"
	"slices"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/cursor"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const fieldDeviceCursorKind = "field_devices"

type fieldDeviceCursorColumn struct {
	expression string
	alias      string
}

type fieldDeviceCursorSort struct {
	key     string
	columns []fieldDeviceCursorColumn
	join    func(*gorm.DB) *gorm.DB
}

type fieldDeviceCursorRow struct {
	ID          uuid.UUID `gorm:"column:id"`
	Value       *string   `gorm:"column:cursor_value"`
	SecondValue *string   `gorm:"column:cursor_second_value"`
}

type fieldDeviceCursorFingerprint struct {
	Search                     string   `json:"search"`
	FilterSearch               string   `json:"filter_search"`
	OrderBy                    string   `json:"order_by"`
	Order                      string   `json:"order"`
	BuildingIDs                []string `json:"building_ids,omitempty"`
	ControlCabinetIDs          []string `json:"control_cabinet_ids,omitempty"`
	SPSControllerIDs           []string `json:"sps_controller_ids,omitempty"`
	SPSControllerSystemTypeIDs []string `json:"sps_controller_system_type_ids,omitempty"`
	ProjectIDs                 []string `json:"project_ids,omitempty"`
}

type fieldDeviceCursorScan struct {
	request    domainFacility.FieldDeviceCursorQuery
	definition fieldDeviceCursorSort
	anchor     *domainFacility.FieldDeviceCursorAnchor
}

type fieldDeviceCursorOrder struct {
	order     string
	direction string
}

type fieldDeviceKeyset struct {
	columns []fieldDeviceCursorColumn
	values  []*string
	id      uuid.UUID
	order   fieldDeviceCursorOrder
}

type cursorColumnKeyset struct {
	column   fieldDeviceCursorColumn
	value    *string
	tail     string
	tailArgs []any
	order    fieldDeviceCursorOrder
}

type fieldDeviceCursorPageBuild struct {
	items       []domainFacility.FieldDevice
	rows        []fieldDeviceCursorRow
	fingerprint string
	anchor      *domainFacility.FieldDeviceCursorAnchor
	hasMore     bool
}

var fieldDeviceCursorSorts = map[string]fieldDeviceCursorSort{
	"created_at":  cursorSort("created_at", "field_devices.created_at"),
	"bmk":         cursorSort("bmk", "field_devices.bmk"),
	"description": cursorSort("description", "field_devices.description"),
	"apparat_nr":  cursorSort("apparat_nr", "field_devices.apparat_nr"),
	"apparat": cursorSortWithJoin("apparat", "apparats_cursor.name", func(query *gorm.DB) *gorm.DB {
		return query.Joins("LEFT JOIN apparats apparats_cursor ON apparats_cursor.id = field_devices.apparat_id")
	}),
	"system_part": cursorSortWithJoin("system_part", "system_parts_cursor.name", func(query *gorm.DB) *gorm.DB {
		return query.Joins("LEFT JOIN system_parts system_parts_cursor ON system_parts_cursor.id = field_devices.system_part_id")
	}),
	"sps_system_type": {
		key: "sps_system_type",
		columns: []fieldDeviceCursorColumn{
			{expression: "scts_cursor.number", alias: "sort_value"},
			{expression: "scts_cursor.document_name", alias: "sort_second_value"},
		},
		join: func(query *gorm.DB) *gorm.DB {
			return query.Joins("LEFT JOIN sps_controller_system_types scts_cursor ON scts_cursor.id = field_devices.sps_controller_system_type_id")
		},
	},
	"spec_supplier":    specificationCursorSort("spec_supplier", "specification_supplier"),
	"spec_brand":       specificationCursorSort("spec_brand", "specification_brand"),
	"spec_type":        specificationCursorSort("spec_type", "specification_type"),
	"spec_motor_valve": specificationCursorSort("spec_motor_valve", "additional_info_motor_valve"),
	"spec_size":        specificationCursorSort("spec_size", "additional_info_size"),
	"spec_install_loc": specificationCursorSort("spec_install_loc", "additional_information_installation_location"),
	"spec_ph":          specificationCursorSort("spec_ph", "electrical_connection_ph"),
	"spec_acdc":        specificationCursorSort("spec_acdc", "electrical_connection_acdc"),
	"spec_amperage":    specificationCursorSort("spec_amperage", "electrical_connection_amperage"),
	"spec_power":       specificationCursorSort("spec_power", "electrical_connection_power"),
	"spec_rotation":    specificationCursorSort("spec_rotation", "electrical_connection_rotation"),
}

func (q fieldDeviceQuery) CursorPage(ctx context.Context, request domainFacility.FieldDeviceCursorQuery) (*domainFacility.FieldDeviceCursorPage, error) {
	request, sortDefinition := normalizeFieldDeviceCursorQuery(request)
	fingerprint, err := fieldDeviceQueryFingerprint(request)
	if err != nil {
		return nil, err
	}
	anchor, err := decodeFieldDeviceCursor(request.Cursor, fingerprint)
	if err != nil {
		return nil, err
	}
	rows, hasMore, err := q.scanCursorRows(ctx, fieldDeviceCursorScan{
		request: request, definition: sortDefinition, anchor: anchor,
	})
	if err != nil {
		return nil, err
	}
	items, err := q.loadCursorItems(ctx, rows)
	if err != nil {
		return nil, err
	}
	return buildFieldDeviceCursorPage(fieldDeviceCursorPageBuild{
		items: items, rows: rows, fingerprint: fingerprint, anchor: anchor, hasMore: hasMore,
	})
}

func normalizeFieldDeviceCursorQuery(query domainFacility.FieldDeviceCursorQuery) (domainFacility.FieldDeviceCursorQuery, fieldDeviceCursorSort) {
	if query.Limit <= 0 || query.Limit > fieldDeviceListMaxLimit {
		query.Limit = fieldDeviceListDefaultLimit
	}
	query.Search = strings.TrimSpace(query.Search)
	query.OrderBy = strings.TrimSpace(query.OrderBy)
	if _, ok := fieldDeviceCursorSorts[query.OrderBy]; !ok {
		query.OrderBy = "created_at"
	}
	query.Order = strings.ToLower(strings.TrimSpace(query.Order))
	if query.Order != "asc" && query.Order != "desc" {
		query.Order = "desc"
	}
	return query, fieldDeviceCursorSorts[query.OrderBy]
}

func (q fieldDeviceQuery) scanCursorRows(ctx context.Context, scan fieldDeviceCursorScan) ([]fieldDeviceCursorRow, bool, error) {
	request, definition, anchor := scan.request, scan.definition, scan.anchor
	query := activeFieldDevices(q.db.WithContext(ctx).Model(&FieldDeviceRecord{}))
	query, _ = applyFieldDeviceFilters(query, request.Filters)
	query = applyFieldDeviceCursorSearch(query, request)
	if definition.join != nil {
		query = definition.join(query)
	}
	query = selectFieldDeviceCursorRows(query, definition)
	if anchor != nil {
		values := []*string{anchor.Value}
		if len(definition.columns) > 1 {
			values = append(values, anchor.SecondValue)
		}
		query = applyFieldDeviceCursorAnchor(query, fieldDeviceKeyset{
			columns: definition.columns, values: values, id: anchor.ID,
			order: fieldDeviceCursorOrder{order: request.Order, direction: anchor.Direction},
		})
	}
	query = orderFieldDeviceCursorRows(query, definition, fieldDeviceCursorOrder{
		order: request.Order, direction: cursorDirection(anchor),
	})

	var rows []fieldDeviceCursorRow
	if err := query.Limit(request.Limit + 1).Scan(&rows).Error; err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > request.Limit
	if hasMore {
		rows = rows[:request.Limit]
	}
	if cursorDirection(anchor) == "previous" {
		slices.Reverse(rows)
	}
	return rows, hasMore, nil
}

func applyFieldDeviceCursorSearch(query *gorm.DB, request domainFacility.FieldDeviceCursorQuery) *gorm.DB {
	if request.Filters.Search != "" {
		query = applyFieldDeviceSearch(query, request.Filters.Search)
	}
	if request.Search != "" {
		query = applyFieldDeviceSearch(query, request.Search)
	}
	return query
}

func selectFieldDeviceCursorRows(query *gorm.DB, definition fieldDeviceCursorSort) *gorm.DB {
	columns := []string{"field_devices.id"}
	for index, column := range definition.columns {
		cursorAlias := "cursor_value"
		if index == 1 {
			cursorAlias = "cursor_second_value"
		}
		columns = append(columns,
			column.expression+" AS "+column.alias,
			"CAST("+column.expression+" AS TEXT) AS "+cursorAlias,
			"CASE WHEN "+column.expression+" IS NULL THEN 1 ELSE 0 END AS "+column.alias+"_null",
		)
	}
	return query.Distinct(strings.Join(columns, ", "))
}

func orderFieldDeviceCursorRows(query *gorm.DB, definition fieldDeviceCursorSort, requested fieldDeviceCursorOrder) *gorm.DB {
	effectiveOrder := requested.order
	nullOrder := "asc"
	if requested.direction == "previous" {
		effectiveOrder = oppositeOrder(requested.order)
		nullOrder = "desc"
	}
	for _, column := range definition.columns {
		query = query.Order(column.alias + "_null " + nullOrder)
		query = query.Order(column.alias + " " + effectiveOrder)
	}
	return query.Order("field_devices.id " + effectiveOrder)
}

func applyFieldDeviceCursorAnchor(query *gorm.DB, keyset fieldDeviceKeyset) *gorm.DB {
	predicate, args := fieldDeviceKeysetPredicate(keyset)
	return query.Where(predicate, args...)
}

func fieldDeviceKeysetPredicate(keyset fieldDeviceKeyset) (string, []any) {
	idOperator := cursorOperator(keyset.order.order, keyset.order.direction)
	predicate := "field_devices.id " + idOperator + " ?"
	args := []any{keyset.id}
	for index := len(keyset.columns) - 1; index >= 0; index-- {
		predicate, args = prependCursorColumn(cursorColumnKeyset{
			column: keyset.columns[index], value: keyset.values[index], tail: predicate,
			tailArgs: args, order: keyset.order,
		})
	}
	return predicate, args
}

func prependCursorColumn(keyset cursorColumnKeyset) (string, []any) {
	column, value := keyset.column, keyset.value
	rank := 1
	if value != nil {
		rank = 0
	}
	rankOperator := ">"
	if keyset.order.direction == "previous" {
		rankOperator = "<"
	}
	rankExpression := "CASE WHEN " + column.expression + " IS NULL THEN 1 ELSE 0 END"
	sameRank := keyset.tail
	sameRankArgs := append([]any(nil), keyset.tailArgs...)
	if value != nil {
		sameRank = "(" + column.expression + " " + cursorOperator(keyset.order.order, keyset.order.direction) + " ? OR (" + column.expression + " = ? AND " + keyset.tail + "))"
		sameRankArgs = append([]any{*value, *value}, keyset.tailArgs...)
	}
	predicate := "(" + rankExpression + " " + rankOperator + " ? OR (" + rankExpression + " = ? AND " + sameRank + "))"
	return predicate, append([]any{rank, rank}, sameRankArgs...)
}

func cursorOperator(order, direction string) string {
	if (order == "asc") == (direction != "previous") {
		return ">"
	}
	return "<"
}

func oppositeOrder(order string) string {
	if order == "asc" {
		return "desc"
	}
	return "asc"
}

func (q fieldDeviceQuery) loadCursorItems(ctx context.Context, rows []fieldDeviceCursorRow) ([]domainFacility.FieldDevice, error) {
	if len(rows) == 0 {
		return []domainFacility.FieldDevice{}, nil
	}
	ids := make([]uuid.UUID, len(rows))
	for index := range rows {
		ids[index] = rows[index].ID
	}
	query := activeFieldDevices(q.db.WithContext(ctx).Model(&FieldDeviceRecord{})).Where("field_devices.id IN ?", ids)
	items, err := scanFieldDeviceListRows(query, len(ids), 0)
	if err != nil {
		return nil, err
	}
	byID := make(map[uuid.UUID]domainFacility.FieldDevice, len(items))
	for _, item := range items {
		byID[item.ID] = item
	}
	ordered := make([]domainFacility.FieldDevice, 0, len(rows))
	for _, row := range rows {
		if item, ok := byID[row.ID]; ok {
			ordered = append(ordered, item)
		}
	}
	return ordered, nil
}

func buildFieldDeviceCursorPage(build fieldDeviceCursorPageBuild) (*domainFacility.FieldDeviceCursorPage, error) {
	page := &domainFacility.FieldDeviceCursorPage{Items: build.items}
	if len(build.rows) == 0 {
		return page, nil
	}
	direction := cursorDirection(build.anchor)
	hasPrevious := (build.anchor != nil && direction == "next") || (direction == "previous" && build.hasMore)
	hasNext := (direction == "next" && build.hasMore) || (build.anchor != nil && direction == "previous")
	var err error
	if hasPrevious {
		page.PreviousCursor, err = encodeFieldDeviceCursor(build.rows[0], build.fingerprint, "previous")
	}
	if err == nil && hasNext {
		page.NextCursor, err = encodeFieldDeviceCursor(build.rows[len(build.rows)-1], build.fingerprint, "next")
	}
	return page, err
}

func encodeFieldDeviceCursor(row fieldDeviceCursorRow, fingerprint, direction string) (string, error) {
	return cursor.Encode(fieldDeviceCursorKind, domainFacility.FieldDeviceCursorAnchor{
		Direction: direction, Fingerprint: fingerprint, ID: row.ID,
		Value: row.Value, SecondValue: row.SecondValue,
	})
}

func decodeFieldDeviceCursor(value, fingerprint string) (*domainFacility.FieldDeviceCursorAnchor, error) {
	if value == "" {
		return nil, nil
	}
	var anchor domainFacility.FieldDeviceCursorAnchor
	if err := cursor.Decode(value, fieldDeviceCursorKind, &anchor); err != nil {
		return nil, err
	}
	if anchor.Fingerprint != fingerprint || anchor.ID == uuid.Nil {
		return nil, cursor.ErrInvalid
	}
	if anchor.Direction != "next" && anchor.Direction != "previous" {
		return nil, cursor.ErrInvalid
	}
	return &anchor, nil
}

func cursorDirection(anchor *domainFacility.FieldDeviceCursorAnchor) string {
	if anchor == nil {
		return "next"
	}
	return anchor.Direction
}

func fieldDeviceQueryFingerprint(query domainFacility.FieldDeviceCursorQuery) (string, error) {
	filters := query.Filters
	return cursor.Fingerprint(fieldDeviceCursorFingerprint{
		Search: query.Search, FilterSearch: strings.TrimSpace(filters.Search), OrderBy: query.OrderBy, Order: query.Order,
		BuildingIDs:                normalizedCursorIDs(filters.BuildingID, filters.BuildingIDs),
		ControlCabinetIDs:          normalizedCursorIDs(filters.ControlCabinetID, filters.ControlCabinetIDs),
		SPSControllerIDs:           normalizedCursorIDs(filters.SPSControllerID, filters.SPSControllerIDs),
		SPSControllerSystemTypeIDs: normalizedCursorIDs(filters.SPSControllerSystemTypeID, filters.SPSControllerSystemTypeIDs),
		ProjectIDs:                 normalizedCursorIDs(filters.ProjectID, filters.ProjectIDs),
	})
}

func normalizedCursorIDs(single *uuid.UUID, many []uuid.UUID) []string {
	if single != nil {
		return []string{single.String()}
	}
	values := make([]string, 0, len(many))
	for _, id := range many {
		if id != uuid.Nil {
			values = append(values, id.String())
		}
	}
	slices.Sort(values)
	return slices.Compact(values)
}

func cursorSort(key, expression string) fieldDeviceCursorSort {
	return fieldDeviceCursorSort{key: key, columns: []fieldDeviceCursorColumn{{expression: expression, alias: "sort_value"}}}
}

func cursorSortWithJoin(key, expression string, join func(*gorm.DB) *gorm.DB) fieldDeviceCursorSort {
	definition := cursorSort(key, expression)
	definition.join = join
	return definition
}

func specificationCursorSort(key, column string) fieldDeviceCursorSort {
	return cursorSortWithJoin(key, "specs_cursor."+column, func(query *gorm.DB) *gorm.DB {
		return query.Joins("LEFT JOIN specifications specs_cursor ON specs_cursor.field_device_id = field_devices.id")
	})
}
