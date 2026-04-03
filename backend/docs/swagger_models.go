package docs

import (
	authdto "github.com/besart951/go_infra_link/backend/internal/handler/dto/auth"
	commondto "github.com/besart951/go_infra_link/backend/internal/handler/dto/common"
	facilitydto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	notificationdto "github.com/besart951/go_infra_link/backend/internal/handler/dto/notification"
	projectdto "github.com/besart951/go_infra_link/backend/internal/handler/dto/project"
	teamdto "github.com/besart951/go_infra_link/backend/internal/handler/dto/team"
	userdto "github.com/besart951/go_infra_link/backend/internal/handler/dto/user"
)

var (
	_ = commondto.PaginationQuery{}
	_ = commondto.ErrorResponse{}

	_ = authdto.LoginRequest{}
	_ = authdto.AuthUserResponse{}
	_ = authdto.AuthResponse{}

	_ = userdto.AdminSetUserRoleRequest{}
	_ = userdto.CreateUserRequest{}
	_ = userdto.UpdateUserRequest{}
	_ = userdto.UserResponse{}
	_ = userdto.UserListResponse{}
	_ = userdto.CreatePermissionRequest{}
	_ = userdto.UpdatePermissionRequest{}
	_ = userdto.PermissionResponse{}
	_ = userdto.UpdateRolePermissionsRequest{}
	_ = userdto.AddRolePermissionRequest{}
	_ = userdto.RoleResponse{}
	_ = userdto.RolePermissionResponse{}

	_ = teamdto.CreateTeamRequest{}
	_ = teamdto.UpdateTeamRequest{}
	_ = teamdto.TeamResponse{}
	_ = teamdto.TeamListResponse{}
	_ = teamdto.AddTeamMemberRequest{}
	_ = teamdto.TeamMemberResponse{}
	_ = teamdto.TeamMemberListResponse{}

	_ = projectdto.CreateProjectRequest{}
	_ = projectdto.UpdateProjectRequest{}
	_ = projectdto.ProjectResponse{}
	_ = projectdto.ProjectListResponse{}
	_ = projectdto.CreateProjectUserRequest{}
	_ = projectdto.ProjectUserResponse{}
	_ = projectdto.ProjectUserListResponse{}
	_ = projectdto.CreateProjectControlCabinetRequest{}
	_ = projectdto.UpdateProjectControlCabinetRequest{}
	_ = projectdto.ProjectControlCabinetResponse{}
	_ = projectdto.ProjectControlCabinetListResponse{}
	_ = projectdto.CreateProjectSPSControllerRequest{}
	_ = projectdto.UpdateProjectSPSControllerRequest{}
	_ = projectdto.ProjectSPSControllerResponse{}
	_ = projectdto.ProjectSPSControllerListResponse{}
	_ = projectdto.CreateProjectFieldDeviceRequest{}
	_ = projectdto.UpdateProjectFieldDeviceRequest{}
	_ = projectdto.ProjectFieldDeviceResponse{}
	_ = projectdto.ProjectFieldDeviceListResponse{}
	_ = projectdto.CreateProjectObjectDataRequest{}
	_ = projectdto.CreatePhaseRequest{}
	_ = projectdto.UpdatePhaseRequest{}
	_ = projectdto.PhaseResponse{}
	_ = projectdto.PhaseListResponse{}
	_ = projectdto.SwissDateTime{}
	_ = projectdto.ObjectDataResponse{}
	_ = projectdto.ObjectDataListResponse{}
	_ = projectdto.FieldDeviceOptionsResponse{}

	_ = facilitydto.CreateBuildingRequest{}
	_ = facilitydto.UpdateBuildingRequest{}
	_ = facilitydto.BuildingResponse{}
	_ = facilitydto.BuildingListResponse{}
	_ = facilitydto.CreateControlCabinetRequest{}
	_ = facilitydto.UpdateControlCabinetRequest{}
	_ = facilitydto.ControlCabinetResponse{}
	_ = facilitydto.ControlCabinetListResponse{}
	_ = facilitydto.ControlCabinetDeleteImpactResponse{}
	_ = facilitydto.CreateSystemTypeRequest{}
	_ = facilitydto.UpdateSystemTypeRequest{}
	_ = facilitydto.SystemTypeResponse{}
	_ = facilitydto.SystemTypeListResponse{}
	_ = facilitydto.CreateSystemPartRequest{}
	_ = facilitydto.UpdateSystemPartRequest{}
	_ = facilitydto.SystemPartResponse{}
	_ = facilitydto.SystemPartListResponse{}
	_ = facilitydto.CreateApparatRequest{}
	_ = facilitydto.UpdateApparatRequest{}
	_ = facilitydto.ApparatResponse{}
	_ = facilitydto.ApparatListResponse{}
	_ = facilitydto.CreateFieldDeviceRequest{}
	_ = facilitydto.UpdateFieldDeviceRequest{}
	_ = facilitydto.FieldDeviceResponse{}
	_ = facilitydto.FieldDeviceListResponse{}
	_ = facilitydto.AvailableApparatNumbersResponse{}
	_ = facilitydto.FieldDeviceOptionsResponse{}
	_ = facilitydto.MultiCreateFieldDeviceRequest{}
	_ = facilitydto.MultiCreateFieldDeviceResponse{}
	_ = facilitydto.BulkUpdateFieldDeviceRequest{}
	_ = facilitydto.BulkUpdateFieldDeviceResponse{}
	_ = facilitydto.BulkDeleteFieldDeviceRequest{}
	_ = facilitydto.BulkDeleteFieldDeviceResponse{}
	_ = facilitydto.CreateSpecificationRequest{}
	_ = facilitydto.UpdateSpecificationRequest{}
	_ = facilitydto.CreateFieldDeviceSpecificationRequest{}
	_ = facilitydto.UpdateFieldDeviceSpecificationRequest{}
	_ = facilitydto.SpecificationResponse{}
	_ = facilitydto.SpecificationListResponse{}
	_ = facilitydto.CreateBacnetObjectRequest{}
	_ = facilitydto.UpdateBacnetObjectRequest{}
	_ = facilitydto.BacnetObjectInput{}
	_ = facilitydto.BacnetObjectResponse{}
	_ = facilitydto.ObjectDataResponse{}
	_ = facilitydto.ObjectDataListResponse{}
	_ = facilitydto.SPSControllerSystemTypeInput{}
	_ = facilitydto.SPSControllerResponse{}
	_ = facilitydto.SPSControllerListResponse{}
	_ = facilitydto.SPSControllerSystemTypeResponse{}
	_ = facilitydto.SPSControllerSystemTypeListResponse{}
	_ = facilitydto.StateTextResponse{}
	_ = facilitydto.StateTextListResponse{}
	_ = facilitydto.NotificationClassResponse{}
	_ = facilitydto.NotificationClassListResponse{}
	_ = facilitydto.AlarmDefinitionResponse{}
	_ = facilitydto.AlarmDefinitionListResponse{}
	_ = facilitydto.AlarmTypeResponse{}
	_ = facilitydto.AlarmTypeListResponse{}
	_ = facilitydto.AlarmValueInput{}
	_ = facilitydto.AlarmValueResponse{}
	_ = facilitydto.AlarmValuesResponse{}
	_ = facilitydto.UnitResponse{}
	_ = facilitydto.UnitListResponse{}
	_ = facilitydto.AlarmFieldResponse{}
	_ = facilitydto.AlarmFieldListResponse{}
	_ = facilitydto.AlarmTypeFieldResponse{}
	_ = facilitydto.PutAlarmValuesRequest{}
	_ = facilitydto.ValidateBuildingRequest{}
	_ = facilitydto.ValidateControlCabinetRequest{}
	_ = facilitydto.ValidateSPSControllerRequest{}
	_ = facilitydto.CreateFieldDeviceExportRequest{}
	_ = facilitydto.FieldDeviceExportJobResponse{}

	_ = notificationdto.UpsertSMTPSettingsRequest{}
	_ = notificationdto.SendSMTPTestEmailRequest{}
	_ = notificationdto.SMTPSettingsResponse{}
)
