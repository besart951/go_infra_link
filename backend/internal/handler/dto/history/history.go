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

func TimelineResponseFrom(list *domain.PaginatedList[domainHistory.ChangeEvent]) TimelineResponse {
	items := make([]ChangeEventResponse, len(list.Items))
	for i, event := range list.Items {
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
	return TimelineResponse{Items: items, Total: list.Total, Page: list.Page, TotalPages: list.TotalPages}
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
