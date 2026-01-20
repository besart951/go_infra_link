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

type UserRepository = domain.Repository[User]
