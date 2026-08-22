package facilitysql

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/repository/postgreserror"
)

const fieldDeviceNumberConstraint = "uq_field_devices_number_scope"

func mapFieldDeviceWriteError(err error) error {
	if postgreserror.IsUniqueConstraint(err, fieldDeviceNumberConstraint) {
		return domain.ErrConflict
	}
	return err
}
