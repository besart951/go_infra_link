package exporting

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	store := &LocalFileStore{baseDir: baseDir}
	if err := store.removeExpired(time.Now().UTC().Add(-90 * 24 * time.Hour)); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *LocalFileStore) removeExpired(before time.Time) error {
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		path := filepath.Join(s.baseDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() != "snapshots" {
				continue
			}
			snapshots, err := os.ReadDir(path)
			if err != nil {
				return err
			}
			for _, snapshot := range snapshots {
				snapshotInfo, err := snapshot.Info()
				if err != nil {
					return err
				}
				if snapshot.IsDir() && snapshotInfo.ModTime().Before(before) {
					if err := os.RemoveAll(filepath.Join(path, snapshot.Name())); err != nil {
						return err
					}
				}
			}
			continue
		}
		if info.ModTime().Before(before) {
			if err := os.Remove(path); err != nil {
				return err
			}
		}
	}
	return nil
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

func (s *LocalFileStore) BuildStagingPath(jobID uuid.UUID, outputType domainExport.OutputType) string {
	path, _ := s.BuildOutputPath(jobID, outputType, "")
	return path + ".partial"
}

func (s *LocalFileStore) Finalize(stagingPath, outputPath string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	return os.Rename(stagingPath, outputPath)
}

func (s *LocalFileStore) Remove(path string) error {
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (s *LocalFileStore) SnapshotDirectory(jobID uuid.UUID) string {
	return filepath.Join(s.baseDir, "snapshots", jobID.String())
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
