package importing

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	fielddeviceimport "github.com/besart951/go_infra_link/backend/internal/application/fielddeviceimport"
)

const workbookContentTypes = "[Content_Types].xml"

type ArchiveLimits struct {
	MaxUploadBytes int64
	MaxEntries     int
	MaxEntryBytes  uint64
	MaxTotalBytes  uint64
}

func DefaultArchiveLimits() ArchiveLimits {
	return ArchiveLimits{
		MaxUploadBytes: 20 << 30,
		MaxEntries:     10_000,
		MaxEntryBytes:  20 << 30,
		MaxTotalBytes:  100 << 30,
	}
}

type ArchiveReader struct {
	workbook fielddeviceimport.WorkbookReader
	limits   ArchiveLimits
}

func NewArchiveReader(workbook fielddeviceimport.WorkbookReader) ArchiveReader {
	return ArchiveReader{workbook: workbook, limits: DefaultArchiveLimits()}
}

func (r ArchiveReader) Read(ctx context.Context, source io.Reader, sink fielddeviceimport.Sink) (fielddeviceimport.Manifest, error) {
	path, err := r.spool(source)
	if err != nil {
		return fielddeviceimport.Manifest{}, err
	}
	defer os.Remove(path)
	archive, err := zip.OpenReader(path)
	if err != nil {
		return fielddeviceimport.Manifest{}, fmt.Errorf("open workbook archive: %w", err)
	}
	defer archive.Close()
	if isWorkbookArchive(archive.File) {
		return r.readWorkbookFile(ctx, path, sink)
	}
	return r.readWorkbookPackage(ctx, archive.File, sink)
}

func (r ArchiveReader) spool(source io.Reader) (string, error) {
	file, err := os.CreateTemp("", "facility-import-*.zip")
	if err != nil {
		return "", err
	}
	path := file.Name()
	written, copyErr := io.Copy(file, io.LimitReader(source, r.limits.MaxUploadBytes+1))
	closeErr := file.Close()
	if copyErr != nil || closeErr != nil || written > r.limits.MaxUploadBytes {
		_ = os.Remove(path)
		return "", errors.Join(copyErr, closeErr, uploadSizeError(written, r.limits.MaxUploadBytes))
	}
	return path, nil
}

func uploadSizeError(written, maximum int64) error {
	if written <= maximum {
		return nil
	}
	return fmt.Errorf("import exceeds maximum upload size of %d bytes", maximum)
}

func (r ArchiveReader) readWorkbookFile(ctx context.Context, path string, sink fielddeviceimport.Sink) (fielddeviceimport.Manifest, error) {
	file, err := os.Open(path)
	if err != nil {
		return fielddeviceimport.Manifest{}, err
	}
	defer file.Close()
	return r.workbook.Read(ctx, file, sink)
}

func (r ArchiveReader) readWorkbookPackage(ctx context.Context, files []*zip.File, sink fielddeviceimport.Sink) (fielddeviceimport.Manifest, error) {
	workbooks, err := r.packageWorkbooks(files)
	if err != nil {
		return fielddeviceimport.Manifest{}, err
	}
	var manifest fielddeviceimport.Manifest
	for index, workbook := range workbooks {
		current, readErr := r.readPackageEntry(ctx, workbook, sink)
		if readErr != nil {
			return manifest, fmt.Errorf("read workbook %s: %w", workbook.Name, readErr)
		}
		if index > 0 && !sameManifest(manifest, current) {
			return manifest, fmt.Errorf("workbook %s has an inconsistent export manifest", workbook.Name)
		}
		manifest = current
	}
	return manifest, nil
}

func (r ArchiveReader) packageWorkbooks(files []*zip.File) ([]*zip.File, error) {
	workbooks := make([]*zip.File, 0, len(files))
	var total uint64
	for _, file := range files {
		if file.FileInfo().IsDir() || !isWorkbookName(file.Name) {
			continue
		}
		total += file.UncompressedSize64
		if file.UncompressedSize64 > r.limits.MaxEntryBytes || total > r.limits.MaxTotalBytes {
			return nil, errors.New("workbook package exceeds import size limits")
		}
		workbooks = append(workbooks, file)
	}
	if len(workbooks) == 0 || len(workbooks) > r.limits.MaxEntries {
		return nil, fmt.Errorf("workbook package contains %d supported workbooks", len(workbooks))
	}
	sort.Slice(workbooks, func(i, j int) bool { return workbooks[i].Name < workbooks[j].Name })
	return workbooks, nil
}

func (r ArchiveReader) readPackageEntry(ctx context.Context, file *zip.File, sink fielddeviceimport.Sink) (fielddeviceimport.Manifest, error) {
	entry, err := file.Open()
	if err != nil {
		return fielddeviceimport.Manifest{}, err
	}
	defer entry.Close()
	return r.workbook.Read(ctx, entry, sink)
}

func isWorkbookArchive(files []*zip.File) bool {
	for _, file := range files {
		if file.Name == workbookContentTypes {
			return true
		}
	}
	return false
}

func isWorkbookName(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, ".xlsx") || strings.HasSuffix(lower, ".xlsm")
}

func sameManifest(left, right fielddeviceimport.Manifest) bool {
	return left.SchemaVersion == right.SchemaVersion &&
		left.SnapshotAt.Equal(right.SnapshotAt) && left.Scope == right.Scope &&
		left.DeviceCount == right.DeviceCount && left.Counts == right.Counts
}
