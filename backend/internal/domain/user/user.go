package user

import (
	"context"
	"strings"
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
)

type User struct {
	domain.Base
	FirstName           string     `gorm:"not null"`
	LastName            string     `gorm:"not null"`
	Email               *string    `json:"email,omitempty" gorm:"type:varchar(320);uniqueIndex"`
	Password            string     `json:"-" gorm:"not null"`
	IsActive            bool       `gorm:"default:true"`
	Role                Role       `gorm:"type:varchar(50);default:'entrepreneur'"`
	DisabledAt          *time.Time `gorm:"index"`
	DeletedAt           *time.Time `gorm:"index"`
	DeletedByID         *uuid.UUID `gorm:"type:uuid;index"`
	RestoreUntil        *time.Time `gorm:"index"`
	ScheduledPurgeAt    *time.Time `gorm:"index"`
	AnonymizedAt        *time.Time `gorm:"index"`
	DeletedEmailHash    *string    `gorm:"type:char(64);index"`
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

const DeletedUserDisplayName = "Deleted User"

func NormalizeEmail(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func EmailPtr(value string) *string {
	normalized := NormalizeEmail(value)
	if normalized == "" {
		return nil
	}
	return &normalized
}

func EmailString(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func (u User) EmailValue() string {
	return EmailString(u.Email)
}

func (u *User) SetEmail(value string) {
	u.Email = EmailPtr(value)
}

func (u User) IsDeleted() bool {
	return u.DeletedAt != nil
}

func (u User) IsAnonymized() bool {
	return u.AnonymizedAt != nil
}

func (u User) IsIdentityHidden() bool {
	return u.IsDeleted() || u.IsAnonymized()
}

func (u User) DisplayName() string {
	if u.IsIdentityHidden() {
		return DeletedUserDisplayName
	}
	name := strings.TrimSpace(strings.Join([]string{strings.TrimSpace(u.FirstName), strings.TrimSpace(u.LastName)}, " "))
	if name != "" {
		return name
	}
	return u.EmailValue()
}

type UserRepository interface {
	domain.Repository[User]
}

type UserLifecycleRepository interface {
	UserRepository
	UserEmailRepository
	ListDueForAnonymization(ctx context.Context, now time.Time, limit int) ([]*User, error)
}
