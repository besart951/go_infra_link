package wire

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	exportservice "github.com/besart951/go_infra_link/backend/internal/service/exporting"
	"github.com/google/uuid"
)

type testUserRepo struct{}

func (*testUserRepo) GetByIds(ctx context.Context, ids []uuid.UUID) ([]*domainUser.User, error) {
	return nil, nil
}
func (*testUserRepo) Create(ctx context.Context, entity *domainUser.User) error { return nil }
func (*testUserRepo) Update(ctx context.Context, entity *domainUser.User) error { return nil }
func (*testUserRepo) DeleteByIds(ctx context.Context, ids []uuid.UUID) error    { return nil }
func (*testUserRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainUser.User], error) {
	return nil, nil
}

type testUserEmailRepo struct {
	testUserRepo
}

func (*testUserEmailRepo) GetByEmail(ctx context.Context, email string) (*domainUser.User, error) {
	return nil, nil
}

type testUserLifecycleRepo struct {
	testUserEmailRepo
}

func (*testUserLifecycleRepo) ListDueForAnonymization(ctx context.Context, now time.Time, limit int) ([]*domainUser.User, error) {
	return nil, nil
}

func TestRequireUserEmailRepository_Missing(t *testing.T) {
	_, err := requireUserEmailRepository(&testUserRepo{})
	if !errors.Is(err, ErrUserRepoMissingEmailLookup) {
		t.Fatalf("expected error %v, got %v", ErrUserRepoMissingEmailLookup, err)
	}
}

func TestRequireUserLifecycleStore_Missing(t *testing.T) {
	_, err := requireUserLifecycleStore(&testUserEmailRepo{})
	if !errors.Is(err, ErrUserRepoMissingLifecycle) {
		t.Fatalf("expected error %v, got %v", ErrUserRepoMissingLifecycle, err)
	}
}

func TestRuntimeBusRejectsUnsupportedBus(t *testing.T) {
	_, err := newRuntimeBus(context.Background(), RuntimeConfig{Bus: "redis"})
	if err == nil {
		t.Fatal("expected unsupported bus error")
	}

	if !strings.Contains(err.Error(), `unsupported realtime bus "redis": expected "memory" or "postgres"`) {
		t.Fatalf("unexpected runtime bus error: %v", err)
	}
}

func TestExportDefaultsAreApplied(t *testing.T) {
	resolved := resolveExportConfig(exportservice.Config{})
	if got, want := resolved.QueueSize, 200; got != want {
		t.Fatalf("QueueSize: expected %d, got %d", want, got)
	}
	if got, want := resolved.MaxConcurrent, 1; got != want {
		t.Fatalf("MaxConcurrent: expected %d, got %d", want, got)
	}
	if got, want := resolved.SingleFileDeviceLimit, int64(5000); got != want {
		t.Fatalf("SingleFileDeviceLimit: expected %d, got %d", want, got)
	}
	if got, want := resolved.PageSize, 1000; got != want {
		t.Fatalf("PageSize: expected %d, got %d", want, got)
	}

	if got := resolveExportDirectory(ServiceConfig{}); got != defaultExportDirectory() {
		t.Fatalf("expected default export directory %q, got %q", defaultExportDirectory(), got)
	}
}

func TestRepositoriesFromUnitWrapsTransactionError(t *testing.T) {
	_, err := repositoriesFromUnit(struct{}{})
	if err == nil {
		t.Fatal("expected repositoriesFromUnit to fail")
	}

	if !strings.Contains(err.Error(), "resolve transaction unit:") {
		t.Fatalf("expected wrapped transaction error, got: %v", err)
	}
}
