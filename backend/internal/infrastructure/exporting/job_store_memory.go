package exporting

import (
	"context"
	"errors"
	"sync"

	domainExport "github.com/besart951/go_infra_link/backend/internal/domain/exporting"
	"github.com/google/uuid"
)

var errJobMissing = errors.New("job not found")

type MemoryJobStore struct {
	mu   sync.RWMutex
	jobs map[uuid.UUID]domainExport.Job
}

func NewMemoryJobStore() *MemoryJobStore {
	return &MemoryJobStore{jobs: map[uuid.UUID]domainExport.Job{}}
}

func (s *MemoryJobStore) Create(_ context.Context, job domainExport.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[job.ID] = job
	return nil
}

func (s *MemoryJobStore) Update(_ context.Context, job domainExport.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.jobs[job.ID]; !ok {
		return errJobMissing
	}
	s.jobs[job.ID] = job
	return nil
}

func (s *MemoryJobStore) Get(_ context.Context, id uuid.UUID) (domainExport.Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	job, ok := s.jobs[id]
	if !ok {
		return domainExport.Job{}, errJobMissing
	}
	return job, nil
}
