package user

import (
	"time"

	domainuser "github.com/besart951/go_infra_link/backend/internal/domain/user"
)

type RoleResponse struct {
	ID          string            `json:"id"`
	Name        domainuser.Role   `json:"name"`
	DisplayName string            `json:"display_name"`
	Description string            `json:"description"`
	Level       int               `json:"level"`
	Permissions []string          `json:"permissions"`
	CanManage   []domainuser.Role `json:"can_manage"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type UpdateRolePermissionsRequest struct {
	Permissions []string `json:"permissions"`
}

type AddRolePermissionRequest struct {
	Permission string `json:"permission" binding:"required"`
}

type RolePermissionResponse struct {
	ID         string          `json:"id"`
	Role       domainuser.Role `json:"role"`
	Permission string          `json:"permission"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}
