package user

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
)

// Permission represents a specific action that can be performed
type Permission struct {
	domain.Base
	Name        string `gorm:"uniqueIndex;not null"`
	Description string
	Resource    string `gorm:"not null"` // e.g., "user", "team", "project"
	Action      string `gorm:"not null"` // e.g., "create", "read", "update", "delete"
}

// RolePermission links roles to permissions
type RolePermission struct {
	domain.Base
	Role       Role   `gorm:"type:varchar(50);not null;index:idx_role_permission,unique"`
	Permission string `gorm:"not null;index:idx_role_permission,unique"`
}
