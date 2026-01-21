package user

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type User struct {
	domain.Base
	FirstName       string
	LastName        string
	Email           string `json:"email"`
	Password        string `json:"-"`
	IsActive        bool
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
