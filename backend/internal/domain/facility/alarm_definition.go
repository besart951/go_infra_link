package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type AlarmDefinition struct {
	domain.Base
	Name        string `gorm:"not null"`
	AlarmNote   *string
	AlarmTypeID *uuid.UUID `gorm:"type:uuid;index"`
	AlarmType   *AlarmType `gorm:"foreignKey:AlarmTypeID"`
	Scope       string     `gorm:"not null;default:'template';size:30"`
}
