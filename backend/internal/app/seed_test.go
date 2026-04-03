package app

import (
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/config"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type seedLogger struct{}

func (seedLogger) Info(string, ...any)  {}
func (seedLogger) Error(string, ...any) {}

type seedUserServiceStub struct {
	createCalls int
}

func (s *seedUserServiceStub) CreateWithPassword(user *domainUser.User, password string) error {
	s.createCalls++
	return nil
}

type userEmailRepoStub struct {
	user *domainUser.User
	err  error
}

func (r userEmailRepoStub) GetByEmail(email string) (*domainUser.User, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.user, nil
}

func TestEnsureSeedUserCreatesMissingUser(t *testing.T) {
	service := &seedUserServiceStub{}
	cfg := config.Config{
		SeedUserEnabled:   true,
		SeedUserFirstName: "Test",
		SeedUserLastName:  "Admin",
		SeedUserEmail:     "test@example.com",
		SeedUserPassword:  "secret123",
	}

	err := ensureSeedUser(cfg, seedLogger{}, service, userEmailRepoStub{err: domain.ErrNotFound})
	if err != nil {
		t.Fatalf("ensureSeedUser returned error: %v", err)
	}
	if service.createCalls != 1 {
		t.Fatalf("expected CreateWithPassword to be called once, got %d", service.createCalls)
	}
}

func TestEnsureSeedUserSkipsExistingUserMutation(t *testing.T) {
	service := &seedUserServiceStub{}
	existing := &domainUser.User{}
	existing.ID = uuid.New()
	existing.Email = "test@example.com"
	existing.Role = domainUser.RoleFZAG
	existing.IsActive = false

	cfg := config.Config{
		SeedUserEnabled:   true,
		SeedUserFirstName: "Updated",
		SeedUserLastName:  "Admin",
		SeedUserEmail:     existing.Email,
		SeedUserPassword:  "another-secret",
	}

	err := ensureSeedUser(cfg, seedLogger{}, service, userEmailRepoStub{user: existing})
	if err != nil {
		t.Fatalf("ensureSeedUser returned error: %v", err)
	}
	if service.createCalls != 0 {
		t.Fatalf("expected CreateWithPassword not to be called, got %d", service.createCalls)
	}
	if existing.Role != domainUser.RoleFZAG {
		t.Fatalf("expected existing role to remain unchanged, got %s", existing.Role)
	}
	if existing.IsActive {
		t.Fatalf("expected existing user active flag to remain unchanged")
	}
}
