package history

import (
	"context"
	"errors"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

var (
	ErrGlobalHistoryNotConfigured = errors.New(
		"global history service is not configured",
	)
	ErrGlobalHistoryAccessDenied = errors.New(
		"global history access denied",
	)
)

type GlobalHistoryStore interface {
	ListTimeline(
		context.Context,
		domainHistory.TimelineFilter,
	) (*domain.PaginatedList[domainHistory.ChangeEvent], error)
	GetEvent(
		context.Context,
		uuid.UUID,
	) (*domainHistory.ChangeEvent, error)
	RestoreEntityToEvent(
		context.Context,
		uuid.UUID,
		domainHistory.RestoreMode,
	) (*domainHistory.RestoreResult, error)
	RestoreControlCabinet(
		context.Context,
		uuid.UUID,
		domainHistory.RestoreControlCabinetRequest,
	) (*domainHistory.RestoreResult, error)
}

type ActiveUserReader interface {
	GetByIds(context.Context, []uuid.UUID) ([]*domainUser.User, error)
}

type GlobalHistoryDependencies struct {
	History GlobalHistoryStore
	Users   ActiveUserReader
	Actor   ActorProvider
}

type GlobalHistoryService struct {
	history GlobalHistoryStore
	users   ActiveUserReader
	actor   ActorProvider
}

func NewGlobalHistoryService(
	deps GlobalHistoryDependencies,
) *GlobalHistoryService {
	return &GlobalHistoryService{
		history: deps.History,
		users:   deps.Users,
		actor:   deps.Actor,
	}
}

func (s *GlobalHistoryService) ListTimeline(
	ctx context.Context,
	filter domainHistory.TimelineFilter,
) (*domain.PaginatedList[domainHistory.ChangeEvent], error) {
	if err := s.requireAccess(ctx); err != nil {
		return nil, err
	}
	return s.history.ListTimeline(ctx, filter)
}

func (s *GlobalHistoryService) GetEvent(
	ctx context.Context,
	id uuid.UUID,
) (*domainHistory.ChangeEvent, error) {
	if err := s.requireAccess(ctx); err != nil {
		return nil, err
	}
	return s.history.GetEvent(ctx, id)
}

func (s *GlobalHistoryService) RestoreEntityToEvent(
	ctx context.Context,
	eventID uuid.UUID,
	mode domainHistory.RestoreMode,
) (*domainHistory.RestoreResult, error) {
	if err := s.requireAccess(ctx); err != nil {
		return nil, err
	}
	return s.history.RestoreEntityToEvent(ctx, eventID, mode)
}

func (s *GlobalHistoryService) RestoreControlCabinet(
	ctx context.Context,
	controlCabinetID uuid.UUID,
	request domainHistory.RestoreControlCabinetRequest,
) (*domainHistory.RestoreResult, error) {
	if err := s.requireAccess(ctx); err != nil {
		return nil, err
	}
	return s.history.RestoreControlCabinet(ctx, controlCabinetID, request)
}

func (s *GlobalHistoryService) requireAccess(ctx context.Context) error {
	if s == nil || s.history == nil || s.users == nil {
		return ErrGlobalHistoryNotConfigured
	}
	actorID := actorFromContext(s.actor, ctx)
	if actorID == nil {
		return ErrGlobalHistoryAccessDenied
	}
	users, err := s.users.GetByIds(ctx, []uuid.UUID{*actorID})
	if err != nil {
		return err
	}
	if len(users) != 1 || users[0] == nil || users[0].ID != *actorID {
		return ErrGlobalHistoryAccessDenied
	}
	user := users[0]
	if !user.IsActive || user.DisabledAt != nil || user.IsDeleted() ||
		user.IsAnonymized() || !globalHistoryRoleAllowed(user.Role) {
		return ErrGlobalHistoryAccessDenied
	}
	return nil
}

func globalHistoryRoleAllowed(role domainUser.Role) bool {
	switch role {
	case domainUser.RoleSuperAdmin, domainUser.RoleAdminFZAG, domainUser.RoleFZAG:
		return true
	default:
		return false
	}
}
