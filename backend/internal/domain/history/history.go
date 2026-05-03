package history

import (
	"time"

	"github.com/google/uuid"
)

type Action string

const (
	ActionCreate  Action = "create"
	ActionUpdate  Action = "update"
	ActionDelete  Action = "delete"
	ActionRestore Action = "restore"
)

type ChangeEvent struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	OccurredAt   time.Time  `gorm:"not null;index" json:"occurred_at"`
	ActorID      *uuid.UUID `gorm:"type:uuid;index" json:"actor_id,omitempty"`
	ActorName    *string    `gorm:"-" json:"actor_name,omitempty"`
	Action       Action     `gorm:"type:varchar(32);not null;index" json:"action"`
	EntityTable  string     `gorm:"type:varchar(96);not null;index:idx_change_events_entity,priority:1" json:"entity_table"`
	EntityID     uuid.UUID  `gorm:"type:uuid;not null;index:idx_change_events_entity,priority:2" json:"entity_id"`
	BatchID      *uuid.UUID `gorm:"type:uuid;index" json:"batch_id,omitempty"`
	Summary      *string    `gorm:"type:text" json:"summary,omitempty"`
	Scopes       []Scope    `gorm:"-" json:"scopes,omitempty"`
	BeforeJSON   JSONB      `gorm:"type:jsonb" json:"before_json,omitempty"`
	AfterJSON    JSONB      `gorm:"type:jsonb" json:"after_json,omitempty"`
	DiffJSON     JSONB      `gorm:"type:jsonb" json:"diff_json,omitempty"`
	MetadataJSON JSONB      `gorm:"type:jsonb" json:"metadata_json,omitempty"`
}

func (ChangeEvent) TableName() string {
	return "change_events"
}

type Scope struct {
	ScopeType string    `json:"scope_type"`
	ScopeID   uuid.UUID `json:"scope_id"`
	Label     *string   `json:"label,omitempty"`
}

type ChangeEventScope struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ChangeEventID uuid.UUID `gorm:"type:uuid;not null;index" json:"change_event_id"`
	ScopeType     string    `gorm:"type:varchar(96);not null;index:idx_change_event_scopes_lookup,priority:1" json:"scope_type"`
	ScopeID       uuid.UUID `gorm:"type:uuid;not null;index:idx_change_event_scopes_lookup,priority:2" json:"scope_id"`
	OccurredAt    time.Time `gorm:"not null;index:idx_change_event_scopes_lookup,priority:3,sort:desc" json:"occurred_at"`
}

func (ChangeEventScope) TableName() string {
	return "change_event_scopes"
}

type EntityVersion struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ChangeEventID uuid.UUID `gorm:"type:uuid;not null;index" json:"change_event_id"`
	EntityTable   string    `gorm:"type:varchar(96);not null;index:idx_entity_versions_entity,priority:1" json:"entity_table"`
	EntityID      uuid.UUID `gorm:"type:uuid;not null;index:idx_entity_versions_entity,priority:2" json:"entity_id"`
	VersionAt     time.Time `gorm:"not null;index:idx_entity_versions_entity,priority:3,sort:desc" json:"version_at"`
	Action        Action    `gorm:"type:varchar(32);not null" json:"action"`
	SnapshotJSON  JSONB     `gorm:"type:jsonb" json:"snapshot_json,omitempty"`
}

func (EntityVersion) TableName() string {
	return "entity_versions"
}

type TimelineFilter struct {
	ScopeType          string
	ScopeID            uuid.UUID
	SecondaryScopeType string
	SecondaryScopeID   uuid.UUID
	EntityTable        string
	EntityID           uuid.UUID
	Page               int
	Limit              int
}

type RestoreMode string

const (
	RestoreModeAfter  RestoreMode = "after"
	RestoreModeBefore RestoreMode = "before"
)

type RestoreEntityRequest struct {
	EventID uuid.UUID   `json:"event_id"`
	Mode    RestoreMode `json:"mode"`
}

type RestoreControlCabinetRequest struct {
	AsOf      *time.Time `json:"as_of"`
	EventID   *uuid.UUID `json:"event_id"`
	ProjectID *uuid.UUID `json:"project_id"`
}

type RestoreResult struct {
	RestoredCount int       `json:"restored_count"`
	DeletedCount  int       `json:"deleted_count"`
	SkippedCount  int       `json:"skipped_count"`
	BatchID       uuid.UUID `json:"batch_id"`
}
