package facilityjobs

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	"github.com/google/uuid"
)

type ItemStatus string

const (
	ItemStatusQueued    ItemStatus = "queued"
	ItemStatusRunning   ItemStatus = "running"
	ItemStatusCompleted ItemStatus = "completed"
	ItemStatusFailed    ItemStatus = "failed"
)

var ErrMappingConflict = errors.New("facility job ID mapping conflicts with persisted target")

type ItemKey struct {
	OwnerID uuid.UUID
	JobID   uuid.UUID
	Ordinal int64
}

type Step struct {
	Key              ItemKey
	EntityType       string
	SourceID         uuid.UUID
	Input            json.RawMessage
	PersistIDMapping bool
}

type StepResult struct {
	TargetID uuid.UUID
	Result   json.RawMessage
}

type Item struct {
	Key       ItemKey
	Status    ItemStatus
	Input     json.RawMessage
	Result    json.RawMessage
	Error     string
	Attempts  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type IDMapping struct {
	OwnerID    uuid.UUID
	JobID      uuid.UUID
	EntityType string
	SourceID   uuid.UUID
	TargetID   uuid.UUID
	CreatedAt  time.Time
}

type FieldDeviceUpdatePlanItem struct {
	OwnerID           uuid.UUID
	JobID             uuid.UUID
	Ordinal           int64
	GroupOrdinal      int64
	DependencyGroupID uuid.UUID
	FieldDeviceID     uuid.UUID
	Command           json.RawMessage
}

type FieldDeviceUpdatePlanStore interface {
	Save(context.Context, []FieldDeviceUpdatePlanItem) error
	Plan(context.Context, uuid.UUID, uuid.UUID) error
	List(context.Context, uuid.UUID, uuid.UUID) ([]FieldDeviceUpdatePlanItem, error)
}

type Mutation func(context.Context, apptransaction.UnitOfWork) (StepResult, error)

// StepStore makes one domain mutation, its source-to-target mapping, and the
// item completion indivisible. Completed steps are returned without rerunning
// the mutation after a worker restart.
type StepStore interface {
	Prepare(context.Context, []Step) error
	ListItems(context.Context, uuid.UUID, uuid.UUID) ([]Item, error)
	Execute(context.Context, Step, Mutation) (StepResult, bool, error)
	GetItem(context.Context, ItemKey) (Item, error)
	GetMapping(context.Context, Step) (IDMapping, error)
}
