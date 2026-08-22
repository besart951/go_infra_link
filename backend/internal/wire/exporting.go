package wire

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

	domainExport "github.com/besart951/go_infra_link/backend/internal/domain/exporting"
	exportservice "github.com/besart951/go_infra_link/backend/internal/service/exporting"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"

	exportinfra "github.com/besart951/go_infra_link/backend/internal/infrastructure/exporting"
	facilityrepo "github.com/besart951/go_infra_link/backend/internal/repository/facilitysql"
	"gorm.io/gorm"
)

func newExportService(db *gorm.DB, repos *Repositories, cfg ServiceConfig, jobs *facilityservice.FacilityJobManager) (*exportservice.Service, error) {
	fileStore, err := exportinfra.NewLocalFileStore(resolveExportDirectory(cfg))
	if err != nil {
		return nil, fmt.Errorf("export file store: %w", err)
	}
	dataProvider := exportinfra.NewDataProvider(
		repos.FacilityFieldDevices,
		repos.FacilitySpecifications,
		repos.FacilityBacnetObjects,
		repos.FacilitySPSControllers,
		repos.FacilityControlCabinet,
		repos.FacilityBacnetObjectAlarmValues,
	)
	if db != nil {
		dataProvider.SetSnapshotRunner(func(ctx context.Context, consume func(domainExport.DataProvider) error) error {
			return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
				txProvider := exportinfra.NewDataProvider(
					facilityrepo.NewFieldDeviceRepository(tx),
					facilityrepo.NewSpecificationRepository(tx),
					facilityrepo.NewBacnetObjectRepository(tx),
					facilityrepo.NewSPSControllerRepository(tx),
					facilityrepo.NewControlCabinetRepository(tx),
					facilityrepo.NewBacnetObjectAlarmValueRepository(tx),
				)
				return consume(txProvider)
			}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
		})
	}
	excelGenerator := exportinfra.NewExcelizeGenerator()
	return exportservice.NewService(
		dataProvider,
		excelGenerator,
		excelGenerator,
		fileStore,
		jobs,
		resolveExportConfig(cfg.Export),
	), nil
}

func resolveExportDirectory(cfg ServiceConfig) string {
	if strings.TrimSpace(cfg.ExportDirectory) == "" {
		return defaultExportDirectory()
	}
	return cfg.ExportDirectory
}

func resolveExportConfig(cfg exportservice.Config) exportservice.Config {
	resolved := defaultExportConfig()
	if cfg.QueueSize > 0 {
		resolved.QueueSize = cfg.QueueSize
	}
	if cfg.MaxConcurrent > 0 {
		resolved.MaxConcurrent = cfg.MaxConcurrent
	}
	if cfg.SingleFileDeviceLimit > 0 {
		resolved.SingleFileDeviceLimit = cfg.SingleFileDeviceLimit
	}
	if cfg.PageSize > 0 {
		resolved.PageSize = cfg.PageSize
	}
	return resolved
}

func defaultExportConfig() exportservice.Config {
	return exportservice.Config{
		QueueSize:             200,
		MaxConcurrent:         1,
		SingleFileDeviceLimit: 5000,
		PageSize:              500,
	}
}

func defaultExportDirectory() string {
	return filepath.Join("data", "exports")
}
