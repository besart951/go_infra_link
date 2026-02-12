package exporting

import (
	"context"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type DataProvider interface {
	ResolveControllers(ctx context.Context, req Request) ([]Controller, error)
	ListFieldDevicesByController(ctx context.Context, controllerID uuid.UUID, req Request, page, limit int) ([]domainFacility.FieldDevice, int64, error)
}

type WorkbookGenerator interface {
	GenerateWorkbook(ctx context.Context, outputPath string, controllers []Controller, perControllerDevices map[uuid.UUID][]domainFacility.FieldDevice) error
}

type ZipGenerator interface {
	GenerateZipByCabinet(ctx context.Context, outputPath string, controllers []Controller, perControllerDevices map[uuid.UUID][]domainFacility.FieldDevice) error
}

type JobStore interface {
	Create(ctx context.Context, job Job) error
	Update(ctx context.Context, job Job) error
	Get(ctx context.Context, id uuid.UUID) (Job, error)
}

type FileStore interface {
	BuildOutputPath(jobID uuid.UUID, outputType OutputType) (string, string)
}
