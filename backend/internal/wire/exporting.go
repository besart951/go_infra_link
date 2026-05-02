package wire

import (
	"fmt"
	"path/filepath"

	exportinfra "github.com/besart951/go_infra_link/backend/internal/infrastructure/exporting"
	exportservice "github.com/besart951/go_infra_link/backend/internal/service/exporting"
)

func newExportService(repos *Repositories) (*exportservice.Service, error) {
	jobStore := exportinfra.NewMemoryJobStore()
	fileStore, err := exportinfra.NewLocalFileStore(filepath.Join("data", "exports"))
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
		exportservice.Config{
			QueueSize:             200,
			MaxConcurrent:         1,
			SingleFileDeviceLimit: 5000,
			PageSize:              1000,
		},
	), nil
}
