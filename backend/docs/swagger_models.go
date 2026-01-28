package docs

import "github.com/besart951/go_infra_link/backend/internal/handler/dto"

var (
	_ = dto.PaginationQuery{}
	_ = dto.ErrorResponse{}

	_ = dto.LoginRequest{}
	_ = dto.AuthUserResponse{}
	_ = dto.AuthResponse{}
	_ = dto.PasswordResetConfirmRequest{}

	_ = dto.AdminPasswordResetResponse{}
	_ = dto.AdminLockUserRequest{}
	_ = dto.AdminSetUserRoleRequest{}
	_ = dto.LoginAttemptResponse{}
	_ = dto.LoginAttemptListResponse{}

	_ = dto.CreateUserRequest{}
	_ = dto.UpdateUserRequest{}
	_ = dto.UserResponse{}
	_ = dto.UserListResponse{}

	_ = dto.CreateTeamRequest{}
	_ = dto.UpdateTeamRequest{}
	_ = dto.TeamResponse{}
	_ = dto.TeamListResponse{}
	_ = dto.AddTeamMemberRequest{}
	_ = dto.TeamMemberResponse{}
	_ = dto.TeamMemberListResponse{}

	_ = dto.CreateProjectRequest{}
	_ = dto.UpdateProjectRequest{}
	_ = dto.ProjectResponse{}
	_ = dto.ProjectListResponse{}

	_ = dto.CreateProjectUserRequest{}
	_ = dto.ProjectUserResponse{}
	_ = dto.CreateProjectControlCabinetRequest{}
	_ = dto.UpdateProjectControlCabinetRequest{}
	_ = dto.ProjectControlCabinetResponse{}
	_ = dto.ProjectControlCabinetListResponse{}
	_ = dto.CreateProjectSPSControllerRequest{}
	_ = dto.UpdateProjectSPSControllerRequest{}
	_ = dto.ProjectSPSControllerResponse{}
	_ = dto.ProjectSPSControllerListResponse{}
	_ = dto.CreateProjectFieldDeviceRequest{}
	_ = dto.UpdateProjectFieldDeviceRequest{}
	_ = dto.ProjectFieldDeviceResponse{}
	_ = dto.ProjectFieldDeviceListResponse{}

	_ = dto.CreateBuildingRequest{}
	_ = dto.UpdateBuildingRequest{}
	_ = dto.BuildingResponse{}
	_ = dto.BuildingListResponse{}

	_ = dto.CreateControlCabinetRequest{}
	_ = dto.UpdateControlCabinetRequest{}
	_ = dto.ControlCabinetResponse{}
	_ = dto.ControlCabinetListResponse{}

	_ = dto.CreateSystemTypeRequest{}
	_ = dto.UpdateSystemTypeRequest{}
	_ = dto.SystemTypeResponse{}
	_ = dto.SystemTypeListResponse{}

	_ = dto.CreateSystemPartRequest{}
	_ = dto.UpdateSystemPartRequest{}
	_ = dto.SystemPartResponse{}
	_ = dto.SystemPartListResponse{}

	_ = dto.CreateApparatRequest{}
	_ = dto.UpdateApparatRequest{}
	_ = dto.ApparatResponse{}
	_ = dto.ApparatListResponse{}

	_ = dto.CreateFieldDeviceRequest{}
	_ = dto.UpdateFieldDeviceRequest{}
	_ = dto.FieldDeviceResponse{}
	_ = dto.FieldDeviceListResponse{}

	_ = dto.CreateSpecificationRequest{}
	_ = dto.UpdateSpecificationRequest{}
	_ = dto.CreateFieldDeviceSpecificationRequest{}
	_ = dto.UpdateFieldDeviceSpecificationRequest{}
	_ = dto.SpecificationResponse{}
	_ = dto.SpecificationListResponse{}

	_ = dto.CreateBacnetObjectRequest{}
	_ = dto.UpdateBacnetObjectRequest{}
	_ = dto.BacnetObjectInput{}
	_ = dto.BacnetObjectResponse{}

	_ = dto.ObjectDataResponse{}
	_ = dto.ObjectDataListResponse{}

	_ = dto.SPSControllerSystemTypeInput{}
	_ = dto.SPSControllerResponse{}
	_ = dto.SPSControllerListResponse{}
	_ = dto.SPSControllerSystemTypeResponse{}
	_ = dto.SPSControllerSystemTypeListResponse{}

	_ = dto.StateTextResponse{}
	_ = dto.StateTextListResponse{}
	_ = dto.NotificationClassResponse{}
	_ = dto.NotificationClassListResponse{}
	_ = dto.AlarmDefinitionResponse{}
	_ = dto.AlarmDefinitionListResponse{}
	_ = dto.ApparatResponse{}
	_ = dto.ApparatListResponse{}
	_ = dto.BacnetObjectResponse{}
)
