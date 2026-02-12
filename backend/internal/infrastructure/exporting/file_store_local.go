package exporting

import (
	"fmt"
	"os"
	"path/filepath"

	domainExport "github.com/besart951/go_infra_link/backend/internal/domain/exporting"
	"github.com/google/uuid"
)

type LocalFileStore struct {
	baseDir string
}

func NewLocalFileStore(baseDir string) (*LocalFileStore, error) {
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, err
	}
	return &LocalFileStore{baseDir: baseDir}, nil
}

func (s *LocalFileStore) BuildOutputPath(jobID uuid.UUID, outputType domainExport.OutputType) (string, string) {
	ext := ".xlsx"
	if outputType == domainExport.OutputTypeZip {
		ext = ".zip"
	}
	fileName := fmt.Sprintf("field-device-export-%s%s", jobID.String(), ext)
	return filepath.Join(s.baseDir, fileName), fileName
}
