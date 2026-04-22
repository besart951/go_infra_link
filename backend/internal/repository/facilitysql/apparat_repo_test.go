package facilitysql

import (
	"errors"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

func TestMapApparatWriteError_MapsPostgresDuplicateToDomainConflict(t *testing.T) {
	err := mapApparatWriteError(&pgconn.PgError{Code: "23505"})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected domain conflict, got %v", err)
	}
}

func TestMapApparatWriteError_MapsTranslatedDuplicateToDomainConflict(t *testing.T) {
	err := mapApparatWriteError(gorm.ErrDuplicatedKey)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected domain conflict, got %v", err)
	}
}
