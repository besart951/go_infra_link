package exporting

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	domainExport "github.com/besart951/go_infra_link/backend/internal/domain/exporting"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/google/uuid"
)

var ErrJobNotFound = errors.New("export job not found")

const fieldDeviceExportTask = "fielddevice.export.v1"

type Config struct {
	QueueSize             int
	MaxConcurrent         int
	SingleFileDeviceLimit int64
	PageSize              int
}

type exportResult struct {
	OutputType  domainExport.OutputType `json:"output_type"`
	FileName    string                  `json:"file_name"`
	ContentType string                  `json:"content_type"`
	DownloadURL string                  `json:"download_url"`
	Size        int64                   `json:"size"`
	ExpiresAt   time.Time               `json:"expires_at"`
}

type Service struct {
	data     domainExport.DataProvider
	workbook domainExport.WorkbookGenerator
	zip      domainExport.ZipGenerator
	files    domainExport.FileStore
	jobs     *facilityservice.FacilityJobManager
	cfg      Config
}

func NewService(
	data domainExport.DataProvider,
	workbook domainExport.WorkbookGenerator,
	zip domainExport.ZipGenerator,
	files domainExport.FileStore,
	jobs *facilityservice.FacilityJobManager,
	cfg Config,
) *Service {
	if cfg.PageSize <= 0 || cfg.PageSize > 500 {
		cfg.PageSize = 500
	}
	service := &Service{data: data, workbook: workbook, zip: zip, files: files, jobs: jobs, cfg: cfg}
	if jobs != nil {
		jobs.RegisterTask(fieldDeviceExportTask, facilityservice.FacilityJobHandlerFunc(service.run))
	}
	return service
}

func (s *Service) Create(ctx context.Context, ownerID, operationID uuid.UUID, req domainExport.Request) (domainExport.Job, error) {
	if s.jobs == nil {
		return domainExport.Job{}, errors.New("facility jobs unavailable")
	}
	if req.AccessScope == "" {
		req.AccessScope = domainExport.AccessScopeGlobal
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return domainExport.Job{}, fmt.Errorf("encode export request: %w", err)
	}
	job, err := s.jobs.SubmitTask(ctx, facilityservice.FacilityJob{
		ID: operationID, OwnerID: ownerID,
		Kind:  facilityservice.FacilityJobKindFieldDevice,
		Class: facilityservice.FacilityJobClassExport,
		Type:  facilityservice.FacilityJobTypeExport,
		Task:  fieldDeviceExportTask, Payload: payload,
	})
	if err != nil {
		return domainExport.Job{}, err
	}
	return s.toExportJob(job)
}

func (s *Service) Get(_ context.Context, ownerID, id uuid.UUID) (domainExport.Job, error) {
	if s.jobs == nil {
		return domainExport.Job{}, ErrJobNotFound
	}
	job, err := s.jobs.Get(ownerID, id)
	if err != nil || job.Type != facilityservice.FacilityJobTypeExport {
		return domainExport.Job{}, ErrJobNotFound
	}
	return s.toExportJob(job)
}

func (s *Service) run(ctx context.Context, execution facilityservice.FacilityJobExecution) (facilityservice.FacilityJobTaskResult, error) {
	job, report := execution.Job, execution.Reporter.Report
	var req domainExport.Request
	if err := json.Unmarshal(job.Payload, &req); err != nil {
		return facilityservice.FacilityJobTaskResult{}, fmt.Errorf("decode export request: %w", err)
	}
	report(facilityservice.FacilityJobProgress{Progress: 5, Stage: "snapshotting"})
	live, ok := s.data.(domainExport.SnapshotDataProvider)
	if !ok {
		return facilityservice.FacilityJobTaskResult{}, errors.New("export data provider does not support consistent snapshots")
	}
	lastProgressAt := time.Time{}
	snapshot, err := createOrOpenSnapshot(ctx, s.files.SnapshotDirectory(job.ID), req, live, s.cfg.PageSize, func(processed int64) {
		if time.Since(lastProgressAt) < time.Second {
			return
		}
		lastProgressAt = time.Now()
		report(facilityservice.FacilityJobProgress{Progress: 15, Stage: "snapshotting", Processed: processed})
	})
	if err != nil {
		return facilityservice.FacilityJobTaskResult{}, err
	}
	defer func() { _ = snapshot.Close() }()
	controllers := snapshot.manifest.Controllers
	snapshotTotal := snapshot.manifest.DeviceCount
	checkpoint, _ := json.Marshal(map[string]any{"snapshot_at": snapshot.manifest.SnapshotAt, "device_count": snapshotTotal})
	report(facilityservice.FacilityJobProgress{Progress: 25, Stage: "snapshotting", Processed: snapshotTotal, Total: &snapshotTotal, Checkpoint: checkpoint})
	req.SnapshotAt = snapshot.manifest.SnapshotAt
	req.SchemaVersion = exportSnapshotSchemaVersion
	req.DeviceCount = snapshotTotal
	req.Manifest = exportManifest(snapshot.manifest)

	outputType := domainExport.OutputTypeExcel
	if uniqueControlCabinetCount(controllers) > 1 {
		outputType = domainExport.OutputTypeZip
	}
	outputPath, fileName := s.files.BuildOutputPath(job.ID, outputType, exportDownloadFileName(outputType, controllers))
	stagingPath := s.files.BuildStagingPath(job.ID, outputType)
	_ = s.files.Remove(stagingPath)
	defer func() { _ = s.files.Remove(stagingPath) }()

	report(facilityservice.FacilityJobProgress{Progress: 30, Stage: "generating"})
	var processed int64
	if outputType == domainExport.OutputTypeZip {
		processed, err = s.zip.GenerateZipByCabinet(ctx, stagingPath, controllers, snapshot, req, s.cfg.PageSize)
	} else {
		processed, err = s.workbook.GenerateWorkbook(ctx, stagingPath, controllers, snapshot, req, s.cfg.PageSize)
	}
	if err != nil {
		return facilityservice.FacilityJobTaskResult{}, fmt.Errorf("generate export: %w", err)
	}
	total := processed
	report(facilityservice.FacilityJobProgress{Progress: 90, Stage: "packaging", Processed: processed, Total: &total, Succeeded: processed})
	if err := s.files.Finalize(stagingPath, outputPath); err != nil {
		return facilityservice.FacilityJobTaskResult{}, fmt.Errorf("finalize export: %w", err)
	}
	info, err := os.Stat(outputPath)
	if err != nil {
		return facilityservice.FacilityJobTaskResult{}, fmt.Errorf("stat export: %w", err)
	}
	contentType := "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	if outputType == domainExport.OutputTypeZip {
		contentType = "application/zip"
	}
	result, err := json.Marshal(exportResult{
		OutputType: outputType, FileName: fileName, ContentType: contentType,
		DownloadURL: "/api/v1/facility/jobs/" + job.ID.String() + "/download",
		Size:        info.Size(), ExpiresAt: time.Now().UTC().Add(90 * 24 * time.Hour),
	})
	if err != nil {
		return facilityservice.FacilityJobTaskResult{}, fmt.Errorf("encode export result: %w", err)
	}
	return facilityservice.FacilityJobTaskResult{Result: result}, nil
}

func exportManifest(snapshot exportSnapshotManifest) domainExport.Manifest {
	checksums := make(map[string]string, len(snapshot.Artifacts))
	for _, artifact := range snapshot.Artifacts {
		checksums[artifact.FileName] = artifact.SHA256
	}
	return domainExport.Manifest{
		Counts: snapshot.Counts, WorkbookShards: workbookShardNames(snapshot.Controllers), SnapshotChecksums: checksums,
	}
}

func workbookShardNames(controllers []domainExport.Controller) []string {
	unique := make(map[string]struct{})
	for _, controller := range controllers {
		unique[controller.ControlCabinetID.String()+".xlsx"] = struct{}{}
	}
	names := make([]string, 0, len(unique))
	for name := range unique {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (s *Service) toExportJob(job facilityservice.FacilityJob) (domainExport.Job, error) {
	status := domainExport.Status(job.Status)
	if job.Status == facilityservice.FacilityJobStatusRunning {
		status = domainExport.StatusProcessing
	}
	result := exportResult{}
	if len(job.Result) > 0 {
		if err := json.Unmarshal(job.Result, &result); err != nil {
			return domainExport.Job{}, fmt.Errorf("decode export result: %w", err)
		}
	}
	filePath := ""
	if result.OutputType != "" {
		filePath, _ = s.files.BuildOutputPath(job.ID, result.OutputType, result.FileName)
	}
	scope, err := exportScopeFromPayload(job.Payload)
	if err != nil {
		return domainExport.Job{}, err
	}
	return domainExport.Job{
		ID: job.ID, Status: status, Progress: job.Progress, Message: job.Stage,
		OutputType: result.OutputType, FileName: result.FileName, ContentType: result.ContentType,
		FilePath: filePath, Error: job.Error, CreatedAt: job.CreatedAt, UpdatedAt: job.UpdatedAt,
		Scope: scope,
	}, nil
}

func exportScopeFromPayload(payload json.RawMessage) (domainExport.Scope, error) {
	var request domainExport.Request
	if err := json.Unmarshal(payload, &request); err != nil {
		return domainExport.Scope{}, fmt.Errorf("decode export scope: %w", err)
	}
	scope := request.AccessScope
	if scope == "" {
		scope = domainExport.AccessScopeGlobal
	}
	return domainExport.Scope{Kind: scope, ProjectIDs: append([]uuid.UUID(nil), request.ProjectIDs...)}, nil
}

func exportDownloadFileName(outputType domainExport.OutputType, controllers []domainExport.Controller) string {
	if outputType != domainExport.OutputTypeExcel {
		return ""
	}
	for _, controller := range controllers {
		if name := strings.TrimSpace(controller.ControlCabinetNr); name != "" {
			return name
		}
	}
	return ""
}

func uniqueControlCabinetCount(controllers []domainExport.Controller) int {
	unique := make(map[uuid.UUID]struct{}, len(controllers))
	for _, controller := range controllers {
		unique[controller.ControlCabinetID] = struct{}{}
	}
	return len(unique)
}
