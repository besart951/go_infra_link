package history

import (
	"context"
	"errors"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type globalHistoryStoreStub struct {
	listCalls    int
	restoreCalls int
}

func (s *globalHistoryStoreStub) ListTimeline(
	context.Context,
	domainHistory.TimelineFilter,
) (*domain.PaginatedList[domainHistory.ChangeEvent], error) {
	s.listCalls++
	return &domain.PaginatedList[domainHistory.ChangeEvent]{}, nil
}

func (*globalHistoryStoreStub) GetEvent(
	context.Context,
	uuid.UUID,
) (*domainHistory.ChangeEvent, error) {
	return &domainHistory.ChangeEvent{}, nil
}

func (s *globalHistoryStoreStub) RestoreEntityToEvent(
	context.Context,
	uuid.UUID,
	domainHistory.RestoreMode,
) (*domainHistory.RestoreResult, error) {
	s.restoreCalls++
	return &domainHistory.RestoreResult{}, nil
}

func (*globalHistoryStoreStub) RestoreControlCabinet(
	context.Context,
	uuid.UUID,
	domainHistory.RestoreControlCabinetRequest,
) (*domainHistory.RestoreResult, error) {
	return &domainHistory.RestoreResult{}, nil
}

type activeUserReaderStub struct {
	users []*domainUser.User
	err   error
}

func (s activeUserReaderStub) GetByIds(
	context.Context,
	[]uuid.UUID,
) ([]*domainUser.User, error) {
	return s.users, s.err
}

func TestGlobalHistoryAllowsOnlyActiveFZAGRoles(t *testing.T) {
	for _, role := range []domainUser.Role{
		domainUser.RoleSuperAdmin,
		domainUser.RoleAdminFZAG,
		domainUser.RoleFZAG,
	} {
		t.Run(string(role), func(t *testing.T) {
			actorID := uuid.New()
			store := &globalHistoryStoreStub{}
			service := NewGlobalHistoryService(GlobalHistoryDependencies{
				History: store,
				Users: activeUserReaderStub{users: []*domainUser.User{{
					Base:     domain.Base{ID: actorID},
					IsActive: true,
					Role:     role,
				}}},
				Actor: func(context.Context) *uuid.UUID { return &actorID },
			})

			if _, err := service.ListTimeline(
				context.Background(),
				domainHistory.TimelineFilter{},
			); err != nil {
				t.Fatalf("list global history as %s: %v", role, err)
			}
			if _, err := service.RestoreEntityToEvent(
				context.Background(),
				uuid.New(),
				domainHistory.RestoreModeAfter,
			); err != nil {
				t.Fatalf("restore global history as %s: %v", role, err)
			}
			if store.listCalls != 1 || store.restoreCalls != 1 {
				t.Fatalf("delegate calls: list=%d restore=%d", store.listCalls, store.restoreCalls)
			}
		})
	}
}

func TestGlobalHistoryRejectsMissingInactiveAndOtherRoles(t *testing.T) {
	actorID := uuid.New()
	for _, test := range []struct {
		name  string
		actor ActorProvider
		users []*domainUser.User
	}{
		{name: "missing actor"},
		{
			name:  "inactive",
			actor: func(context.Context) *uuid.UUID { return &actorID },
			users: []*domainUser.User{{
				Base: domain.Base{ID: actorID},
				Role: domainUser.RoleFZAG,
			}},
		},
		{
			name:  "project role",
			actor: func(context.Context) *uuid.UUID { return &actorID },
			users: []*domainUser.User{{
				Base:     domain.Base{ID: actorID},
				IsActive: true,
				Role:     domainUser.RoleAdminPlaner,
			}},
		},
		{
			name:  "missing user",
			actor: func(context.Context) *uuid.UUID { return &actorID },
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := &globalHistoryStoreStub{}
			service := NewGlobalHistoryService(GlobalHistoryDependencies{
				History: store,
				Users:   activeUserReaderStub{users: test.users},
				Actor:   test.actor,
			})

			_, err := service.ListTimeline(
				context.Background(),
				domainHistory.TimelineFilter{},
			)
			if !errors.Is(err, ErrGlobalHistoryAccessDenied) {
				t.Fatalf("access error: got %v, want %v", err, ErrGlobalHistoryAccessDenied)
			}
			if store.listCalls != 0 {
				t.Fatalf("denied request delegated %d calls", store.listCalls)
			}
		})
	}
}
