package facilitybenchmark

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

func (d *Database) Seed(ctx context.Context) error {
	if err := d.Migrate(); err != nil {
		return err
	}
	if err := d.resetBenchmarkData(ctx); err != nil {
		return err
	}
	if err := d.seedDimensions(ctx); err != nil {
		return err
	}
	return d.seedLargeTables(ctx)
}

func (d *Database) resetBenchmarkData(ctx context.Context) error {
	_, err := d.PGX.Exec(ctx, `TRUNCATE project_field_devices,specifications,field_devices,sps_controller_system_types,sps_controllers,control_cabinets,buildings,system_types,system_parts,apparats,projects CASCADE`)
	return err
}

func (d *Database) seedDimensions(ctx context.Context) error {
	if _, err := d.PGX.Exec(ctx, `SET session_replication_role = replica`); err != nil {
		return err
	}
	defer d.PGX.Exec(context.Background(), `SET session_replication_role = origin`)
	steps := []func(context.Context) error{d.seedBuildings, d.seedSystemTypes, d.seedSystemParts, d.seedApparats, d.seedCabinets, d.seedControllers, d.seedSystemScopes, d.seedProjects}
	for _, step := range steps {
		if err := step(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (d *Database) seedLargeTables(ctx context.Context) error {
	if _, err := d.PGX.Exec(ctx, `SET session_replication_role = replica`); err != nil {
		return err
	}
	defer d.PGX.Exec(context.Background(), `SET session_replication_role = origin`)
	for _, step := range []func(context.Context) error{
		d.seedFieldDevices,
		d.seedSpecifications,
		d.seedCursorValues,
		d.seedProjectLinks,
	} {
		if err := step(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (d *Database) copy(ctx context.Context, table string, columns []string, source pgx.CopyFromSource) error {
	count, err := d.PGX.CopyFrom(ctx, pgx.Identifier{table}, columns, source)
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("COPY %s inserted no rows", table)
	}
	return nil
}

func baseColumns(extra ...string) []string {
	return append([]string{"id", "created_at", "updated_at", "version"}, extra...)
}

func baseValues(kind byte, index int64) []any {
	moment := benchmarkEpoch.Add(time.Duration(index) * time.Microsecond)
	return []any{deterministicID(kind, index), moment, moment, int64(1)}
}
