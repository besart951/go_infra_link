package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
)

type SystemType struct {
	domain.Base
	NumberMin int
	NumberMax int
	Name      string
}
