package exporting

import (
	"context"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type DataProvider interface {
	ResolveControllers(ctx context.Context, req Request) ([]Controller, error)
	ListFieldDevicesByControllerAfter(ctx context.Context, controllerID uuid.UUID, req Request, afterID uuid.UUID, limit int) ([]domainFacility.FieldDevice, error)
}

type SnapshotDataProvider interface {
	DataProvider
	WithinSnapshot(ctx context.Context, consume func(DataProvider) error) error
}

// WorkbookBuild groups generator inputs so callers do not depend on a long
// positional parameter list.
type WorkbookBuild struct {
	OutputPath  string
	Controllers []Controller
	Source      DataProvider
	Request     Request
	PageSize    int
}

type WorkbookGenerator interface {
	GenerateWorkbook(ctx context.Context, build WorkbookBuild) (int64, error)
}

type ZipGenerator interface {
	GenerateZipByCabinet(ctx context.Context, build WorkbookBuild) (int64, error)
}

type FileStore interface {
	BuildOutputPath(jobID uuid.UUID, outputType OutputType, downloadFileName string) (string, string)
	BuildStagingPath(jobID uuid.UUID, outputType OutputType) string
	Finalize(stagingPath, outputPath string) error
	Remove(path string) error
	SnapshotDirectory(jobID uuid.UUID) string
}
