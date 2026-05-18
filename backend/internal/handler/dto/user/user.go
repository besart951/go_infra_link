package user

import (
	"time"

	"github.com/google/uuid"
)

// User DTOs

type CreateUserRequest struct {
	FirstName   string     `json:"first_name" binding:"required,min=1,max=100"`
	LastName    string     `json:"last_name" binding:"required,min=1,max=100"`
	Email       string     `json:"email" binding:"required,email"`
	Password    string     `json:"password" binding:"required,min=8"`
	IsActive    bool       `json:"is_active"`
	Role        string     `json:"role" binding:"omitempty,oneof=superadmin admin_fzag fzag admin_planer planer admin_entrepreneur entrepreneur"`
	CreatedByID *uuid.UUID `json:"created_by_id"`
}

type UpdateUserRequest struct {
	FirstName string  `json:"first_name" binding:"omitempty,min=1,max=100"`
	LastName  string  `json:"last_name" binding:"omitempty,min=1,max=100"`
	Email     string  `json:"email" binding:"omitempty,email"`
	Password  *string `json:"password" binding:"omitempty,min=8"`
	IsActive  *bool   `json:"is_active"`
	Role      *string `json:"role" binding:"omitempty,oneof=superadmin admin_fzag fzag admin_planer planer admin_entrepreneur entrepreneur"`
}

type UpdateOwnPasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

type UserResponse struct {
	ID                  uuid.UUID  `json:"id"`
	FirstName           string     `json:"first_name"`
	LastName            string     `json:"last_name"`
	Email               string     `json:"email"`
	IsActive            bool       `json:"is_active"`
	Role                string     `json:"role"`
	RoleDisplayName     string     `json:"role_display_name"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	LastLoginAt         *time.Time `json:"last_login_at,omitempty"`
	DisabledAt          *time.Time `json:"disabled_at,omitempty"`
	LockedUntil         *time.Time `json:"locked_until,omitempty"`
	FailedLoginAttempts int        `json:"failed_login_attempts"`
	IsDeleted           bool       `json:"is_deleted"`
	IsAnonymized        bool       `json:"is_anonymized"`
	DeletedAt           *time.Time `json:"deleted_at,omitempty"`
	RestoreUntil        *time.Time `json:"restore_until,omitempty"`
}

type UserListResponse struct {
	Items      []UserResponse `json:"items"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	TotalPages int            `json:"total_pages"`
}

type UserDirectoryTeamResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type UserDirectoryCapabilitiesResponse struct {
	CanUpdate     bool `json:"can_update"`
	CanDelete     bool `json:"can_delete"`
	CanDisable    bool `json:"can_disable"`
	CanEnable     bool `json:"can_enable"`
	CanRestore    bool `json:"can_restore"`
	CanChangeRole bool `json:"can_change_role"`
}

type RegistrationProcessStepResponse struct {
	Key       string     `json:"key"`
	Label     string     `json:"label"`
	Status    string     `json:"status"`
	Timestamp *time.Time `json:"timestamp,omitempty"`
}

type RegistrationProcessResponse struct {
	Status            string                            `json:"status"`
	EmailStatus       string                            `json:"email_status"`
	Steps             []RegistrationProcessStepResponse `json:"steps"`
	CanResend         bool                              `json:"can_resend"`
	ExpiresAt         *time.Time                        `json:"expires_at,omitempty"`
	AcceptedAt        *time.Time                        `json:"accepted_at,omitempty"`
	LastSentAt        *time.Time                        `json:"last_sent_at,omitempty"`
	ResendAvailableAt *time.Time                        `json:"resend_available_at,omitempty"`
	SendCount         int                               `json:"send_count"`
	LastError         string                            `json:"last_error,omitempty"`
}

type UserDirectoryPageCapabilitiesResponse struct {
	CanCreateUser  bool `json:"can_create_user"`
	CanReadDeleted bool `json:"can_read_deleted"`
}

type UserDirectoryUserResponse struct {
	ID                  uuid.UUID                         `json:"id"`
	FirstName           string                            `json:"first_name"`
	LastName            string                            `json:"last_name"`
	Email               string                            `json:"email"`
	IsActive            bool                              `json:"is_active"`
	Role                string                            `json:"role"`
	RoleDisplayName     string                            `json:"role_display_name"`
	CreatedAt           time.Time                         `json:"created_at"`
	UpdatedAt           time.Time                         `json:"updated_at"`
	LastLoginAt         *time.Time                        `json:"last_login_at,omitempty"`
	DisabledAt          *time.Time                        `json:"disabled_at,omitempty"`
	LockedUntil         *time.Time                        `json:"locked_until,omitempty"`
	FailedLoginAttempts int                               `json:"failed_login_attempts"`
	IsDeleted           bool                              `json:"is_deleted"`
	IsAnonymized        bool                              `json:"is_anonymized"`
	DeletedAt           *time.Time                        `json:"deleted_at,omitempty"`
	RestoreUntil        *time.Time                        `json:"restore_until,omitempty"`
	Teams               []UserDirectoryTeamResponse       `json:"teams"`
	Capabilities        UserDirectoryCapabilitiesResponse `json:"capabilities"`
	RegistrationProcess *RegistrationProcessResponse      `json:"registration_process,omitempty"`
}

type UserDirectoryTeamFilterResponse struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Count int64     `json:"count"`
}

type UserDirectoryRoleFilterResponse struct {
	Role        string `json:"role"`
	DisplayName string `json:"display_name"`
	Count       int64  `json:"count"`
}

type UserDirectoryListResponse struct {
	Items        []UserDirectoryUserResponse           `json:"items"`
	Total        int64                                 `json:"total"`
	Page         int                                   `json:"page"`
	TotalPages   int                                   `json:"total_pages"`
	Teams        []UserDirectoryTeamFilterResponse     `json:"teams"`
	Roles        []UserDirectoryRoleFilterResponse     `json:"roles"`
	Capabilities UserDirectoryPageCapabilitiesResponse `json:"capabilities"`
}

type AllowedRole struct {
	Role        string `json:"role"`
	DisplayName string `json:"display_name"`
}

type AllowedRolesResponse struct {
	Roles []AllowedRole `json:"roles"`
}

type AddUserToTeamRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
	TeamID uuid.UUID `json:"team_id" binding:"required"`
}

type CreateUserInvitationRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required,oneof=superadmin admin_fzag fzag admin_planer planer admin_entrepreneur entrepreneur"`
}

type CreateUserInvitationResponse struct {
	User                UserResponse                `json:"user"`
	RegistrationProcess RegistrationProcessResponse `json:"registration_process"`
}
