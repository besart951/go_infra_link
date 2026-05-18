package user

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	domainTeam "github.com/besart951/go_infra_link/backend/internal/domain/team"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	userdirectory "github.com/besart951/go_infra_link/backend/internal/service/userdirectory"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestRegisterUserRoutes_ListUsersIncludeDeletedRequiresReadDeletedPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authz := &routeAuthCheckerStub{
		role: domainUser.RoleAdminFZAG,
		permissions: map[string]bool{
			domainUser.PermissionUserRead:        true,
			domainUser.PermissionUserReadDeleted: false,
		},
	}
	handlers := &Handlers{
		User:         NewUserHandler(&fakeUserService{}, nil, &routeDirectoryServiceStub{}, nil),
		Registration: &RegistrationHandler{},
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, uuid.New())
		c.Next()
	})
	protectedV1 := router.Group("/api/v1")
	RegisterUserRoutes(protectedV1, handlers, authz)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users?include_deleted=true", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.Code)
	}
}

func TestRegisterUserRoutes_ListUsersWithoutIncludeDeletedDoesNotRequireReadDeletedPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authz := &routeAuthCheckerStub{
		role: domainUser.RoleAdminFZAG,
		permissions: map[string]bool{
			domainUser.PermissionUserRead:        true,
			domainUser.PermissionUserReadDeleted: false,
		},
	}
	handlers := &Handlers{
		User:         NewUserHandler(&fakeUserService{}, nil, &routeDirectoryServiceStub{}, nil),
		Registration: &RegistrationHandler{},
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, uuid.New())
		c.Next()
	})
	protectedV1 := router.Group("/api/v1")
	RegisterUserRoutes(protectedV1, handlers, authz)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
}

func TestRegisterUserRoutes_DirectoryIncludeDeletedRequiresReadDeletedPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authz := &routeAuthCheckerStub{
		role: domainUser.RoleAdminFZAG,
		permissions: map[string]bool{
			domainUser.PermissionUserReadDeleted: false,
		},
	}
	handlers := &Handlers{
		User:         NewUserHandler(&fakeUserService{}, nil, &routeDirectoryServiceStub{}, nil),
		Registration: &RegistrationHandler{},
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, uuid.New())
		c.Next()
	})
	protectedV1 := router.Group("/api/v1")
	RegisterUserRoutes(protectedV1, handlers, authz)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/directory?include_deleted=true", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.Code)
	}
}

func TestRegisterAdminRoutes_RestoreRequiresDeletePermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authz := &routeAuthCheckerStub{
		role: domainUser.RoleAdminFZAG,
		permissions: map[string]bool{
			domainUser.PermissionUserDelete:      false,
			domainUser.PermissionUserReadDeleted: true,
		},
	}
	handlers := &Handlers{
		Admin: NewAdminHandler(&routeAdminServiceStub{}, &fakeUserService{}),
	}

	router := newRouteAuthRouter()
	protectedV1 := router.Group("/api/v1")
	RegisterAdminRoutes(protectedV1, handlers, authz)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/"+uuid.NewString()+"/restore", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.Code)
	}
}

func TestRegisterAdminRoutes_RestoreRequiresReadDeletedPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authz := &routeAuthCheckerStub{
		role: domainUser.RoleAdminFZAG,
		permissions: map[string]bool{
			domainUser.PermissionUserDelete:      true,
			domainUser.PermissionUserReadDeleted: false,
		},
	}
	handlers := &Handlers{
		Admin: NewAdminHandler(&routeAdminServiceStub{}, &fakeUserService{}),
	}

	router := newRouteAuthRouter()
	protectedV1 := router.Group("/api/v1")
	RegisterAdminRoutes(protectedV1, handlers, authz)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/"+uuid.NewString()+"/restore", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.Code)
	}
}

func newRouteAuthRouter() *gin.Engine {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, uuid.New())
		c.Next()
	})
	return router
}

type routeAuthCheckerStub struct {
	role        domainUser.Role
	permissions map[string]bool
}

func (s *routeAuthCheckerStub) GetGlobalRole(context.Context, uuid.UUID) (domainUser.Role, error) {
	return s.role, nil
}

func (s *routeAuthCheckerStub) GetTeamRole(context.Context, uuid.UUID, uuid.UUID) (*domainTeam.MemberRole, error) {
	return nil, nil
}

func (s *routeAuthCheckerStub) HasPermission(_ context.Context, _ domainUser.Role, permission string) (bool, error) {
	if s.permissions == nil {
		return false, nil
	}
	return s.permissions[permission], nil
}

type routeDirectoryServiceStub struct{}

func (s *routeDirectoryServiceStub) List(context.Context, uuid.UUID, int, int, string, string, string, string, string, bool) (*userdirectory.ListResult, error) {
	return &userdirectory.ListResult{Items: []userdirectory.Item{}, Total: 0, Page: 1, TotalPages: 1, Teams: []userdirectory.TeamFilter{}, Roles: []userdirectory.RoleFilter{}}, nil
}

type routeAdminServiceStub struct{}

func (s *routeAdminServiceStub) DisableUser(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

func (s *routeAdminServiceStub) EnableUser(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

func (s *routeAdminServiceStub) SetUserRole(context.Context, uuid.UUID, uuid.UUID, domainUser.Role) error {
	return nil
}
