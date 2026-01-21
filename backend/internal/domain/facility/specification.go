package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type Specification struct {
	domain.Base
	FieldDeviceID                             *uuid.UUID
	SpecificationSupplier                     *string
	SpecificationBrand                        *string
	SpecificationType                         *string
	AdditionalInfoMotorValve                  *string
	AdditionalInfoSize                        *int
	AdditionalInformationInstallationLocation *string
	ElectricalConnectionPH                    *int
	ElectricalConnectionACDC                  *string
	ElectricalConnectionAmperage              *float64
	ElectricalConnectionPower                 *float64
	ElectricalConnectionRotation              *int
}
