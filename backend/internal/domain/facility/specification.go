package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
)

type Specification struct {
	domain.Base
	SpecificationSupplier                     *string `gorm:"size:250"`
	SpecificationBrand                        *string `gorm:"size:250"`
	SpecificationType                         *string `gorm:"size:250"`
	AdditionalInfoMotorValve                  *string `gorm:"size:250"`
	AdditionalInfoSize                        *int
	AdditionalInformationInstallationLocation *string `gorm:"size:250"`
	ElectricalConnectionPH                    *int
	ElectricalConnectionACDC                  *string `gorm:"size:2"`
	ElectricalConnectionAmperage              *float64
	ElectricalConnectionPower                 *float64
	ElectricalConnectionRotation              *int
}
