package project

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ChangeAction describes a semantic mutation in a project aggregate.
type ChangeAction string

const (
	ChangeCreated ChangeAction = "created"
	ChangeUpdated ChangeAction = "updated"
	ChangeDeleted ChangeAction = "deleted"
	ChangeCopied  ChangeAction = "copied"
	ChangeInvited ChangeAction = "invited"
	ChangeRemoved ChangeAction = "removed"
)

// AggregateKind identifies the project-owned aggregate changed by an event.
// It is deliberately separate from database table names.
type AggregateKind string

const (
	AggregateProject                 AggregateKind = "project"
	AggregateControlCabinet          AggregateKind = "control_cabinet"
	AggregateSPSController           AggregateKind = "sps_controller"
	AggregateSPSControllerSystemType AggregateKind = "sps_controller_system_type"
	AggregateFieldDevice             AggregateKind = "field_device"
	AggregateObjectData              AggregateKind = "object_data"
	AggregateProjectUser             AggregateKind = "project_user"
)

// Change is a durable, client-facing project event. It intentionally contains
// semantic data only; audit snapshots and database table names never cross this
// boundary.
type Change struct {
	EventID       uuid.UUID
	ProjectID     uuid.UUID
	Revision      uint64
	AggregateType AggregateKind
	AggregateID   *uuid.UUID
	Action        ChangeAction
	ActorID       *uuid.UUID
	ChangedFields []string
	OccurredAt    time.Time
}

// NewChange describes an event before its event ID and revision are allocated.
type NewChange struct {
	ProjectID     uuid.UUID
	AggregateType AggregateKind
	AggregateID   *uuid.UUID
	Action        ChangeAction
	ActorID       *uuid.UUID
	ChangedFields []string
	OccurredAt    time.Time
}

type ChangePage struct {
	ProjectID       uuid.UUID
	CurrentRevision uint64
	Events          []Change
	HasMore         bool
	ResetRequired   bool
}

// ChangeStore is defined in the consuming project domain. Implementations own
// revision allocation and retention details.
type ChangeStore interface {
	Append(ctx context.Context, change NewChange) (*Change, error)
	ListAfter(ctx context.Context, projectID uuid.UUID, afterRevision uint64, limit int) (*ChangePage, error)
}

// BatchChangeStore is an optional stronger capability used for one semantic
// mutation that affects several aggregate IDs. Implementations must allocate
// consecutive revisions and insert the complete batch atomically.
type BatchChangeStore interface {
	AppendBatch(ctx context.Context, changes []NewChange) ([]Change, error)
}
