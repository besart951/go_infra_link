package facility

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type copyJobRecord struct {
	OwnerID     uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Kind        string     `gorm:"type:varchar(64);not null"`
	Class       string     `gorm:"type:varchar(32);not null;default:mutation;index"`
	JobType     string     `gorm:"column:job_type;type:varchar(32);not null;default:copy;index"`
	Status      string     `gorm:"type:varchar(32);not null;index"`
	Progress    int        `gorm:"not null;default:0"`
	Stage       string     `gorm:"type:varchar(96);not null"`
	Error       string     `gorm:"column:error_message;type:text"`
	WorkerID    *string    `gorm:"type:varchar(64);index"`
	LeaseUntil  *time.Time `gorm:"index"`
	Attempts    int        `gorm:"not null;default:0"`
	Task        string     `gorm:"type:varchar(96)"`
	Payload     []byte     `gorm:"type:jsonb"`
	Checkpoint  []byte     `gorm:"type:jsonb"`
	ResultID    *uuid.UUID `gorm:"type:uuid"`
	Result      []byte     `gorm:"type:jsonb"`
	Processed   int64      `gorm:"not null;default:0"`
	Total       *int64
	Succeeded   int64     `gorm:"not null;default:0"`
	Failed      int64     `gorm:"not null;default:0"`
	Retryable   bool      `gorm:"not null;default:true"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null;index"`
	CompletedAt *time.Time
}

type facilityJobItemRecord struct {
	OwnerID   uuid.UUID `gorm:"type:uuid;primaryKey"`
	JobID     uuid.UUID `gorm:"type:uuid;primaryKey"`
	Ordinal   int64     `gorm:"primaryKey"`
	Status    string    `gorm:"type:varchar(32);not null;index"`
	Input     []byte    `gorm:"type:jsonb;not null"`
	Result    []byte    `gorm:"type:jsonb"`
	Error     string    `gorm:"column:error_message;type:text"`
	Attempts  int       `gorm:"not null;default:0"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (facilityJobItemRecord) TableName() string {
	return "facility_job_items"
}

type facilityJobIDMappingRecord struct {
	OwnerID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	JobID      uuid.UUID `gorm:"type:uuid;primaryKey"`
	EntityType string    `gorm:"type:varchar(64);primaryKey"`
	SourceID   uuid.UUID `gorm:"type:uuid;primaryKey"`
	TargetID   uuid.UUID `gorm:"type:uuid;not null;index"`
	CreatedAt  time.Time `gorm:"not null"`
}

func (facilityJobIDMappingRecord) TableName() string {
	return "facility_job_id_mappings"
}

func (copyJobRecord) TableName() string {
	return "facility_jobs"
}

// MigrateCopyJobs creates the durable store used by asynchronous facility
// operations. It is exported only for the forward-only database migration.
func MigrateCopyJobs(db *gorm.DB) error {
	if err := db.AutoMigrate(&copyJobRecord{}, &facilityJobItemRecord{}, &facilityJobIDMappingRecord{}); err != nil {
		return err
	}
	if db.Dialector != nil && db.Dialector.Name() == "postgres" {
		if err := db.Exec(`DROP INDEX CONCURRENTLY IF EXISTS idx_facility_jobs_active_owner`).Error; err != nil {
			return err
		}
		return db.Exec(`
			CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_facility_jobs_active_mutation_owner
			ON facility_jobs (owner_id)
			WHERE class = 'mutation' AND status IN ('queued', 'running')
		`).Error
	}
	if err := db.Exec(`DROP INDEX IF EXISTS idx_facility_jobs_active_owner`).Error; err != nil {
		return err
	}
	return db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_facility_jobs_active_mutation_owner
		ON facility_jobs (owner_id)
		WHERE class = 'mutation' AND status IN ('queued', 'running')
	`).Error
}

type copyJobStore interface {
	CreateOrGetActive(context.Context, CopyJob) (CopyJob, bool, error)
	Get(context.Context, uuid.UUID, uuid.UUID) (CopyJob, error)
	List(context.Context, uuid.UUID, int) ([]CopyJob, error)
	ListPage(context.Context, uuid.UUID, *facilityJobCursor, int) ([]CopyJob, error)
	Claim(context.Context, uuid.UUID, uuid.UUID, string, time.Time, time.Time) (bool, error)
	ClaimNext(context.Context, FacilityJobClass, []string, string, time.Time, time.Time) (CopyJob, bool, error)
	Heartbeat(context.Context, uuid.UUID, uuid.UUID, string, time.Time, time.Time) error
	Save(context.Context, CopyJob, string) error
	Prune(context.Context, time.Time) error
	Retry(context.Context, uuid.UUID, uuid.UUID, time.Time) (CopyJob, error)
}

type sqlCopyJobStore struct {
	db *gorm.DB
}

func newSQLCopyJobStore(db *gorm.DB) copyJobStore {
	if db == nil {
		return nil
	}
	return &sqlCopyJobStore{db: db}
}

func (s *sqlCopyJobStore) CreateOrGetActive(ctx context.Context, candidate CopyJob) (CopyJob, bool, error) {
	var selected CopyJob
	created := false
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if tx.Dialector != nil && tx.Dialector.Name() == "postgres" {
			if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtextextended(?, 0))", candidate.OwnerID.String()).Error; err != nil {
				return err
			}
		}
		var record copyJobRecord
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("owner_id = ? AND id = ?", candidate.OwnerID, candidate.ID).
			First(&record).Error
		if err == nil {
			selected = record.toDomain()
			return nil
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("owner_id = ? AND class = ? AND status IN ?", candidate.OwnerID, candidate.Class, []string{
				string(CopyJobStatusQueued),
				string(CopyJobStatusRunning),
			}).
			Order("created_at ASC").
			First(&record).Error
		if err == nil && candidate.Class == FacilityJobClassMutation {
			selected = record.toDomain()
			return nil
		}
		if err == nil {
			var activeExports int64
			if countErr := tx.Model(&copyJobRecord{}).
				Where("owner_id = ? AND class = ? AND status IN ?", candidate.OwnerID, FacilityJobClassExport, []string{string(CopyJobStatusQueued), string(CopyJobStatusRunning)}).
				Count(&activeExports).Error; countErr != nil {
				return countErr
			}
			if activeExports >= 2 {
				return ErrFacilityJobLimit
			}
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		record = copyJobRecordFromDomain(candidate)
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		selected = candidate
		created = true
		return nil
	})
	if err == nil {
		return selected, created, nil
	}

	// PostgreSQL aborts a transaction after a unique violation. Resolve the
	// concurrent-winner case only after the failed transaction has rolled back.
	var record copyJobRecord
	lookupErr := s.db.WithContext(ctx).
		Where("owner_id = ? AND id = ?", candidate.OwnerID, candidate.ID).
		First(&record).Error
	if lookupErr == nil {
		return record.toDomain(), false, nil
	}
	lookupErr = s.db.WithContext(ctx).
		Where("owner_id = ? AND class = ? AND status IN ?", candidate.OwnerID, candidate.Class, []string{
			string(CopyJobStatusQueued),
			string(CopyJobStatusRunning),
		}).
		Order("created_at ASC").
		First(&record).Error
	if lookupErr == nil && candidate.Class == FacilityJobClassMutation {
		return record.toDomain(), false, nil
	}
	return CopyJob{}, false, err
}

func (s *sqlCopyJobStore) Get(ctx context.Context, ownerID, jobID uuid.UUID) (CopyJob, error) {
	var record copyJobRecord
	if err := s.db.WithContext(ctx).
		Where("owner_id = ? AND id = ?", ownerID, jobID).
		First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return CopyJob{}, ErrCopyJobNotFound
		}
		return CopyJob{}, err
	}
	return record.toDomain(), nil
}

func (s *sqlCopyJobStore) List(ctx context.Context, ownerID uuid.UUID, limit int) ([]CopyJob, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	var records []copyJobRecord
	if err := s.db.WithContext(ctx).
		Where("owner_id = ?", ownerID).
		Order("updated_at DESC, id DESC").
		Limit(limit).
		Find(&records).Error; err != nil {
		return nil, err
	}
	jobs := make([]CopyJob, len(records))
	for i := range records {
		jobs[i] = records[i].toDomain()
	}
	return jobs, nil
}

func (s *sqlCopyJobStore) ListPage(ctx context.Context, ownerID uuid.UUID, after *facilityJobCursor, limit int) ([]CopyJob, error) {
	if limit <= 0 || limit > 101 {
		limit = 51
	}
	query := s.db.WithContext(ctx).Where("owner_id = ?", ownerID)
	if after != nil {
		if after.Direction == "previous" {
			query = query.Where("updated_at > ? OR (updated_at = ? AND id > ?)", after.UpdatedAt, after.UpdatedAt, after.ID)
		} else {
			query = query.Where("updated_at < ? OR (updated_at = ? AND id < ?)", after.UpdatedAt, after.UpdatedAt, after.ID)
		}
	}
	var records []copyJobRecord
	order := "updated_at DESC, id DESC"
	if after != nil && after.Direction == "previous" {
		order = "updated_at ASC, id ASC"
	}
	if err := query.Order(order).Limit(limit).Find(&records).Error; err != nil {
		return nil, err
	}
	jobs := make([]CopyJob, len(records))
	for i := range records {
		jobs[i] = records[i].toDomain()
	}
	return jobs, nil
}

func (s *sqlCopyJobStore) Claim(
	ctx context.Context,
	ownerID, jobID uuid.UUID,
	workerID string,
	now, leaseUntil time.Time,
) (bool, error) {
	if s.db.Dialector != nil && s.db.Dialector.Name() == "postgres" {
		claimed := false
		err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			var record copyJobRecord
			err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
				Where("owner_id = ? AND id = ?", ownerID, jobID).
				Where(
					"status = ? OR (status = ? AND (lease_until IS NULL OR lease_until <= ?))",
					string(CopyJobStatusQueued),
					string(CopyJobStatusRunning),
					now,
				).
				First(&record).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			if err != nil {
				return err
			}
			result := tx.Model(&copyJobRecord{}).
				Where("owner_id = ? AND id = ?", ownerID, jobID).
				Updates(map[string]any{
					"status":      string(CopyJobStatusRunning),
					"worker_id":   workerID,
					"lease_until": leaseUntil,
					"attempts":    gorm.Expr("attempts + 1"),
					"updated_at":  now,
				})
			claimed = result.RowsAffected == 1
			return result.Error
		})
		return claimed, err
	}

	result := s.db.WithContext(ctx).Model(&copyJobRecord{}).
		Where("owner_id = ? AND id = ?", ownerID, jobID).
		Where(
			"status = ? OR (status = ? AND (lease_until IS NULL OR lease_until <= ?))",
			string(CopyJobStatusQueued),
			string(CopyJobStatusRunning),
			now,
		).
		Updates(map[string]any{
			"status":      string(CopyJobStatusRunning),
			"worker_id":   workerID,
			"lease_until": leaseUntil,
			"attempts":    gorm.Expr("attempts + 1"),
			"updated_at":  now,
		})
	return result.RowsAffected == 1, result.Error
}

func (s *sqlCopyJobStore) ClaimNext(
	ctx context.Context,
	class FacilityJobClass,
	tasks []string,
	workerID string,
	now, leaseUntil time.Time,
) (CopyJob, bool, error) {
	if len(tasks) == 0 {
		return CopyJob{}, false, nil
	}
	var claimed copyJobRecord
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		query := tx.Where("class = ? AND task IN ?", class, tasks).
			Where("status = ? OR (status = ? AND (lease_until IS NULL OR lease_until <= ?))", string(CopyJobStatusQueued), string(CopyJobStatusRunning), now).
			Order("created_at ASC")
		if tx.Dialector != nil && tx.Dialector.Name() == "postgres" {
			query = query.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"})
		}
		if err := query.First(&claimed).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		return tx.Model(&copyJobRecord{}).
			Where("owner_id = ? AND id = ?", claimed.OwnerID, claimed.ID).
			Updates(map[string]any{
				"status": string(CopyJobStatusRunning), "worker_id": workerID,
				"lease_until": leaseUntil, "attempts": gorm.Expr("attempts + 1"), "updated_at": now,
			}).Error
	})
	if err != nil || claimed.ID == uuid.Nil {
		return CopyJob{}, false, err
	}
	claimed.Status = string(CopyJobStatusRunning)
	claimed.WorkerID = &workerID
	claimed.LeaseUntil = &leaseUntil
	claimed.Attempts++
	claimed.UpdatedAt = now
	return claimed.toDomain(), true, nil
}

func (s *sqlCopyJobStore) Heartbeat(
	ctx context.Context,
	ownerID, jobID uuid.UUID,
	workerID string,
	now, leaseUntil time.Time,
) error {
	return s.db.WithContext(ctx).Model(&copyJobRecord{}).
		Where("owner_id = ? AND id = ? AND worker_id = ? AND status = ?", ownerID, jobID, workerID, string(CopyJobStatusRunning)).
		Updates(map[string]any{"lease_until": leaseUntil, "updated_at": now}).Error
}

func (s *sqlCopyJobStore) Save(ctx context.Context, job CopyJob, workerID string) error {
	updates := map[string]any{
		"status":        string(job.Status),
		"progress":      job.Progress,
		"stage":         job.Stage,
		"error_message": job.Error,
		"updated_at":    job.UpdatedAt,
		"checkpoint":    []byte(job.Checkpoint),
		"result":        []byte(job.Result),
		"processed":     job.Processed,
		"total":         job.Total,
		"succeeded":     job.Succeeded,
		"failed":        job.Failed,
		"retryable":     job.Retryable,
	}
	if job.IsTerminal() {
		updates["completed_at"] = job.UpdatedAt
		updates["lease_until"] = nil
		updates["worker_id"] = nil
	}
	query := s.db.WithContext(ctx).Model(&copyJobRecord{}).
		Where("owner_id = ? AND id = ?", job.OwnerID, job.ID)
	if workerID != "" {
		query = query.Where("worker_id = ?", workerID)
	}
	return query.Updates(updates).Error
}

func (s *sqlCopyJobStore) Retry(ctx context.Context, ownerID, jobID uuid.UUID, now time.Time) (CopyJob, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if tx.Dialector != nil && tx.Dialector.Name() == "postgres" {
			if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtextextended(?, 0))", ownerID.String()).Error; err != nil {
				return err
			}
		}
		var record copyJobRecord
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("owner_id = ? AND id = ?", ownerID, jobID).First(&record).Error; err != nil {
			return ErrFacilityJobNotRetryable
		}
		if record.Status != string(CopyJobStatusFailed) || !record.Retryable {
			return ErrFacilityJobNotRetryable
		}
		var active int64
		if err := tx.Model(&copyJobRecord{}).Where("owner_id = ? AND class = ? AND status IN ?", ownerID, record.Class, []string{string(CopyJobStatusQueued), string(CopyJobStatusRunning)}).Count(&active).Error; err != nil {
			return err
		}
		limit := int64(1)
		if FacilityJobClass(record.Class) == FacilityJobClassExport {
			limit = 2
		}
		if active >= limit {
			return ErrFacilityJobLimit
		}
		return tx.Model(&copyJobRecord{}).Where("owner_id = ? AND id = ?", ownerID, jobID).Updates(map[string]any{
			"status": string(CopyJobStatusQueued), "stage": copyJobStageQueued,
			"progress": 0, "error_message": "", "worker_id": nil,
			"lease_until": nil, "completed_at": nil, "updated_at": now,
		}).Error
	})
	if err != nil {
		return CopyJob{}, err
	}
	return s.Get(ctx, ownerID, jobID)
}

func (s *sqlCopyJobStore) Prune(ctx context.Context, before time.Time) error {
	return s.db.WithContext(ctx).
		Where("status IN ? AND updated_at < ?", []string{
			string(CopyJobStatusCompleted),
			string(CopyJobStatusFailed),
		}, before).
		Delete(&copyJobRecord{}).Error
}

func copyJobRecordFromDomain(job CopyJob) copyJobRecord {
	return copyJobRecord{
		OwnerID:    job.OwnerID,
		ID:         job.ID,
		Kind:       string(job.Kind),
		Class:      string(job.Class),
		JobType:    string(job.Type),
		Task:       job.Task,
		Payload:    []byte(job.Payload),
		Checkpoint: []byte(job.Checkpoint),
		Status:     string(job.Status),
		Progress:   job.Progress,
		Stage:      job.Stage,
		Error:      job.Error,
		Processed:  job.Processed,
		Total:      job.Total,
		Succeeded:  job.Succeeded,
		Failed:     job.Failed,
		Retryable:  job.Retryable,
		Result:     []byte(job.Result),
		CreatedAt:  job.CreatedAt,
		UpdatedAt:  job.UpdatedAt,
	}
}

func (r copyJobRecord) toDomain() CopyJob {
	return CopyJob{
		ID:          r.ID,
		OwnerID:     r.OwnerID,
		Kind:        CopyJobKind(r.Kind),
		Class:       FacilityJobClass(r.Class),
		Type:        FacilityJobType(r.JobType),
		Task:        r.Task,
		Payload:     append([]byte(nil), r.Payload...),
		Checkpoint:  append([]byte(nil), r.Checkpoint...),
		Status:      CopyJobStatus(r.Status),
		Progress:    r.Progress,
		Stage:       r.Stage,
		Error:       r.Error,
		Attempts:    r.Attempts,
		Processed:   r.Processed,
		Total:       r.Total,
		Succeeded:   r.Succeeded,
		Failed:      r.Failed,
		Retryable:   r.Retryable,
		Result:      append([]byte(nil), r.Result...),
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
		CompletedAt: r.CompletedAt,
	}
}
