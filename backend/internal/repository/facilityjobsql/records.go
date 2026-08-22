package facilityjobsql

import (
	"time"

	facilityjobs "github.com/besart951/go_infra_link/backend/internal/application/facilityjobs"
	"github.com/google/uuid"
)

type itemRecord struct {
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

func (itemRecord) TableName() string { return "facility_job_items" }

func itemRecordFromStep(step facilityjobs.Step, now time.Time) itemRecord {
	return itemRecord{
		OwnerID: step.Key.OwnerID, JobID: step.Key.JobID, Ordinal: step.Key.Ordinal,
		Status: string(facilityjobs.ItemStatusRunning), Input: []byte(step.Input), Attempts: 1,
		CreatedAt: now, UpdatedAt: now,
	}
}

func preparedItemRecord(step facilityjobs.Step, now time.Time) itemRecord {
	record := itemRecordFromStep(step, now)
	record.Status = string(facilityjobs.ItemStatusQueued)
	record.Attempts = 0
	return record
}

func (r itemRecord) toDomain() facilityjobs.Item {
	return facilityjobs.Item{
		Key:    facilityjobs.ItemKey{OwnerID: r.OwnerID, JobID: r.JobID, Ordinal: r.Ordinal},
		Status: facilityjobs.ItemStatus(r.Status), Input: append([]byte(nil), r.Input...),
		Result: append([]byte(nil), r.Result...), Error: r.Error, Attempts: r.Attempts,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

type mappingRecord struct {
	OwnerID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	JobID      uuid.UUID `gorm:"type:uuid;primaryKey"`
	EntityType string    `gorm:"type:varchar(64);primaryKey"`
	SourceID   uuid.UUID `gorm:"type:uuid;primaryKey"`
	TargetID   uuid.UUID `gorm:"type:uuid;not null;index"`
	CreatedAt  time.Time `gorm:"not null"`
}

func (mappingRecord) TableName() string { return "facility_job_id_mappings" }

func mappingRecordFromStep(step facilityjobs.Step, targetID uuid.UUID, now time.Time) mappingRecord {
	return mappingRecord{
		OwnerID: step.Key.OwnerID, JobID: step.Key.JobID, EntityType: step.EntityType,
		SourceID: step.SourceID, TargetID: targetID, CreatedAt: now,
	}
}

func (r mappingRecord) toDomain() facilityjobs.IDMapping {
	return facilityjobs.IDMapping{
		OwnerID: r.OwnerID, JobID: r.JobID, EntityType: r.EntityType,
		SourceID: r.SourceID, TargetID: r.TargetID, CreatedAt: r.CreatedAt,
	}
}
