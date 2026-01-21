package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
)

type NotificationClass struct {
	domain.Base
	EventCategory        string
	Nc                   int
	ObjectDescription    string
	InternalDescription  string
	Meaning              string
	AckRequiredNotNormal bool
	AckRequiredError     bool
	AckRequiredNormal    bool
	NormNotNormal        int
	NormError            int
	NormNormal           int
}
