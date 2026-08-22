package postgreserror

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestIsUniqueConstraintMatchesCodeAndName(t *testing.T) {
	err := &pgconn.PgError{Code: "23505", ConstraintName: "expected"}
	if !IsUniqueConstraint(err, "expected") {
		t.Fatal("expected matching unique constraint")
	}
	if IsUniqueConstraint(err, "other") || IsUniqueConstraint(errors.New("other"), "expected") {
		t.Fatal("unexpected unique constraint match")
	}
}
