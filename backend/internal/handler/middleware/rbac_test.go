package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	domainTeam "github.com/besart951/go_infra_link/backend/internal/domain/team"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestRequireTeamPermission_AllowsMemberBundlePermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	teamID := uuid.New()
	authz := &authCheckerStub{
		globalRole: domainUser.RolePlaner,
		teamRole:   rolePtr(domainTeam.MemberRoleManager),
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Next()
	})
	router.GET("/teams/:id", RequireTeamPermission(authz, "id", domainTeam.PermissionTeamMemberAdd), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/teams/"+teamID.String(), nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNoContent {
		t.Fatalf("expected request to succeed, got status %d", resp.Code)
	}
}

func TestRequireTeamPermission_RejectsMissingMemberPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	teamID := uuid.New()
	authz := &authCheckerStub{
		globalRole: domainUser.RolePlaner,
		teamRole:   rolePtr(domainTeam.MemberRoleMember),
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Next()
	})
	router.GET("/teams/:id", RequireTeamPermission(authz, "id", domainTeam.PermissionTeamDelete), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/teams/"+teamID.String(), nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden, got status %d", resp.Code)
	}
}

func TestRequireTeamPermission_GlobalPermissionBypassesMembershipRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	teamID := uuid.New()
	authz := &authCheckerStub{
		globalRole:            domainUser.RoleSuperAdmin,
		hasPermissionResponse: true,
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Next()
	})
	router.GET("/teams/:id", RequireTeamPermission(authz, "id", domainTeam.PermissionTeamDelete), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/teams/"+teamID.String(), nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNoContent {
		t.Fatalf("expected global permission bypass to succeed, got status %d", resp.Code)
	}
	if authz.lastPermission != domainTeam.PermissionTeamDelete {
		t.Fatalf("expected global permission check for %q, got %q", domainTeam.PermissionTeamDelete, authz.lastPermission)
	}
}

type authCheckerStub struct {
	globalRole            domainUser.Role
	teamRole              *domainTeam.MemberRole
	hasPermissionResponse bool
	lastPermission        string
}

func (a *authCheckerStub) GetGlobalRole(context.Context, uuid.UUID) (domainUser.Role, error) {
	return a.globalRole, nil
}

func (a *authCheckerStub) GetTeamRole(context.Context, uuid.UUID, uuid.UUID) (*domainTeam.MemberRole, error) {
	return a.teamRole, nil
}

func (a *authCheckerStub) HasPermission(_ context.Context, _ domainUser.Role, permission string) (bool, error) {
	a.lastPermission = permission
	return a.hasPermissionResponse, nil
}

func rolePtr(role domainTeam.MemberRole) *domainTeam.MemberRole {
	return &role
}