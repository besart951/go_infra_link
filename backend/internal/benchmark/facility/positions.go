package facilitybenchmark

import (
	"context"
	"fmt"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/repository/facilitysql"
	"github.com/google/uuid"
)

type anchorSort struct {
	expression string
	second     string
	source     string
	id         string
	nullable   bool
}

var anchorSorts = map[string]anchorSort{
	"created_at":      fieldDeviceAnchor("field_devices.created_at", false),
	"apparat_nr":      fieldDeviceAnchor("field_devices.apparat_nr", true),
	"description":     fieldDeviceAnchor("field_devices.description", true),
	"sps_system_type": projectionAnchor("fdcv.sps_number", "fdcv.sps_document_name"),
	"spec_supplier":   projectionAnchor("fdcv.specification_supplier", ""),
	"spec_size":       projectionAnchor("fdcv.additional_info_size", ""),
	"spec_acdc":       projectionAnchor("fdcv.electrical_connection_acdc", ""),
	"spec_power":      projectionAnchor("fdcv.electrical_connection_power", ""),
}

func fieldDeviceAnchor(expression string, nullable bool) anchorSort {
	return anchorSort{expression: expression, source: "field_devices", id: "field_devices.id", nullable: nullable}
}

func projectionAnchor(expression, second string) anchorSort {
	return anchorSort{
		expression: expression, second: second, source: "field_device_cursor_values fdcv",
		id: "fdcv.field_device_id", nullable: true,
	}
}

func (d *Database) allScenarios(ctx context.Context) ([]Scenario, error) {
	scenarios := canonicalScenarios()
	positioned, err := d.positionedScenarios(ctx)
	if err != nil {
		return nil, err
	}
	return append(scenarios, positioned...), nil
}

func (d *Database) positionedScenarios(ctx context.Context) ([]Scenario, error) {
	result := make([]Scenario, 0, 80)
	for _, orderBy := range benchmarkSorts {
		for _, order := range []string{"asc", "desc"} {
			query := CursorQuery{Limit: 300, OrderBy: orderBy, Order: order}
			positioned, err := d.positionsForQuery(ctx, query)
			if err != nil {
				return nil, err
			}
			result = append(result, positioned...)
		}
	}
	return result, nil
}

func (d *Database) positionedScenarioByName(ctx context.Context, name string) ([]Scenario, bool, error) {
	for _, orderBy := range benchmarkSorts {
		for _, order := range []string{"asc", "desc"} {
			query := CursorQuery{Limit: 300, OrderBy: orderBy, Order: order}
			for _, position := range []int{25, 50, 75, 99} {
				for _, direction := range []string{"next", "previous"} {
					if name != positionedScenarioName(query, position, direction) {
						continue
					}
					scenario, err := d.positionedScenarioAt(ctx, query, position, direction)
					return []Scenario{scenario}, true, err
				}
			}
		}
	}
	return nil, false, nil
}

func (d *Database) positionedScenarioAt(ctx context.Context, query CursorQuery, position int, direction string) (Scenario, error) {
	anchor, err := d.anchorAt(ctx, query, position)
	if err != nil {
		return Scenario{}, err
	}
	return positionedScenario(query, anchor, position, direction)
}

func (d *Database) positionsForQuery(ctx context.Context, query CursorQuery) ([]Scenario, error) {
	result := make([]Scenario, 0, 8)
	for _, position := range []int{25, 50, 75, 99} {
		anchor, err := d.anchorAt(ctx, query, position)
		if err != nil {
			return nil, err
		}
		for _, direction := range []string{"next", "previous"} {
			scenario, err := positionedScenario(query, anchor, position, direction)
			if err != nil {
				return nil, err
			}
			result = append(result, scenario)
		}
	}
	return result, nil
}

type benchmarkAnchor struct {
	id     uuid.UUID
	value  *string
	second *string
}

func positionedScenario(query CursorQuery, anchor benchmarkAnchor, position int, direction string) (Scenario, error) {
	cursor, err := facilitysql.EncodeFieldDeviceBenchmarkCursor(query, direction, anchor.id, anchor.value, anchor.second)
	if err != nil {
		return Scenario{}, err
	}
	query.Cursor = cursor
	name := positionedScenarioName(query, position, direction)
	return Scenario{Name: name, Query: query}, nil
}

func positionedScenarioName(query CursorQuery, position int, direction string) string {
	return fmt.Sprintf("%s_%s_p%d_%s", query.OrderBy, query.Order, position, direction)
}

func (d *Database) anchorAt(ctx context.Context, query CursorQuery, percent int) (benchmarkAnchor, error) {
	sort := anchorSorts[query.OrderBy]
	secondSelect := "NULL::text"
	if sort.second != "" {
		secondSelect = "CAST(" + sort.second + " AS text)"
	}
	statement := anchorStatement(sort, secondSelect, query.Order)
	offset := (d.FieldDeviceCount - 1) * int64(percent) / 100
	var anchor benchmarkAnchor
	err := d.PGX.QueryRow(ctx, statement, offset).Scan(&anchor.id, &anchor.value, &anchor.second)
	return anchor, err
}

func anchorStatement(sort anchorSort, secondSelect, order string) string {
	orderBy := anchorOrder(sort, order)
	return fmt.Sprintf(`SELECT %s,CAST(%s AS text),%s FROM %s ORDER BY %s OFFSET $1 LIMIT 1`, sort.id, sort.expression, secondSelect, sort.source, orderBy)
}

func anchorOrder(sort anchorSort, order string) string {
	parts := make([]string, 0, 5)
	if sort.nullable {
		parts = append(parts, "CASE WHEN "+sort.expression+" IS NULL THEN 1 ELSE 0 END ASC")
	}
	parts = append(parts, sort.expression+" "+order)
	if sort.second != "" {
		parts = append(parts,
			"CASE WHEN "+sort.second+" IS NULL THEN 1 ELSE 0 END ASC",
			sort.second+" "+order,
		)
	}
	return strings.Join(append(parts, sort.id+" "+order), ",")
}
