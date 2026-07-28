package facilitysql

import (
	"errors"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestMapFieldDeviceWriteErrorMapsPlacementConstraint(t *testing.T) {
	mapped := MapFieldDeviceWriteError(&pgconn.PgError{
		Code:           "23505",
		ConstraintName: fieldDevicePlacementConstraintName,
	})
	validation, ok := domain.AsValidationError(mapped)
	if !ok ||
		validation.Fields["fielddevice.apparat_nr"] !=
			"apparatnummer ist bereits vergeben" {
		t.Fatalf("mapped placement error: %#v", mapped)
	}
}

func TestMapFieldDeviceWriteErrorPreservesOtherErrors(t *testing.T) {
	original := errors.New("write failed")
	if mapped := MapFieldDeviceWriteError(original); !errors.Is(mapped, original) {
		t.Fatalf("other error changed: %v", mapped)
	}
}
