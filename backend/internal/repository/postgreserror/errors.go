package postgreserror

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

const uniqueViolation = "23505"

func IsUniqueConstraint(err error, constraint string) bool {
	var databaseError *pgconn.PgError
	return errors.As(err, &databaseError) &&
		databaseError.Code == uniqueViolation &&
		databaseError.ConstraintName == constraint
}
