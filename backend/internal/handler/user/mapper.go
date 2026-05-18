package user

import (
	"github.com/besart951/go_infra_link/backend/internal/domain/user"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/user"
	userdirectory "github.com/besart951/go_infra_link/backend/internal/service/userdirectory"
	userregistration "github.com/besart951/go_infra_link/backend/internal/service/userregistration"
	"github.com/google/uuid"
)

// ToUserModel converts a CreateUserRequest to a User domain model
func ToUserModel(req dto.CreateUserRequest) *user.User {
	usr := &user.User{
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Email:       user.EmailPtr(req.Email),
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
		target.SetEmail(req.Email)
	}
}

// ToUserResponse converts a User domain model to a UserResponse DTO
func ToUserResponse(usr *user.User) dto.UserResponse {
	firstName := usr.FirstName
	lastName := usr.LastName
	email := usr.EmailValue()
	if usr.IsIdentityHidden() {
		firstName = "Deleted"
		lastName = "User"
		email = ""
	}

	return dto.UserResponse{
		ID:                  usr.ID,
		FirstName:           firstName,
		LastName:            lastName,
		Email:               email,
		IsActive:            usr.IsActive,
		Role:                string(usr.Role),
		RoleDisplayName:     user.RoleDisplayName(usr.Role),
		CreatedAt:           usr.CreatedAt,
		UpdatedAt:           usr.UpdatedAt,
		LastLoginAt:         usr.LastLoginAt,
		DisabledAt:          usr.DisabledAt,
		LockedUntil:         usr.LockedUntil,
		FailedLoginAttempts: usr.FailedLoginAttempts,
		IsDeleted:           usr.IsDeleted(),
		IsAnonymized:        usr.IsAnonymized(),
		DeletedAt:           usr.DeletedAt,
		RestoreUntil:        usr.RestoreUntil,
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

func ToUserDirectoryListResponse(result *userdirectory.ListResult, registrationProcesses map[uuid.UUID]*userregistration.Process) dto.UserDirectoryListResponse {
	showDeletedEmail := result != nil && result.PageCapabilities.CanReadDeleted

	items := make([]dto.UserDirectoryUserResponse, len(result.Items))
	for i, item := range result.Items {
		teams := make([]dto.UserDirectoryTeamResponse, len(item.Teams))
		for j, team := range item.Teams {
			teams[j] = dto.UserDirectoryTeamResponse{ID: team.ID, Name: team.Name}
		}
		process := registrationProcesses[item.User.ID]
		capabilities := item.Capabilities

		items[i] = dto.UserDirectoryUserResponse{
			ID:                  item.User.ID,
			FirstName:           projectedFirstName(item.User),
			LastName:            projectedLastName(item.User),
			Email:               projectedEmail(item.User, showDeletedEmail),
			IsActive:            item.User.IsActive,
			Role:                string(item.User.Role),
			RoleDisplayName:     user.RoleDisplayName(item.User.Role),
			CreatedAt:           item.User.CreatedAt,
			UpdatedAt:           item.User.UpdatedAt,
			LastLoginAt:         item.User.LastLoginAt,
			DisabledAt:          item.User.DisabledAt,
			LockedUntil:         item.User.LockedUntil,
			FailedLoginAttempts: item.User.FailedLoginAttempts,
			IsDeleted:           item.User.IsDeleted(),
			IsAnonymized:        item.User.IsAnonymized(),
			DeletedAt:           item.User.DeletedAt,
			RestoreUntil:        item.User.RestoreUntil,
			Teams:               teams,
			Capabilities: dto.UserDirectoryCapabilitiesResponse{
				CanUpdate:     capabilities.CanUpdate,
				CanDelete:     capabilities.CanDelete,
				CanDisable:    capabilities.CanDisable,
				CanEnable:     capabilities.CanEnable,
				CanRestore:    capabilities.CanRestore,
				CanChangeRole: capabilities.CanChangeRole,
			},
			RegistrationProcess: ToRegistrationProcessResponse(process),
		}
	}

	teams := make([]dto.UserDirectoryTeamFilterResponse, len(result.Teams))
	for i, team := range result.Teams {
		teams[i] = dto.UserDirectoryTeamFilterResponse{ID: team.ID, Name: team.Name, Count: team.Count}
	}
	roles := make([]dto.UserDirectoryRoleFilterResponse, len(result.Roles))
	for i, role := range result.Roles {
		roles[i] = dto.UserDirectoryRoleFilterResponse{
			Role:        string(role.Role),
			DisplayName: role.DisplayName,
			Count:       role.Count,
		}
	}

	return dto.UserDirectoryListResponse{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
		Teams:      teams,
		Roles:      roles,
		Capabilities: dto.UserDirectoryPageCapabilitiesResponse{
			CanCreateUser:  result.PageCapabilities.CanCreateUser,
			CanReadDeleted: result.PageCapabilities.CanReadDeleted,
		},
	}
}

func projectedFirstName(usr user.User) string {
	if usr.IsIdentityHidden() {
		return "Deleted"
	}
	return usr.FirstName
}

func projectedLastName(usr user.User) string {
	if usr.IsIdentityHidden() {
		return "User"
	}
	return usr.LastName
}

func projectedEmail(usr user.User, showDeletedEmail bool) string {
	if usr.IsIdentityHidden() {
		if showDeletedEmail && usr.IsDeleted() && !usr.IsAnonymized() {
			return usr.EmailValue()
		}
		return ""
	}
	return usr.EmailValue()
}

func ToRegistrationProcessResponse(process *userregistration.Process) *dto.RegistrationProcessResponse {
	if process == nil {
		return nil
	}
	steps := make([]dto.RegistrationProcessStepResponse, len(process.Steps))
	for i, step := range process.Steps {
		steps[i] = dto.RegistrationProcessStepResponse{
			Key:       step.Key,
			Label:     step.Label,
			Status:    step.Status,
			Timestamp: step.Timestamp,
		}
	}
	return &dto.RegistrationProcessResponse{
		Status:            process.Status,
		EmailStatus:       process.EmailStatus,
		Steps:             steps,
		CanResend:         process.CanResend,
		ExpiresAt:         process.ExpiresAt,
		AcceptedAt:        process.AcceptedAt,
		LastSentAt:        process.LastSentAt,
		ResendAvailableAt: process.ResendAvailableAt,
		SendCount:         process.SendCount,
		LastError:         process.LastError,
	}
}
