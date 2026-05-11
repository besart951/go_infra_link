package wire

import (
	"fmt"
	"path/filepath"
	"strings"

	exportservice "github.com/besart951/go_infra_link/backend/internal/service/exporting"

	exportinfra "github.com/besart951/go_infra_link/backend/internal/infrastructure/exporting"
)

func newExportService(repos *Repositories, cfg ServiceConfig) (*exportservice.Service, error) {
	jobStore := exportinfra.NewMemoryJobStore()
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
	)
	excelGenerator := exportinfra.NewExcelizeGenerator()
	return exportservice.NewService(
		dataProvider,
		excelGenerator,
		excelGenerator,
		jobStore,
		fileStore,
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
		PageSize:              1000,
	}
}

func defaultExportDirectory() string {
	return filepath.Join("data", "exports")
}
