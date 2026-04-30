package searchspec

import (
	"fmt"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
)

type Spec struct {
	Table   string
	Prefix  string
	Columns []Column
}

type Column struct {
	Name       string
	Expression string
}

func (s Spec) SearchColumns(qualifier string) []gormbase.TrigramSearchColumn {
	out := make([]gormbase.TrigramSearchColumn, 0, len(s.Columns))
	for _, column := range s.Columns {
		out = append(out, gormbase.SearchColumn(column.QualifiedExpression(qualifier)))
	}
	return out
}

func (s Spec) NamedSearchColumns(qualifier string, names ...string) []gormbase.TrigramSearchColumn {
	wanted := make(map[string]struct{}, len(names))
	for _, name := range names {
		wanted[name] = struct{}{}
	}

	out := make([]gormbase.TrigramSearchColumn, 0, len(wanted))
	for _, column := range s.Columns {
		if _, ok := wanted[column.Name]; !ok {
			continue
		}
		out = append(out, gormbase.SearchColumn(column.QualifiedExpression(qualifier)))
	}
	return out
}

func (s Spec) IndexStatements() []string {
	statements := make([]string, 0, len(s.Columns))
	for _, column := range s.Columns {
		statements = append(statements, fmt.Sprintf(
			"CREATE INDEX IF NOT EXISTS idx_%s_%s_trgm ON %s USING gin (LOWER(%s) gin_trgm_ops)",
			s.Prefix,
			column.Name,
			s.Table,
			column.Expression,
		))
	}
	return statements
}

func (c Column) QualifiedExpression(qualifier string) string {
	if qualifier == "" {
		return c.Expression
	}

	result := c.Expression
	for _, identifier := range extractIdentifiers(c.Expression) {
		result = strings.ReplaceAll(result, identifier, qualifier+identifier)
	}
	return result
}

func extractIdentifiers(expression string) []string {
	seen := map[string]struct{}{}
	var identifiers []string

	for _, token := range strings.FieldsFunc(expression, func(r rune) bool {
		return !(r == '_' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9')
	}) {
		if token == "" || token[0] >= '0' && token[0] <= '9' {
			continue
		}
		if _, ok := sqlKeywords[strings.ToLower(token)]; ok {
			continue
		}
		if _, ok := seen[token]; ok {
			continue
		}
		seen[token] = struct{}{}
		identifiers = append(identifiers, token)
	}
	return identifiers
}

var sqlKeywords = map[string]struct{}{
	"as":   {},
	"cast": {},
	"text": {},
}

var (
	Users = Spec{Table: "users", Prefix: "users", Columns: []Column{
		{Name: "first_name", Expression: "first_name"},
		{Name: "last_name", Expression: "last_name"},
		{Name: "full_name", Expression: "first_name || ' ' || last_name"},
		{Name: "email", Expression: "email"},
	}}
	Permissions = Spec{Table: "permissions", Prefix: "permissions", Columns: []Column{
		{Name: "name", Expression: "name"},
		{Name: "description", Expression: "description"},
		{Name: "resource", Expression: "resource"},
		{Name: "action", Expression: "action"},
	}}
	Teams = Spec{Table: "teams", Prefix: "teams", Columns: []Column{
		{Name: "name", Expression: "name"},
		{Name: "description", Expression: "description"},
	}}
	Projects = Spec{Table: "projects", Prefix: "projects", Columns: []Column{
		{Name: "name", Expression: "name"},
		{Name: "description", Expression: "description"},
	}}
	Phases = Spec{Table: "phases", Prefix: "phases", Columns: []Column{
		{Name: "name", Expression: "name"},
	}}
	Buildings = Spec{Table: "buildings", Prefix: "buildings", Columns: []Column{
		{Name: "iws_code", Expression: "iws_code"},
		{Name: "group_text", Expression: "CAST(building_group AS TEXT)"},
		{Name: "label", Expression: "iws_code || '-' || CAST(building_group AS TEXT)"},
	}}
	ControlCabinets = Spec{Table: "control_cabinets", Prefix: "control_cabinets", Columns: []Column{
		{Name: "nr", Expression: "control_cabinet_nr"},
	}}
	SPSControllers = Spec{Table: "sps_controllers", Prefix: "sps_controllers", Columns: []Column{
		{Name: "device_name", Expression: "device_name"},
		{Name: "ip_address", Expression: "ip_address"},
	}}
	SPSControllerSystemTypes = Spec{Table: "sps_controller_system_types", Prefix: "sps_controller_system_types", Columns: []Column{
		{Name: "doc", Expression: "document_name"},
	}}
	SystemTypes = Spec{Table: "system_types", Prefix: "system_types", Columns: []Column{
		{Name: "name", Expression: "name"},
	}}
	SystemParts = Spec{Table: "system_parts", Prefix: "system_parts", Columns: []Column{
		{Name: "short_name", Expression: "short_name"},
		{Name: "name", Expression: "name"},
	}}
	Apparats = Spec{Table: "apparats", Prefix: "apparats", Columns: []Column{
		{Name: "short_name", Expression: "short_name"},
		{Name: "name", Expression: "name"},
		{Name: "description", Expression: "description"},
	}}
	Specifications = Spec{Table: "specifications", Prefix: "specifications", Columns: []Column{
		{Name: "supplier", Expression: "specification_supplier"},
		{Name: "brand", Expression: "specification_brand"},
		{Name: "type", Expression: "specification_type"},
	}}
	FieldDevices = Spec{Table: "field_devices", Prefix: "field_devices", Columns: []Column{
		{Name: "bmk", Expression: "bmk"},
		{Name: "description", Expression: "description"},
	}}
	ObjectData = Spec{Table: "object_data", Prefix: "object_data", Columns: []Column{
		{Name: "description", Expression: "description"},
	}}
	StateTexts = Spec{Table: "state_texts", Prefix: "state_texts", Columns: []Column{
		{Name: "state_text1", Expression: "state_text1"},
	}}
	NotificationClasses = Spec{Table: "notification_classes", Prefix: "notification_classes", Columns: []Column{
		{Name: "object_desc", Expression: "object_description"},
		{Name: "event_category", Expression: "event_category"},
		{Name: "meaning", Expression: "meaning"},
	}}
	BacnetObjects = Spec{Table: "bacnet_objects", Prefix: "bacnet_objects", Columns: []Column{
		{Name: "text_fix", Expression: "text_fix"},
	}}
	AlarmDefinitions = Spec{Table: "alarm_definitions", Prefix: "alarm_definitions", Columns: []Column{
		{Name: "name", Expression: "name"},
	}}
	Units = Spec{Table: "units", Prefix: "units", Columns: []Column{
		{Name: "name", Expression: "name"},
		{Name: "code", Expression: "code"},
	}}
	AlarmFields = Spec{Table: "alarm_fields", Prefix: "alarm_fields", Columns: []Column{
		{Name: "label", Expression: "label"},
		{Name: "key", Expression: "key"},
	}}
	AlarmTypes = Spec{Table: "alarm_types", Prefix: "alarm_types", Columns: []Column{
		{Name: "name", Expression: "name"},
		{Name: "code", Expression: "code"},
	}}
)

var All = []Spec{
	Users,
	Permissions,
	Teams,
	Projects,
	Phases,
	Buildings,
	ControlCabinets,
	SPSControllers,
	SPSControllerSystemTypes,
	SystemTypes,
	SystemParts,
	Apparats,
	Specifications,
	FieldDevices,
	ObjectData,
	StateTexts,
	NotificationClasses,
	BacnetObjects,
	AlarmDefinitions,
	Units,
	AlarmFields,
	AlarmTypes,
}
