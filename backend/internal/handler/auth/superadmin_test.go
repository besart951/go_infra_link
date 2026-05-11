package auth

import (
	"context"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainAuth "github.com/besart951/go_infra_link/backend/internal/domain/auth"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

func TestAuthHandler_BuildAuthResponse_SuperadminCanAccessUserDirectoryWithoutStoredPermission(t *testing.T) {
	handler := NewAuthHandler(nil, nil, permissionQueryStub{permissions: []string{}}, nil, 0, 0, CookieSettings{})
	userID := uuid.New()

	response := handler.buildAuthResponse(context.Background(), &domainAuth.LoginResult{
		User: &domainUser.User{
			Base:      domain.Base{ID: userID},
			FirstName: "Super",
			LastName:  "Admin",
			Email:     domainUser.EmailPtr("superadmin@example.com"),
			IsActive:  true,
			Role:      domainUser.RoleSuperAdmin,
		},
	})

	if !response.User.CanAccessUserDirectory {
		t.Fatal("expected superadmin to access the user directory without a stored user.read permission")
	}
}

type permissionQueryStub struct {
	permissions []string
}

func (s permissionQueryStub) GetRolePermissions(context.Context, domainUser.Role) ([]string, error) {
	return append([]string{}, s.permissions...), nil
}
