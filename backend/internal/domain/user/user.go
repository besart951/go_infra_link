package user

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
	"time"
)

type Role string

const (
	RoleUser       Role = "user"
	RoleAdmin      Role = "admin"
	RoleSuperAdmin Role = "superadmin"
)

type User struct {
	domain.Base
	FirstName       string
	LastName        string
	Email           string `json:"email"`
	Password        string `json:"-"`
	IsActive        bool
	Role            Role
	DisabledAt      *time.Time
	LockedUntil     *time.Time
	FailedLoginAttempts int
	LastLoginAt     *time.Time
	CreatedByID     *uuid.UUID
	CreatedBy       *User
	BusinessDetails *BusinessDetails `json:"business_details,omitempty"`
}

type BusinessDetails struct {
	domain.Base
	UserID      uuid.UUID
	CompanyName string
	VatNumber   string
}

type UserRepository = domain.Repository[User]
