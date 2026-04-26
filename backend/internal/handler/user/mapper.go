package user

import (
	"github.com/besart951/go_infra_link/backend/internal/domain/user"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/user"
	userdirectory "github.com/besart951/go_infra_link/backend/internal/service/userdirectory"
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

func ToUserDirectoryListResponse(result *userdirectory.ListResult) dto.UserDirectoryListResponse {
	items := make([]dto.UserDirectoryUserResponse, len(result.Items))
	for i, item := range result.Items {
		teams := make([]dto.UserDirectoryTeamResponse, len(item.Teams))
		for j, team := range item.Teams {
			teams[j] = dto.UserDirectoryTeamResponse{ID: team.ID, Name: team.Name}
		}

		items[i] = dto.UserDirectoryUserResponse{
			ID:                  item.User.ID,
			FirstName:           item.User.FirstName,
			LastName:            item.User.LastName,
			Email:               item.User.Email,
			IsActive:            item.User.IsActive,
			Role:                string(item.User.Role),
			RoleDisplayName:     user.RoleDisplayName(item.User.Role),
			CreatedAt:           item.User.CreatedAt,
			UpdatedAt:           item.User.UpdatedAt,
			LastLoginAt:         item.User.LastLoginAt,
			DisabledAt:          item.User.DisabledAt,
			LockedUntil:         item.User.LockedUntil,
			FailedLoginAttempts: item.User.FailedLoginAttempts,
			Teams:               teams,
			Capabilities: dto.UserDirectoryCapabilitiesResponse{
				CanUpdate:     item.Capabilities.CanUpdate,
				CanDelete:     item.Capabilities.CanDelete,
				CanDisable:    item.Capabilities.CanDisable,
				CanEnable:     item.Capabilities.CanEnable,
				CanChangeRole: item.Capabilities.CanChangeRole,
			},
		}
	}

	teams := make([]dto.UserDirectoryTeamFilterResponse, len(result.Teams))
	for i, team := range result.Teams {
		teams[i] = dto.UserDirectoryTeamFilterResponse{ID: team.ID, Name: team.Name, Count: team.Count}
	}

	return dto.UserDirectoryListResponse{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
		Teams:      teams,
		Capabilities: dto.UserDirectoryPageCapabilitiesResponse{
			CanCreateUser: result.PageCapabilities.CanCreateUser,
		},
	}
}
