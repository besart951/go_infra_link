package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
)

type Apparat struct {
	domain.Base
	ShortName   string
	Name        string
	Description *string

	SystemParts  []*SystemPart
	FieldDevices []FieldDevice
}
