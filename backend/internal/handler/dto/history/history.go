package history

import (
	"encoding/json"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	commondto "github.com/besart951/go_infra_link/backend/internal/handler/dto/common"
	"github.com/google/uuid"
)

type ErrorResponse = commondto.ErrorResponse

type ScopeResponse struct {
	ScopeType string    `json:"scope_type"`
	ScopeID   uuid.UUID `json:"scope_id"`
	Label     *string   `json:"label,omitempty"`
}

type ChangeEventResponse struct {
	ID           uuid.UUID       `json:"id"`
	OccurredAt   time.Time       `json:"occurred_at"`
	ActorID      *uuid.UUID      `json:"actor_id,omitempty"`
	ActorName    *string         `json:"actor_name,omitempty"`
	Action       string          `json:"action" enums:"create,update,delete,restore"`
	EntityTable  string          `json:"entity_table"`
	EntityID     uuid.UUID       `json:"entity_id"`
	BatchID      *uuid.UUID      `json:"batch_id,omitempty"`
	Summary      *string         `json:"summary,omitempty"`
	Scopes       []ScopeResponse `json:"scopes"`
	BeforeJSON   json.RawMessage `json:"before_json,omitempty" swaggertype:"object"`
	AfterJSON    json.RawMessage `json:"after_json,omitempty" swaggertype:"object"`
	DiffJSON     json.RawMessage `json:"diff_json,omitempty" swaggertype:"object"`
	MetadataJSON json.RawMessage `json:"metadata_json,omitempty" swaggertype:"object"`
}

type TimelineResponse struct {
	Items      []ChangeEventResponse `json:"items"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	TotalPages int                   `json:"total_pages"`
}

type TimelineCursorResponse struct {
	Items          []ChangeEventResponse `json:"items"`
	NextCursor     string                `json:"next_cursor,omitempty"`
	PreviousCursor string                `json:"previous_cursor,omitempty"`
}

type UndoConflictResponse struct {
	Code            string    `json:"code"`
	EntityTable     string    `json:"entity_table"`
	EntityID        uuid.UUID `json:"entity_id"`
	ExpectedVersion *uint64   `json:"expected_version,omitempty"`
	CurrentVersion  *uint64   `json:"current_version,omitempty"`
	Fields          []string  `json:"fields"`
}

func UndoConflictResponseFrom(conflict domainHistory.UndoConflict) UndoConflictResponse {
	return UndoConflictResponse{
		Code: conflict.Code, EntityTable: conflict.EntityTable, EntityID: conflict.EntityID,
		ExpectedVersion: conflict.ExpectedVersion, CurrentVersion: conflict.CurrentVersion,
		Fields: conflict.Fields,
	}
}

func TimelineResponseFrom(list *domain.PaginatedList[domainHistory.ChangeEvent]) TimelineResponse {
	return TimelineResponse{Items: changeEventResponses(list.Items), Total: list.Total, Page: list.Page, TotalPages: list.TotalPages}
}

func TimelineCursorResponseFrom(page *domainHistory.TimelineCursorPage) TimelineCursorResponse {
	return TimelineCursorResponse{
		Items:      changeEventResponses(page.Items),
		NextCursor: page.NextCursor, PreviousCursor: page.PreviousCursor,
	}
}

func changeEventResponses(events []domainHistory.ChangeEvent) []ChangeEventResponse {
	items := make([]ChangeEventResponse, len(events))
	for i, event := range events {
		items[i] = ChangeEventResponse{
			ID:           event.ID,
			OccurredAt:   event.OccurredAt,
			ActorID:      event.ActorID,
			ActorName:    event.ActorName,
			Action:       string(event.Action),
			EntityTable:  event.EntityTable,
			EntityID:     event.EntityID,
			BatchID:      event.BatchID,
			Summary:      event.Summary,
			Scopes:       scopesFrom(event.Scopes),
			BeforeJSON:   json.RawMessage(event.BeforeJSON),
			AfterJSON:    json.RawMessage(event.AfterJSON),
			DiffJSON:     json.RawMessage(event.DiffJSON),
			MetadataJSON: json.RawMessage(event.MetadataJSON),
		}
	}
	return items
}

func scopesFrom(scopes []domainHistory.Scope) []ScopeResponse {
	if len(scopes) == 0 {
		return []ScopeResponse{}
	}
	items := make([]ScopeResponse, len(scopes))
	for i, scope := range scopes {
		items[i] = ScopeResponse{ScopeType: scope.ScopeType, ScopeID: scope.ScopeID, Label: scope.Label}
	}
	return items
}
