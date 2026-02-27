package mapper

import (
	"github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
)

// ToUserModel converts a CreateUserRequest to a User domain model
func ToUserModel(req dto.CreateUserRequest) *user.User {
	usr := &user.User{
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Email:       req.Email,
		IsActive:    req.IsActive,
		CreatedByID: req.CreatedByID,
	}
	if req.Role != "" {
		usr.Role = user.Role(req.Role)
	}
	return usr
}

// ApplyUserUpdate applies UpdateUserRequest fields to an existing User
func ApplyUserUpdate(target *user.User, req dto.UpdateUserRequest) {
	if req.FirstName != "" {
		target.FirstName = req.FirstName
	}
	if req.LastName != "" {
		target.LastName = req.LastName
	}
	if req.Email != "" {
		target.Email = req.Email
	}
	if req.Password != "" {
		target.Password = req.Password
	}
	if req.IsActive != nil {
		target.IsActive = *req.IsActive
	}
	if req.Role != "" {
		target.Role = user.Role(req.Role)
	}
}

// ToUserResponse converts a User domain model to a UserResponse DTO
func ToUserResponse(usr *user.User) dto.UserResponse {
	return dto.UserResponse{
		ID:                  usr.ID,
		FirstName:           usr.FirstName,
		LastName:            usr.LastName,
		Email:               usr.Email,
		IsActive:            usr.IsActive,
		Role:                string(usr.Role),
		RoleDisplayName:     user.RoleDisplayName(usr.Role),
		CreatedAt:           usr.CreatedAt,
		UpdatedAt:           usr.UpdatedAt,
		LastLoginAt:         usr.LastLoginAt,
		DisabledAt:          usr.DisabledAt,
		LockedUntil:         usr.LockedUntil,
		FailedLoginAttempts: usr.FailedLoginAttempts,
	}
}

// ToUserListResponse converts a list of Users to UserResponses
func ToUserListResponse(users []user.User) []dto.UserResponse {
	items := make([]dto.UserResponse, len(users))
	for i, usr := range users {
		items[i] = ToUserResponse(&usr)
	}
	return items
}
