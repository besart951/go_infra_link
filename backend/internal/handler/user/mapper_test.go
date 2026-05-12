package user

import (
	"testing"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	userdirectory "github.com/besart951/go_infra_link/backend/internal/service/userdirectory"
	"github.com/google/uuid"
)

func TestToUserDirectoryListResponseHidesDeletedEmailWithoutReadDeletedPermission(t *testing.T) {
	now := time.Now().UTC()
	result := &userdirectory.ListResult{
		Items: []userdirectory.Item{
			{
				User: domainUser.User{
					Base:      domain.Base{ID: uuid.New()},
					FirstName: "Delete",
					LastName:  "Me",
					Email:     domainUser.EmailPtr("deleted.user@example.com"),
					DeletedAt: &now,
				},
			},
		},
		PageCapabilities: userdirectory.PageCapabilities{CanReadDeleted: false},
	}

	response := ToUserDirectoryListResponse(result, nil)
	if response.Items[0].Email != "" {
		t.Fatalf("expected deleted user email to be hidden without permission, got %q", response.Items[0].Email)
	}
}

func TestToUserDirectoryListResponseRevealsDeletedEmailWithReadDeletedPermission(t *testing.T) {
	now := time.Now().UTC()
	email := "deleted.user@example.com"
	result := &userdirectory.ListResult{
		Items: []userdirectory.Item{
			{
				User: domainUser.User{
					Base:      domain.Base{ID: uuid.New()},
					FirstName: "Delete",
					LastName:  "Me",
					Email:     domainUser.EmailPtr(email),
					DeletedAt: &now,
				},
			},
		},
		PageCapabilities: userdirectory.PageCapabilities{CanReadDeleted: true},
	}

	response := ToUserDirectoryListResponse(result, nil)
	if response.Items[0].Email != email {
		t.Fatalf("expected deleted user email to be revealed with permission, got %q", response.Items[0].Email)
	}
}

func TestToUserDirectoryListResponseKeepsAnonymizedEmailHidden(t *testing.T) {
	now := time.Now().UTC()
	anonymizedAt := now
	result := &userdirectory.ListResult{
		Items: []userdirectory.Item{
			{
				User: domainUser.User{
					Base:         domain.Base{ID: uuid.New()},
					FirstName:    "Delete",
					LastName:     "Me",
					Email:        domainUser.EmailPtr("anonymized.user@example.com"),
					DeletedAt:    &now,
					AnonymizedAt: &anonymizedAt,
				},
			},
		},
		PageCapabilities: userdirectory.PageCapabilities{CanReadDeleted: true},
	}

	response := ToUserDirectoryListResponse(result, nil)
	if response.Items[0].Email != "" {
		t.Fatalf("expected anonymized deleted user email to stay hidden, got %q", response.Items[0].Email)
	}
}
