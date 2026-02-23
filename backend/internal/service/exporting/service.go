package exporting

import (
	"context"
	"errors"
	"sync"
	"time"

	domainExport "github.com/besart951/go_infra_link/backend/internal/domain/exporting"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

var ErrJobNotFound = errors.New("export job not found")

type Config struct {
	QueueSize             int
	MaxConcurrent         int
	SingleFileDeviceLimit int64
	PageSize              int
}

type Service struct {
	data      domainExport.DataProvider
	workbook  domainExport.WorkbookGenerator
	zip       domainExport.ZipGenerator
	jobs      domainExport.JobStore
	files     domainExport.FileStore
	cfg       Config
	queue     chan uuid.UUID
	requests  map[uuid.UUID]domainExport.Request
	requestsM sync.RWMutex
}

func NewService(
	data domainExport.DataProvider,
	workbook domainExport.WorkbookGenerator,
	zip domainExport.ZipGenerator,
	jobs domainExport.JobStore,
	files domainExport.FileStore,
	cfg Config,
) *Service {
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = 100
	}
	if cfg.MaxConcurrent <= 0 {
		cfg.MaxConcurrent = 1
	}
	if cfg.SingleFileDeviceLimit <= 0 {
		cfg.SingleFileDeviceLimit = 5000
	}
	if cfg.PageSize <= 0 {
		cfg.PageSize = 1000
	}

	s := &Service{
		data:     data,
		workbook: workbook,
		zip:      zip,
		jobs:     jobs,
		files:    files,
		cfg:      cfg,
		queue:    make(chan uuid.UUID, cfg.QueueSize),
		requests: map[uuid.UUID]domainExport.Request{},
	}

	for i := 0; i < cfg.MaxConcurrent; i++ {
		go s.worker()
	}

	return s
}

func (s *Service) Create(ctx context.Context, req domainExport.Request) (domainExport.Job, error) {
	now := time.Now().UTC()
	job := domainExport.Job{
		ID:        uuid.New(),
		Status:    domainExport.StatusQueued,
		Progress:  0,
		Message:   "queued",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.jobs.Create(ctx, job); err != nil {
		return domainExport.Job{}, err
	}

	s.requestsM.Lock()
	s.requests[job.ID] = req
	s.requestsM.Unlock()

	shouldQueue := req.ForceAsync || len(s.queue) > 0
	if shouldQueue {
		s.queue <- job.ID
		return s.jobs.Get(ctx, job.ID)
	}

	if err := s.process(ctx, job.ID); err != nil {
		if failErr := s.failJob(ctx, job.ID, err); failErr != nil {
			return domainExport.Job{}, failErr
		}
	}

	return s.jobs.Get(ctx, job.ID)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (domainExport.Job, error) {
	job, err := s.jobs.Get(ctx, id)
	if err != nil {
		return domainExport.Job{}, ErrJobNotFound
	}
	return job, nil
}

func (s *Service) worker() {
	for jobID := range s.queue {
		ctx := context.Background()
		if err := s.process(ctx, jobID); err != nil {
			if failErr := s.failJob(ctx, jobID, err); failErr != nil {
				s.deleteRequest(jobID)
			}
		}
	}
}

func (s *Service) process(ctx context.Context, jobID uuid.UUID) error {
	job, err := s.jobs.Get(ctx, jobID)
	if err != nil {
		return err
	}

	req, ok := s.getRequest(jobID)
	if !ok {
		return errors.New("job request payload missing")
	}

	job.Status = domainExport.StatusProcessing
	job.Progress = 5
	job.Message = "resolving controllers"
	job.UpdatedAt = time.Now().UTC()
	if err := s.jobs.Update(ctx, job); err != nil {
		return err
	}

	controllers, err := s.data.ResolveControllers(ctx, req)
	if err != nil {
		return err
	}

	perController := make(map[uuid.UUID][]domainFacility.FieldDevice, len(controllers))
	var total int64

	for idx, controller := range controllers {
		page := 1
		for {
			items, count, listErr := s.data.ListFieldDevicesByController(ctx, controller.ID, req, page, s.cfg.PageSize)
			if listErr != nil {
				return listErr
			}
			if page == 1 {
				total += count
			}
			if len(items) == 0 {
				break
			}
			perController[controller.ID] = append(perController[controller.ID], items...)
			if len(items) < s.cfg.PageSize {
				break
			}
			page++
		}

		job.Progress = 5 + ((idx + 1) * 55 / max(1, len(controllers)))
		job.Message = "collecting field devices"
		job.UpdatedAt = time.Now().UTC()
		if err := s.jobs.Update(ctx, job); err != nil {
			return err
		}
	}

	outputType := domainExport.OutputTypeExcel
	if total > s.cfg.SingleFileDeviceLimit {
		outputType = domainExport.OutputTypeZip
	}

	filePath, fileName := s.files.BuildOutputPath(jobID, outputType)
	job.OutputType = outputType
	job.FilePath = filePath
	job.FileName = fileName
	job.Message = "generating file"
	job.Progress = 75
	job.UpdatedAt = time.Now().UTC()
	if err := s.jobs.Update(ctx, job); err != nil {
		return err
	}

	if outputType == domainExport.OutputTypeZip {
		if err := s.zip.GenerateZipByCabinet(ctx, filePath, controllers, perController); err != nil {
			return err
		}
		job.ContentType = "application/zip"
	} else {
		if err := s.workbook.GenerateWorkbook(ctx, filePath, controllers, perController); err != nil {
			return err
		}
		job.ContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	}

	job.Status = domainExport.StatusCompleted
	job.Progress = 100
	job.Message = "completed"
	job.UpdatedAt = time.Now().UTC()
	if err := s.jobs.Update(ctx, job); err != nil {
		return err
	}

	s.deleteRequest(jobID)
	return nil
}

func (s *Service) failJob(ctx context.Context, jobID uuid.UUID, processErr error) error {
	job, err := s.jobs.Get(ctx, jobID)
	if err != nil {
		return err
	}
	job.Status = domainExport.StatusFailed
	job.Error = processErr.Error()
	job.Message = "failed"
	job.UpdatedAt = time.Now().UTC()
	s.deleteRequest(jobID)
	return s.jobs.Update(ctx, job)
}

func (s *Service) getRequest(jobID uuid.UUID) (domainExport.Request, bool) {
	s.requestsM.RLock()
	defer s.requestsM.RUnlock()
	req, ok := s.requests[jobID]
	return req, ok
}

func (s *Service) deleteRequest(jobID uuid.UUID) {
	s.requestsM.Lock()
	defer s.requestsM.Unlock()
	delete(s.requests, jobID)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
