package facilitysql

import (
	"errors"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

const fieldDevicePlacementConstraintName = "uq_field_devices_placement_apparat_nr"

// MapFieldDeviceWriteError translates database-enforced placement conflicts to
// the same field-level domain error used by preflight validation. It is
// exported so the composition root can also translate a deferred constraint
// failure returned at transaction commit.
func MapFieldDeviceWriteError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) &&
		pgErr.Code == "23505" &&
		pgErr.ConstraintName == fieldDevicePlacementConstraintName {
		return domain.NewValidationError().
			Add("fielddevice.apparat_nr", "apparatnummer ist bereits vergeben")
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) &&
		strings.Contains(err.Error(), fieldDevicePlacementConstraintName) {
		return domain.NewValidationError().
			Add("fielddevice.apparat_nr", "apparatnummer ist bereits vergeben")
	}
	return err
}
