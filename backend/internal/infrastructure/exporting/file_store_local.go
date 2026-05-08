package exporting

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

func (s *LocalFileStore) BuildOutputPath(jobID uuid.UUID, outputType domainExport.OutputType, downloadFileName string) (string, string) {
	ext := ".xlsx"
	if outputType == domainExport.OutputTypeZip {
		ext = ".zip"
	}

	storageFileName := fmt.Sprintf("field-device-export-%s%s", jobID.String(), ext)
	fileName := sanitizeDownloadFileName(downloadFileName, ext)
	if fileName == "" {
		fileName = storageFileName
	}

	return filepath.Join(s.baseDir, storageFileName), fileName
}

func sanitizeDownloadFileName(name string, ext string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}

	if currentExt := filepath.Ext(name); currentExt != "" {
		name = strings.TrimSuffix(name, currentExt)
	}

	invalid := []string{"\\", "/", "*", "?", ":", "[", "]", "<", ">", "|", "\""}
	for _, ch := range invalid {
		name = strings.ReplaceAll(name, ch, "-")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}

	if len(name) > 120 {
		name = name[:120]
	}

	return name + ext
}
