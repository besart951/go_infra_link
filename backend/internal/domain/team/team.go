package team

import "github.com/besart951/go_infra_link/backend/internal/domain"

type Team struct {
	domain.Base
	Name        string  `gorm:"not null"`
	Description *string
}
