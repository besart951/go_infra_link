package project

import (
	"time"

	"github.com/google/uuid"
)

type ListProjectChangesQuery struct {
	AfterRevision uint64 `form:"after_revision"`
	Limit         int    `form:"limit" binding:"omitempty,min=1,max=500"`
}

type ProjectChangeResponse struct {
	EventID       uuid.UUID            `json:"event_id"`
	Revision      uint64               `json:"revision"`
	AggregateType string               `json:"aggregate_type"`
	AggregateID   *uuid.UUID           `json:"aggregate_id"`
	Action        string               `json:"action"`
	ActorID       *uuid.UUID           `json:"actor_id"`
	ChangedFields []string             `json:"changed_fields"`
	ParentRefs    map[string]uuid.UUID `json:"parent_refs"`
	OccurredAt    time.Time            `json:"occurred_at"`
}

type ProjectChangesResponse struct {
	ProjectID       uuid.UUID               `json:"project_id"`
	CurrentRevision uint64                  `json:"current_revision"`
	Events          []ProjectChangeResponse `json:"events"`
	HasMore         bool                    `json:"has_more"`
	ResetRequired   bool                    `json:"reset_required"`
}
