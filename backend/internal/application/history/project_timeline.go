package history

import (
	"context"
	"errors"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
)

var (
	ErrProjectTimelineNotConfigured = errors.New(
		"project timeline query is not configured",
	)
	ErrProjectTimelineAccessDenied = errors.New(
		"project timeline access denied",
	)
)

type ProjectAccessScope interface {
	RequireProjectAccess(context.Context, uuid.UUID, uuid.UUID) error
}

type TimelineReader interface {
	ListTimeline(
		context.Context,
		domainHistory.TimelineFilter,
	) (*domain.PaginatedList[domainHistory.ChangeEvent], error)
}

type ActorProvider func(context.Context) *uuid.UUID

type ListProjectTimelineQuery struct {
	ProjectID uuid.UUID
	Filter    domainHistory.TimelineFilter
}

type ProjectTimelineDependencies struct {
	Access   ProjectAccessScope
	Timeline TimelineReader
	Actor    ActorProvider
}

type ProjectTimelineHandler struct {
	access   ProjectAccessScope
	timeline TimelineReader
	actor    ActorProvider
}

func NewProjectTimelineHandler(
	deps ProjectTimelineDependencies,
) *ProjectTimelineHandler {
	return &ProjectTimelineHandler{
		access:   deps.Access,
		timeline: deps.Timeline,
		actor:    deps.Actor,
	}
}

func (h *ProjectTimelineHandler) ListProjectTimeline(
	ctx context.Context,
	query ListProjectTimelineQuery,
) (*domain.PaginatedList[domainHistory.ChangeEvent], error) {
	if h == nil || h.access == nil || h.timeline == nil {
		return nil, ErrProjectTimelineNotConfigured
	}
	if query.ProjectID == uuid.Nil {
		return nil, domain.ErrInvalidArgument
	}
	actorID := actorFromContext(h.actor, ctx)
	if actorID == nil {
		return nil, ErrProjectTimelineAccessDenied
	}
	if err := h.access.RequireProjectAccess(ctx, *actorID, query.ProjectID); err != nil {
		return nil, err
	}

	filter := query.Filter
	if filter.ScopeType != "" && filter.ScopeID != uuid.Nil {
		filter.SecondaryScopeType = filter.ScopeType
		filter.SecondaryScopeID = filter.ScopeID
	}
	filter.ScopeType = "project"
	filter.ScopeID = query.ProjectID
	return h.timeline.ListTimeline(ctx, filter)
}

func actorFromContext(provider ActorProvider, ctx context.Context) *uuid.UUID {
	if provider == nil {
		return nil
	}
	actorID := provider(ctx)
	if actorID == nil || *actorID == uuid.Nil {
		return nil
	}
	clone := *actorID
	return &clone
}

type Services struct {
	ProjectTimeline *ProjectTimelineHandler
	Global          *GlobalHistoryService
}
