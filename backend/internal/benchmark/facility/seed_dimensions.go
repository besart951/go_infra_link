package facilitybenchmark

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (d *Database) seedBuildings(ctx context.Context) error {
	rows := dimensionRows(100, func(index int64) []any {
		return append(baseValues(1, index), fmt.Sprintf("B%03d", index), int(index%10)+1)
	})
	return d.copy(ctx, "buildings", baseColumns("iws_code", "building_group"), pgx.CopyFromRows(rows))
}

func (d *Database) seedSystemTypes(ctx context.Context) error {
	rows := dimensionRows(10, func(index int64) []any {
		return append(baseValues(2, index), int(index*10+1), int(index*10+10), fmt.Sprintf("System %02d", index))
	})
	return d.copy(ctx, "system_types", baseColumns("number_min", "number_max", "name"), pgx.CopyFromRows(rows))
}

func (d *Database) seedSystemParts(ctx context.Context) error {
	rows := dimensionRows(100, func(index int64) []any {
		return append(baseValues(3, index), fmt.Sprintf("P%02d", index), fmt.Sprintf("System part %03d", index), nil)
	})
	return d.copy(ctx, "system_parts", baseColumns("short_name", "name", "description"), pgx.CopyFromRows(rows))
}

func (d *Database) seedApparats(ctx context.Context) error {
	rows := dimensionRows(100, func(index int64) []any {
		return append(baseValues(4, index), fmt.Sprintf("A%02d", index), fmt.Sprintf("Apparat %03d", index), nil)
	})
	return d.copy(ctx, "apparats", baseColumns("short_name", "name", "description"), pgx.CopyFromRows(rows))
}

func (d *Database) seedCabinets(ctx context.Context) error {
	rows := dimensionRows(CabinetCount, func(index int64) []any {
		return append(baseValues(5, index), deterministicID(1, index%100), fmt.Sprintf("S%04d", index))
	})
	return d.copy(ctx, "control_cabinets", baseColumns("building_id", "control_cabinet_nr"), pgx.CopyFromRows(rows))
}

func (d *Database) seedControllers(ctx context.Context) error {
	rows := dimensionRows(ControllerCount, func(index int64) []any {
		return append(baseValues(6, index), deterministicID(5, index%CabinetCount), fmt.Sprintf("GA%05d", index), fmt.Sprintf("Controller %05d", index), nil, nil, nil, nil, nil, nil)
	})
	columns := baseColumns("control_cabinet_id", "ga_device", "device_name", "device_description", "device_location", "ip_address", "subnet", "gateway", "vlan")
	return d.copy(ctx, "sps_controllers", columns, pgx.CopyFromRows(rows))
}

func (d *Database) seedSystemScopes(ctx context.Context) error {
	rows := dimensionRows(SystemScopeCount, func(index int64) []any {
		var document any = fmt.Sprintf("DOC-%06d", index)
		if index%7 == 0 {
			document = nil
		}
		return append(baseValues(7, index), int(index%1000)+1, document, deterministicID(6, index%ControllerCount), deterministicID(2, index%10))
	})
	columns := baseColumns("number", "document_name", "sps_controller_id", "system_type_id")
	return d.copy(ctx, "sps_controller_system_types", columns, pgx.CopyFromRows(rows))
}

func (d *Database) seedProjects(ctx context.Context) error {
	rows := dimensionRows(ProjectCount, func(index int64) []any {
		return append(baseValues(11, index), fmt.Sprintf("Project %03d", index), "benchmark", "active", nil, deterministicID(12, 0), deterministicID(13, 0))
	})
	columns := baseColumns("name", "description", "status", "start_date", "phase_id", "creator_id")
	return d.copy(ctx, "projects", columns, pgx.CopyFromRows(rows))
}

func dimensionRows(count int64, build func(int64) []any) [][]any {
	rows := make([][]any, count)
	for index := int64(0); index < count; index++ {
		rows[index] = build(index)
	}
	return rows
}
