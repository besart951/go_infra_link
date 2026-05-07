package projectsql

import (
	"errors"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

func mapWriteError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return domain.ErrConflict
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return domain.ErrConflict
	}
	if strings.Contains(err.Error(), "UNIQUE constraint failed") {
		return domain.ErrConflict
	}
	return err
}
