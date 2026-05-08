package userregistration

import (
	"errors"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

func TestMapWriteErrorMapsDuplicateKeyToConflict(t *testing.T) {
	cases := []error{
		gorm.ErrDuplicatedKey,
		&pgconn.PgError{Code: "23505"},
		errors.New("UNIQUE constraint failed: users.email"),
	}

	for _, err := range cases {
		if !errors.Is(mapWriteError(err), domain.ErrConflict) {
			t.Fatalf("expected duplicate error %v to map to conflict", err)
		}
	}
}
