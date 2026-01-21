package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
)

type AlarmDefinition struct {
	domain.Base
	Name      string
	AlarmNote *string
}
