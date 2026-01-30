package user

import (
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type Role string

const (
	RoleSuperAdmin        Role = "superadmin"
	RoleAdminFZAG         Role = "admin_fzag"
	RoleFZAG              Role = "fzag"
	RoleAdminPlaner       Role = "admin_planer"
	RolePlaner            Role = "planer"
	RoleAdminEnterpreneur Role = "admin_entrepreneur"
	RoleEnterpreneur      Role = "entrepreneur"
	// Legacy roles (kept for backwards compatibility)
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type User struct {
	domain.Base
	FirstName           string     `gorm:"not null"`
	LastName            string     `gorm:"not null"`
	Email               string     `json:"email" gorm:"uniqueIndex;not null"`
	Password            string     `json:"-" gorm:"not null"`
	IsActive            bool       `gorm:"default:true"`
	Role                Role       `gorm:"type:varchar(50);default:'user'"`
	DisabledAt          *time.Time `gorm:"index"`
	LockedUntil         *time.Time `gorm:"index"`
	FailedLoginAttempts int        `gorm:"default:0"`
	LastLoginAt         *time.Time
	CreatedByID         *uuid.UUID       `gorm:"type:uuid"`
	CreatedBy           *User            `gorm:"foreignKey:CreatedByID"`
	BusinessDetails     *BusinessDetails `json:"business_details,omitempty" gorm:"foreignKey:UserID"`
	Teams               []UserTeam       `gorm:"foreignKey:UserID"`
}

type BusinessDetails struct {
	domain.Base
	UserID      uuid.UUID `gorm:"type:uuid;uniqueIndex;not null"`
	CompanyName string
	VatNumber   string
}

type UserRepository = domain.Repository[User]
