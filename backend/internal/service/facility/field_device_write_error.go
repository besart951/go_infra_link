package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/repository/postgreserror"
)

func mapFieldDeviceNumberConflict(err error) error {
	if postgreserror.IsUniqueConstraint(err, "uq_field_devices_number_scope") {
		return domain.ErrConflict
	}
	return err
}
