package facilitybenchmark

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/config"
	appdb "github.com/besart951/go_infra_link/backend/internal/db"
	"github.com/jackc/pgx/v5"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	Gorm             *gorm.DB
	PGX              *pgx.Conn
	FieldDeviceCount int64
	Samples          SampleProfile
}

func Open(ctx context.Context, dsn string) (*Database, error) {
	if err := requireBenchmarkDatabase(dsn); err != nil {
		return nil, err
	}
	gormDB, err := appdb.Connect(config.DBConfig{Type: "postgres", Dsn: dsn})
	if err != nil {
		return nil, err
	}
	gormDB = gormDB.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})
	connection, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return &Database{Gorm: gormDB, PGX: connection, FieldDeviceCount: FieldDeviceCount, Samples: FullSampleProfile()}, nil
}

func (d *Database) SetFieldDeviceCount(count int64) error {
	if count < 50_000 || count > FieldDeviceCount {
		return fmt.Errorf("field device count must be between 50000 and %d", FieldDeviceCount)
	}
	d.FieldDeviceCount = count
	return nil
}

func (d *Database) Close(ctx context.Context) {
	if d == nil {
		return
	}
	if d.PGX != nil {
		_ = d.PGX.Close(ctx)
	}
	if d.Gorm != nil {
		if sqlDB, err := d.Gorm.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}
}

func (d *Database) Migrate() error { return appdb.ApplyMigrations(d.Gorm) }

func requireBenchmarkDatabase(dsn string) error {
	parsed, err := url.Parse(dsn)
	if err != nil {
		return err
	}
	name := strings.TrimPrefix(parsed.Path, "/")
	if !strings.HasSuffix(strings.ToLower(name), "_benchmark") {
		return fmt.Errorf("refusing non-benchmark database %q; name must end in _benchmark", name)
	}
	return nil
}
