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

const (
	fieldDeviceCursorKind       = "field_devices"
	fieldDeviceSearchProbeLimit = 50_000
)

type fieldDeviceCursorColumn struct {
	expression string
	alias      string
	nullable   bool
}

type fieldDeviceCursorSort struct {
	key                   string
	columns               []fieldDeviceCursorColumn
	idExpression          string
	includesProjectScope  bool
	includesBuildingScope bool
	join                  func(*gorm.DB) *gorm.DB
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
	columns      []fieldDeviceCursorColumn
	values       []*string
	id           uuid.UUID
	idExpression string
	order        fieldDeviceCursorOrder
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
	"created_at":  requiredCursorSort("created_at", "field_devices.created_at"),
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
			{expression: "scts_cursor.number", alias: "sort_value", nullable: true},
			{expression: "scts_cursor.document_name", alias: "sort_second_value", nullable: true},
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
	sortDefinition = projectedFieldDeviceCursorSort(q.db, sortDefinition)
	sortDefinition = projectScopedFieldDeviceCursorSort(q.db, request, sortDefinition)
	sortDefinition = buildingScopedFieldDeviceCursorSort(q.db, request, sortDefinition)
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

func buildingScopedFieldDeviceCursorSort(db *gorm.DB, request domainFacility.FieldDeviceCursorQuery, definition fieldDeviceCursorSort) fieldDeviceCursorSort {
	buildingID := request.Filters.BuildingID
	if definition.key != "created_at" || buildingID == nil || db.Dialector == nil || db.Dialector.Name() != "postgres" {
		return definition
	}
	definition.columns = []fieldDeviceCursorColumn{{expression: "fd_building_scope.field_device_created_at", alias: "sort_value"}}
	definition.idExpression = "fd_building_scope.field_device_id"
	definition.includesBuildingScope = true
	definition.join = func(query *gorm.DB) *gorm.DB {
		return query.Joins("JOIN field_device_building_cursor_values fd_building_scope ON fd_building_scope.field_device_id=field_devices.id AND fd_building_scope.building_id=?", *buildingID)
	}
	return definition
}

func projectScopedFieldDeviceCursorSort(db *gorm.DB, request domainFacility.FieldDeviceCursorQuery, definition fieldDeviceCursorSort) fieldDeviceCursorSort {
	projectID := request.Filters.ProjectID
	if definition.key != "created_at" || projectID == nil || db.Dialector == nil || db.Dialector.Name() != "postgres" {
		return definition
	}
	definition.columns = []fieldDeviceCursorColumn{{expression: "pfd_cursor.field_device_created_at", alias: "sort_value"}}
	definition.idExpression = "pfd_cursor.field_device_id"
	definition.includesProjectScope = true
	definition.join = func(query *gorm.DB) *gorm.DB {
		return query.Joins("JOIN project_field_devices pfd_cursor ON pfd_cursor.field_device_id=field_devices.id AND pfd_cursor.project_id=?", *projectID)
	}
	return definition
}

func projectedFieldDeviceCursorSort(db *gorm.DB, definition fieldDeviceCursorSort) fieldDeviceCursorSort {
	columns, ok := fieldDeviceProjectionColumns[definition.key]
	if !ok || db.Dialector == nil || db.Dialector.Name() != "postgres" {
		return definition
	}
	definition.columns = columns
	definition.idExpression = "fd_cursor.field_device_id"
	definition.join = func(query *gorm.DB) *gorm.DB {
		return query.Joins("JOIN field_device_cursor_values fd_cursor ON fd_cursor.field_device_id=field_devices.id")
	}
	return definition
}

var fieldDeviceProjectionColumns = map[string][]fieldDeviceCursorColumn{
	"sps_system_type": {
		{expression: "fd_cursor.sps_number", alias: "sort_value", nullable: true},
		{expression: "fd_cursor.sps_document_name", alias: "sort_second_value", nullable: true},
	},
	"spec_supplier":    projectedSpecificationColumn("specification_supplier"),
	"spec_brand":       projectedSpecificationColumn("specification_brand"),
	"spec_type":        projectedSpecificationColumn("specification_type"),
	"spec_motor_valve": projectedSpecificationColumn("additional_info_motor_valve"),
	"spec_size":        projectedSpecificationColumn("additional_info_size"),
	"spec_install_loc": projectedSpecificationColumn("additional_information_installation_location"),
	"spec_ph":          projectedSpecificationColumn("electrical_connection_ph"),
	"spec_acdc":        projectedSpecificationColumn("electrical_connection_acdc"),
	"spec_amperage":    projectedSpecificationColumn("electrical_connection_amperage"),
	"spec_power":       projectedSpecificationColumn("electrical_connection_power"),
	"spec_rotation":    projectedSpecificationColumn("electrical_connection_rotation"),
}

func projectedSpecificationColumn(column string) []fieldDeviceCursorColumn {
	return []fieldDeviceCursorColumn{{expression: "fd_cursor." + column, alias: "sort_value", nullable: true}}
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
	query := q.buildCursorRowsQuery(ctx, scan)
	keyset := fieldDeviceCursorKeyset(scan)
	if shouldProbeFieldDeviceSearch(q.db, scan) {
		if rows, hasMore, complete, err := q.scanBoundedSearch(ctx, query, scan, keyset); complete || err != nil {
			return rows, hasMore, err
		}
	}
	query = applyFieldDeviceCursorSearch(query, scan.request)
	query = selectFieldDeviceCursorRows(query, scan.definition)
	if needsSecondNullNextSplit(keyset) {
		return scanSecondNullNextCursorRows(query, scan, keyset)
	}
	if needsCursorRankSplit(keyset) {
		return scanSplitCursorRankRows(query, scan, keyset)
	}
	if scan.anchor != nil {
		query = applyFieldDeviceCursorAnchor(query, keyset)
	}
	return executeCursorRows(query, scan, scan.request.Limit+1)
}

func needsSecondNullNextSplit(keyset fieldDeviceKeyset) bool {
	return len(keyset.columns) == 2 && len(keyset.values) == 2 && keyset.values[0] != nil &&
		keyset.values[1] == nil && keyset.order.direction == "next"
}

func scanSecondNullNextCursorRows(query *gorm.DB, scan fieldDeviceCursorScan, keyset fieldDeviceKeyset) ([]fieldDeviceCursorRow, bool, error) {
	primary, primaryArgs := secondNullSameValuePredicate(keyset)
	rows, err := scanOrderedCursorRows(query.Where(primary, primaryArgs...), scan, scan.request.Limit+1)
	if err != nil || len(rows) > scan.request.Limit {
		rows, hasMore := finalizeCursorRows(rows, scan.request.Limit, "next")
		return rows, hasMore, err
	}
	rows, err = appendCursorBranch(rows, query, scan, outerValueRemainder(keyset), scan.request.Limit+1)
	if err == nil && len(rows) <= scan.request.Limit {
		rank := nullRankExpression(keyset.columns[0])
		rows, err = appendCursorBranch(rows, query, scan, cursorBranch{predicate: rank + " > ?", args: []any{0}}, scan.request.Limit+1)
	}
	rows, hasMore := finalizeCursorRows(rows, scan.request.Limit, "next")
	return rows, hasMore, err
}

type cursorBranch struct {
	predicate string
	args      []any
}

func appendCursorBranch(rows []fieldDeviceCursorRow, query *gorm.DB, scan fieldDeviceCursorScan, branch cursorBranch, target int) ([]fieldDeviceCursorRow, error) {
	missing := target - len(rows)
	if missing <= 0 {
		return rows, nil
	}
	additional, err := scanOrderedCursorRows(query.Where(branch.predicate, branch.args...), scan, missing)
	return append(rows, additional...), err
}

func secondNullSameValuePredicate(keyset fieldDeviceKeyset) (string, []any) {
	first, second := keyset.columns[0], keyset.columns[1]
	id := cursorIDExpression(keyset.idExpression)
	idOperator := cursorOperator(keyset.order.order, keyset.order.direction)
	predicate := nullRankExpression(first) + "=0 AND " + first.expression + "=? AND " +
		nullRankExpression(second) + "=1 AND " + second.expression + " IS NULL AND " + id + " " + idOperator + " ?"
	return predicate, []any{*keyset.values[0], keyset.id}
}

func outerValueRemainder(keyset fieldDeviceKeyset) cursorBranch {
	column := keyset.columns[0]
	operator := cursorOperator(keyset.order.order, keyset.order.direction)
	return cursorBranch{
		predicate: nullRankExpression(column) + "=0 AND " + column.expression + " " + operator + " ?",
		args:      []any{*keyset.values[0]},
	}
}

func (q fieldDeviceQuery) buildCursorRowsQuery(ctx context.Context, scan fieldDeviceCursorScan) *gorm.DB {
	query := activeFieldDevices(q.db.WithContext(ctx).Model(&FieldDeviceRecord{}))
	query, _ = applyFieldDeviceFilters(query, fieldDeviceCursorFilters(scan))
	if scan.definition.join != nil {
		query = scan.definition.join(query)
	}
	return query
}

func shouldProbeFieldDeviceSearch(db *gorm.DB, scan fieldDeviceCursorScan) bool {
	return scan.definition.key == "created_at" && hasFieldDeviceCursorSearch(scan.request) &&
		allFieldDeviceSearchTermsLong(scan.request) && db.Dialector != nil && db.Dialector.Name() == "postgres"
}

func hasFieldDeviceCursorSearch(request domainFacility.FieldDeviceCursorQuery) bool {
	return strings.TrimSpace(request.Search) != "" || strings.TrimSpace(request.Filters.Search) != ""
}

func allFieldDeviceSearchTermsLong(request domainFacility.FieldDeviceCursorQuery) bool {
	for _, term := range []string{request.Search, request.Filters.Search} {
		term = strings.TrimSpace(term)
		if term != "" && len(term) < 3 {
			return false
		}
	}
	return true
}

func (q fieldDeviceQuery) scanBoundedSearch(ctx context.Context, query *gorm.DB, scan fieldDeviceCursorScan, keyset fieldDeviceKeyset) ([]fieldDeviceCursorRow, bool, bool, error) {
	candidate := selectFieldDeviceCursorCandidates(query, scan.definition)
	if scan.anchor != nil {
		candidate = applyFieldDeviceCursorAnchor(candidate, keyset)
	}
	candidate = orderFieldDeviceCursorRows(candidate, scan.definition, fieldDeviceCursorOrder{
		order: scan.request.Order, direction: cursorDirection(scan.anchor),
	}).Limit(fieldDeviceSearchProbeLimit)
	outer := q.db.WithContext(ctx).Table("(?) AS candidates", candidate)
	outer = applyCandidateSearch(outer, scan.request)
	outerDefinition := scan.definition
	outerDefinition.idExpression = "candidates.id"
	outer = orderFieldDeviceCursorRows(outer, outerDefinition, fieldDeviceCursorOrder{
		order: scan.request.Order, direction: cursorDirection(scan.anchor),
	})
	var rows []fieldDeviceCursorRow
	if err := outer.Limit(scan.request.Limit + 1).Scan(&rows).Error; err != nil {
		return nil, false, false, err
	}
	if len(rows) <= scan.request.Limit {
		return nil, false, false, nil
	}
	rows, hasMore := finalizeCursorRows(rows, scan.request.Limit, cursorDirection(scan.anchor))
	return rows, hasMore, true, nil
}

func selectFieldDeviceCursorCandidates(query *gorm.DB, definition fieldDeviceCursorSort) *gorm.DB {
	columns := fieldDeviceCursorSelect(definition) + ",field_devices.bmk AS search_bmk,field_devices.description AS search_description"
	return query.Select(columns)
}

func applyCandidateSearch(query *gorm.DB, request domainFacility.FieldDeviceCursorQuery) *gorm.DB {
	for _, term := range []string{request.Filters.Search, request.Search} {
		term = strings.TrimSpace(term)
		if term == "" {
			continue
		}
		pattern := "%" + strings.ToLower(term) + "%"
		query = query.Where("LOWER(COALESCE(candidates.search_bmk,'') || CHR(1) || COALESCE(candidates.search_description,'')) LIKE ?", pattern)
	}
	return query
}

func fieldDeviceCursorFilters(scan fieldDeviceCursorScan) domainFacility.FieldDeviceFilterParams {
	filters := scan.request.Filters
	if scan.definition.includesProjectScope {
		filters.ProjectID = nil
	}
	if scan.definition.includesBuildingScope {
		filters.BuildingID = nil
	}
	return filters
}

func fieldDeviceCursorKeyset(scan fieldDeviceCursorScan) fieldDeviceKeyset {
	if scan.anchor == nil {
		return fieldDeviceKeyset{}
	}
	values := []*string{scan.anchor.Value}
	if len(scan.definition.columns) > 1 {
		values = append(values, scan.anchor.SecondValue)
	}
	return fieldDeviceKeyset{
		columns: scan.definition.columns, values: values, id: scan.anchor.ID,
		idExpression: scan.definition.idExpression,
		order:        fieldDeviceCursorOrder{order: scan.request.Order, direction: scan.anchor.Direction},
	}
}

func executeCursorRows(query *gorm.DB, scan fieldDeviceCursorScan, fetchLimit int) ([]fieldDeviceCursorRow, bool, error) {
	rows, err := scanOrderedCursorRows(query, scan, fetchLimit)
	if err != nil {
		return nil, false, err
	}
	rows, hasMore := finalizeCursorRows(rows, scan.request.Limit, cursorDirection(scan.anchor))
	return rows, hasMore, nil
}

func scanOrderedCursorRows(query *gorm.DB, scan fieldDeviceCursorScan, fetchLimit int) ([]fieldDeviceCursorRow, error) {
	query = orderFieldDeviceCursorRows(query, scan.definition, fieldDeviceCursorOrder{
		order: scan.request.Order, direction: cursorDirection(scan.anchor),
	})
	var rows []fieldDeviceCursorRow
	if err := query.Limit(fetchLimit).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func finalizeCursorRows(rows []fieldDeviceCursorRow, limit int, direction string) ([]fieldDeviceCursorRow, bool) {
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	if direction == "previous" {
		slices.Reverse(rows)
	}
	return rows, hasMore
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
	return query.Select(fieldDeviceCursorSelect(definition))
}

func fieldDeviceCursorSelect(definition fieldDeviceCursorSort) string {
	columns := []string{"field_devices.id"}
	for index, column := range definition.columns {
		cursorAlias := "cursor_value"
		if index == 1 {
			cursorAlias = "cursor_second_value"
		}
		columns = append(columns,
			column.expression+" AS "+column.alias,
			"CAST("+column.expression+" AS TEXT) AS "+cursorAlias,
			nullRankExpression(column)+" AS "+column.alias+"_null",
		)
	}
	return strings.Join(columns, ", ")
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
	return query.Order(cursorIDExpression(definition.idExpression) + " " + effectiveOrder)
}

func applyFieldDeviceCursorAnchor(query *gorm.DB, keyset fieldDeviceKeyset) *gorm.DB {
	predicate, args := fieldDeviceKeysetPredicate(keyset)
	return query.Where(predicate, args...)
}

func needsCursorRankSplit(keyset fieldDeviceKeyset) bool {
	if len(keyset.columns) == 0 || len(keyset.values) == 0 || !keyset.columns[0].nullable {
		return false
	}
	return keyset.order.direction == "next" && keyset.values[0] != nil ||
		keyset.order.direction == "previous" && keyset.values[0] == nil
}

func scanSplitCursorRankRows(query *gorm.DB, scan fieldDeviceCursorScan, keyset fieldDeviceKeyset) ([]fieldDeviceCursorRow, bool, error) {
	predicate, args := fieldDeviceOuterSameRankPredicate(keyset)
	rows, err := scanOrderedCursorRows(query.Where(predicate, args...), scan, scan.request.Limit+1)
	if err != nil || len(rows) > scan.request.Limit {
		rows, hasMore := finalizeCursorRows(rows, scan.request.Limit, "previous")
		return rows, hasMore, err
	}
	missing := scan.request.Limit + 1 - len(rows)
	rank := nullRankExpression(keyset.columns[0])
	fallback, err := scanOrderedCursorRows(query.Where(rank+" "+cursorRankOperator(keyset)+" ?", cursorRank(keyset.values[0])), scan, missing)
	rows = append(rows, fallback...)
	rows, hasMore := finalizeCursorRows(rows, scan.request.Limit, "previous")
	return rows, hasMore, err
}

func fieldDeviceOuterSameRankPredicate(keyset fieldDeviceKeyset) (string, []any) {
	if allCursorValuesPresent(keyset) {
		return cursorTuplePredicate(keyset)
	}
	predicate := cursorIDExpression(keyset.idExpression) + " " + cursorOperator(keyset.order.order, keyset.order.direction) + " ?"
	args := []any{keyset.id}
	for index := len(keyset.columns) - 1; index >= 1; index-- {
		predicate, args = prependCursorColumn(cursorColumnKeyset{
			column: keyset.columns[index], value: keyset.values[index], tail: predicate,
			tailArgs: args, order: keyset.order,
		})
	}
	outer := cursorColumnKeyset{column: keyset.columns[0], value: keyset.values[0], tail: predicate, tailArgs: args, order: keyset.order}
	sameRank, sameRankArgs := cursorSameRank(outer)
	return "(" + nullRankExpression(keyset.columns[0]) + " = ? AND " + sameRank + ")", append([]any{cursorRank(keyset.values[0])}, sameRankArgs...)
}

func cursorTuplePredicate(keyset fieldDeviceKeyset) (string, []any) {
	if len(keyset.columns) > 1 {
		return multiColumnCursorTuplePredicate(keyset)
	}
	column := keyset.columns[0]
	operator := cursorOperator(keyset.order.order, keyset.order.direction)
	tuple := "(" + column.expression + "," + cursorIDExpression(keyset.idExpression) + ") " + operator + " (?,?)"
	predicate := "(" + nullRankExpression(column) + " = ? AND " + tuple + ")"
	return predicate, []any{0, *keyset.values[0], keyset.id}
}

func multiColumnCursorTuplePredicate(keyset fieldDeviceKeyset) (string, []any) {
	left := make([]string, 0, len(keyset.columns)*2+1)
	right := make([]string, 0, len(left))
	args := make([]any, 0, len(left))
	for index, column := range keyset.columns {
		left = append(left, nullRankExpression(column), column.expression)
		right = append(right, "?", "?")
		args = append(args, 0, *keyset.values[index])
	}
	left = append(left, cursorIDExpression(keyset.idExpression))
	right = append(right, "?")
	args = append(args, keyset.id)
	operator := cursorOperator(keyset.order.order, keyset.order.direction)
	return "(" + strings.Join(left, ",") + ") " + operator + " (" + strings.Join(right, ",") + ")", args
}

func allCursorValuesPresent(keyset fieldDeviceKeyset) bool {
	if len(keyset.columns) == 0 || len(keyset.columns) != len(keyset.values) {
		return false
	}
	for _, value := range keyset.values {
		if value == nil {
			return false
		}
	}
	return true
}

func cursorRank(value *string) int {
	if value == nil {
		return 1
	}
	return 0
}

func cursorRankOperator(keyset fieldDeviceKeyset) string {
	if keyset.order.direction == "previous" {
		return "<"
	}
	return ">"
}

func fieldDeviceKeysetPredicate(keyset fieldDeviceKeyset) (string, []any) {
	if allCursorValuesPresent(keyset) {
		return cursorTuplePredicate(keyset)
	}
	idOperator := cursorOperator(keyset.order.order, keyset.order.direction)
	predicate := cursorIDExpression(keyset.idExpression) + " " + idOperator + " ?"
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
	rank := cursorRank(value)
	sameRank, sameRankArgs := cursorSameRank(keyset)
	rankOperator := ">"
	if keyset.order.direction == "previous" {
		rankOperator = "<"
	}
	rankExpression := nullRankExpression(column)
	if cursorRankIsBoundary(rank, keyset.order.direction) {
		return "(" + rankExpression + " = ? AND " + sameRank + ")", append([]any{rank}, sameRankArgs...)
	}
	predicate := "(" + rankExpression + " " + rankOperator + " ? OR (" + rankExpression + " = ? AND " + sameRank + "))"
	return predicate, append([]any{rank, rank}, sameRankArgs...)
}

func cursorSameRank(keyset cursorColumnKeyset) (string, []any) {
	if keyset.value == nil {
		predicate := keyset.column.expression + " IS NULL AND (" + keyset.tail + ")"
		return predicate, append([]any(nil), keyset.tailArgs...)
	}
	operator := cursorOperator(keyset.order.order, keyset.order.direction)
	predicate := "(" + keyset.column.expression + " " + operator + " ? OR (" + keyset.column.expression + " = ? AND " + keyset.tail + "))"
	return predicate, append([]any{*keyset.value, *keyset.value}, keyset.tailArgs...)
}

func cursorRankIsBoundary(rank int, direction string) bool {
	return direction == "next" && rank == 1 || direction == "previous" && rank == 0
}

func cursorIDExpression(expression string) string {
	if expression == "" {
		return "field_devices.id"
	}
	return expression
}

func nullRankExpression(column fieldDeviceCursorColumn) string {
	if !column.nullable {
		return "0"
	}
	return "CASE WHEN " + column.expression + " IS NULL THEN 1 ELSE 0 END"
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
	return fieldDeviceCursorSort{key: key, columns: []fieldDeviceCursorColumn{{expression: expression, alias: "sort_value", nullable: true}}}
}

func requiredCursorSort(key, expression string) fieldDeviceCursorSort {
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
