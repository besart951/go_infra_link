package user

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type User struct {
	domain.Base
	FirstName       string
	LastName        string
	Email           string `gorm:"uniqueIndex" json:"email"`
	Password        string `json:"-"`
	IsActive        bool   `gorm:"default:true"`
	CreatedByID     *uuid.UUID
	CreatedBy       *User            `gorm:"foreignKey:CreatedByID"`
	BusinessDetails *BusinessDetails `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"business_details,omitempty"`
}

type BusinessDetails struct {
	domain.Base
	UserID      uuid.UUID `gorm:"uniqueIndex"`
	CompanyName string
	VatNumber   string
}

// Repository Interface definition
type UserRepository interface {
	GetByIds(ids []uuid.UUID) ([]*User, error)
	Create(entity *User) error
	Update(entity *User) error
	DeleteByIds(ids []uuid.UUID) error
	GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[User], error)
}
