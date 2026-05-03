package history

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	domainTeam "github.com/besart951/go_infra_link/backend/internal/domain/team"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestRegisterRoutes_RequiresTimelineReadPermission(t *testing.T) {
	router, authz := setupHistoryRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/history/timeline", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden without timeline.read, got %d", res.Code)
	}
	if authz.lastPermission != domainUser.PermissionTimelineRead {
		t.Fatalf("expected %q check, got %q", domainUser.PermissionTimelineRead, authz.lastPermission)
	}

	authz.granted[domainUser.PermissionTimelineRead] = true
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected timeline.read to allow timeline listing, got %d", res.Code)
	}
}

func TestRegisterRoutes_RequiresTimelineRestorePermission(t *testing.T) {
	router, authz := setupHistoryRouter(t)
	eventID := uuid.NewString()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/history/events/"+eventID+"/restore", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden without timeline.restore, got %d", res.Code)
	}
	if authz.lastPermission != domainUser.PermissionTimelineRestore {
		t.Fatalf("expected %q check, got %q", domainUser.PermissionTimelineRestore, authz.lastPermission)
	}
}

func setupHistoryRouter(t *testing.T) (*gin.Engine, *historyAuthzStub) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	authz := &historyAuthzStub{granted: map[string]bool{}}
	router := gin.New()
	group := router.Group("/api/v1")
	group.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, uuid.New())
		c.Next()
	})
	RegisterRoutes(group, NewHandler(historyServiceStub{}), authz)
	return router, authz
}

type historyAuthzStub struct {
	granted        map[string]bool
	lastPermission string
}

func (a *historyAuthzStub) GetGlobalRole(context.Context, uuid.UUID) (domainUser.Role, error) {
	return domainUser.RolePlaner, nil
}

func (a *historyAuthzStub) GetTeamRole(context.Context, uuid.UUID, uuid.UUID) (*domainTeam.MemberRole, error) {
	return nil, nil
}

func (a *historyAuthzStub) HasPermission(_ context.Context, _ domainUser.Role, permission string) (bool, error) {
	a.lastPermission = permission
	return a.granted[permission], nil
}

type historyServiceStub struct{}

func (historyServiceStub) ListTimeline(context.Context, domainHistory.TimelineFilter) (*domain.PaginatedList[domainHistory.ChangeEvent], error) {
	return &domain.PaginatedList[domainHistory.ChangeEvent]{Items: []domainHistory.ChangeEvent{}}, nil
}

func (historyServiceStub) GetEvent(context.Context, uuid.UUID) (*domainHistory.ChangeEvent, error) {
	return &domainHistory.ChangeEvent{}, nil
}

func (historyServiceStub) RestoreEntityToEvent(context.Context, uuid.UUID, domainHistory.RestoreMode) (*domainHistory.RestoreResult, error) {
	return &domainHistory.RestoreResult{}, nil
}

func (historyServiceStub) RestoreControlCabinet(context.Context, uuid.UUID, domainHistory.RestoreControlCabinetRequest) (*domainHistory.RestoreResult, error) {
	return &domainHistory.RestoreResult{}, nil
}
