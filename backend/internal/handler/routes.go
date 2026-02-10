package handler

import (
	domainAuth "github.com/besart951/go_infra_link/backend/internal/domain/auth"
	domainTeam "github.com/besart951/go_infra_link/backend/internal/domain/team"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	facilityhandler "github.com/besart951/go_infra_link/backend/internal/handler/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	ProjectHandler         *ProjectHandler
	UserHandler            *UserHandler
	AuthHandler            *AuthHandler
	TeamHandler            *TeamHandler
	AdminHandler           *AdminHandler
	PhaseHandler           *PhaseHandler
	PhasePermissionHandler *PhasePermissionHandler

	FacilityBuildingHandler       *facilityhandler.BuildingHandler
	FacilitySystemTypeHandler     *facilityhandler.SystemTypeHandler
	FacilitySystemPartHandler     *facilityhandler.SystemPartHandler
	FacilityApparatHandler        *facilityhandler.ApparatHandler
	FacilityFieldDeviceHandler    *facilityhandler.FieldDeviceHandler
	FacilityBacnetObjectHandler   *facilityhandler.BacnetObjectHandler
	FacilityControlCabinetHandler *facilityhandler.ControlCabinetHandler
	FacilitySPSControllerHandler  *facilityhandler.SPSControllerHandler
	FacilityValidationHandler     *facilityhandler.ValidationHandler

	FacilityStateTextHandler               *facilityhandler.StateTextHandler
	FacilityNotificationClassHandler       *facilityhandler.NotificationClassHandler
	FacilityAlarmDefinitionHandler         *facilityhandler.AlarmDefinitionHandler
	FacilityObjectDataHandler              *facilityhandler.ObjectDataHandler
	FacilitySPSControllerSystemTypeHandler *facilityhandler.SPSControllerSystemTypeHandler
}

// RegisterRoutes registers all API routes
func RegisterRoutes(r *gin.Engine, handlers *Handlers, tokenValidator domainAuth.TokenValidator, authChecker middleware.AuthorizationChecker, userStatusSvc middleware.UserStatusService) {
	// Public API v1 group (login only)
	publicV1 := r.Group("/api/v1")
	publicAuth := publicV1.Group("/auth")
	{
		publicAuth.POST("/login", handlers.AuthHandler.Login)
		publicAuth.POST("/dev-login", handlers.AuthHandler.DevLogin)
		publicAuth.POST("/password-reset/confirm", handlers.AuthHandler.ConfirmPasswordReset)
	}

	// CSRF-protected auth endpoints (no access token required)
	authCsrf := publicV1.Group("/auth")
	authCsrf.Use(middleware.CSRFMiddleware())
	{
		authCsrf.POST("/refresh", handlers.AuthHandler.Refresh)
		authCsrf.POST("/logout", handlers.AuthHandler.Logout)
	}

	// Protected API v1 group (all other routes)
	protectedV1 := r.Group("/api/v1")
	protectedV1.Use(middleware.AuthGuard(tokenValidator))
	protectedV1.Use(middleware.AccountStatusGuard(userStatusSvc))
	protectedV1.Use(middleware.CSRFMiddleware())

	// Project routes
	projects := protectedV1.Group("/projects")
	{
		projects.POST("", handlers.ProjectHandler.CreateProject)
		projects.GET("", handlers.ProjectHandler.ListProjects)
		projects.GET("/:id", handlers.ProjectHandler.GetProject)
		projects.GET("/:id/field-device-options", handlers.FacilityFieldDeviceHandler.GetFieldDeviceOptionsForProject)
		projects.POST("/:id/users", handlers.ProjectHandler.InviteProjectUser)
		projects.POST("/:id/control-cabinets", handlers.ProjectHandler.CreateProjectControlCabinet)
		projects.GET("/:id/control-cabinets", handlers.ProjectHandler.ListProjectControlCabinets)
		projects.PUT("/:id/control-cabinets/:linkId", handlers.ProjectHandler.UpdateProjectControlCabinet)
		projects.DELETE("/:id/control-cabinets/:linkId", handlers.ProjectHandler.DeleteProjectControlCabinet)
		projects.POST("/:id/sps-controllers", handlers.ProjectHandler.CreateProjectSPSController)
		projects.GET("/:id/sps-controllers", handlers.ProjectHandler.ListProjectSPSControllers)
		projects.PUT("/:id/sps-controllers/:linkId", handlers.ProjectHandler.UpdateProjectSPSController)
		projects.DELETE("/:id/sps-controllers/:linkId", handlers.ProjectHandler.DeleteProjectSPSController)
		projects.POST("/:id/field-devices", handlers.ProjectHandler.CreateProjectFieldDevice)
		projects.GET("/:id/field-devices", handlers.ProjectHandler.ListProjectFieldDevices)
		projects.PUT("/:id/field-devices/:linkId", handlers.ProjectHandler.UpdateProjectFieldDevice)
		projects.DELETE("/:id/field-devices/:linkId", handlers.ProjectHandler.DeleteProjectFieldDevice)
		projects.GET("/:id/users", handlers.ProjectHandler.ListProjectUsers)
		projects.DELETE("/:id/users/:userId", handlers.ProjectHandler.RemoveProjectUser)
		projects.GET("/:id/object-data", handlers.ProjectHandler.ListProjectObjectData)
		projects.POST("/:id/object-data", handlers.ProjectHandler.AddProjectObjectData)
		projects.DELETE("/:id/object-data/:objectDataId", handlers.ProjectHandler.RemoveProjectObjectData)
		projects.PUT("/:id", handlers.ProjectHandler.UpdateProject)
		projects.DELETE("/:id", handlers.ProjectHandler.DeleteProject)
	}

	// Phase routes - Anyone authenticated can view, but only admins can modify
	phases := protectedV1.Group("/phases")
	{
		phases.GET("", handlers.PhaseHandler.ListPhases)
		phases.GET("/:id", handlers.PhaseHandler.GetPhase)
		// Creating, updating, and deleting phases requires admin role
		phases.POST("", middleware.RequireGlobalRole(authChecker, domainUser.RoleAdminFZAG), handlers.PhaseHandler.CreatePhase)
		phases.PUT("/:id", middleware.RequireGlobalRole(authChecker, domainUser.RoleAdminFZAG), handlers.PhaseHandler.UpdatePhase)
		phases.DELETE("/:id", middleware.RequireGlobalRole(authChecker, domainUser.RoleAdminFZAG), handlers.PhaseHandler.DeletePhase)
	}

	// Phase Permission routes - Only admins can manage permissions
	phasePermissions := protectedV1.Group("/phase-permissions")
	phasePermissions.Use(middleware.RequireGlobalRole(authChecker, domainUser.RoleAdminFZAG))
	{
		phasePermissions.POST("", handlers.PhasePermissionHandler.CreatePhasePermission)
		phasePermissions.GET("", handlers.PhasePermissionHandler.ListPhasePermissions)
		phasePermissions.GET("/:id", handlers.PhasePermissionHandler.GetPhasePermission)
		phasePermissions.PUT("/:id", handlers.PhasePermissionHandler.UpdatePhasePermission)
		phasePermissions.DELETE("/:id", handlers.PhasePermissionHandler.DeletePhasePermission)
	}

	// User routes
	users := protectedV1.Group("/users")
	{
		// Anyone authenticated can get their allowed roles
		users.GET("/allowed-roles", handlers.UserHandler.GetAllowedRoles)
	}

	// Admin-only user management routes
	usersAdmin := protectedV1.Group("/users")
	usersAdmin.Use(middleware.RequireGlobalRole(authChecker, domainUser.RoleAdminFZAG))
	{
		usersAdmin.POST("", handlers.UserHandler.CreateUser)
		usersAdmin.GET("", handlers.UserHandler.ListUsers)
		usersAdmin.GET("/:id", handlers.UserHandler.GetUser)
		usersAdmin.PUT("/:id", handlers.UserHandler.UpdateUser)
		usersAdmin.DELETE("/:id", handlers.UserHandler.DeleteUser)
	}

	// Team routes
	teams := protectedV1.Group("/teams")
	{
		teams.POST("", middleware.RequireGlobalRole(authChecker, domainUser.RoleAdminFZAG), handlers.TeamHandler.CreateTeam)
		teams.GET("", middleware.RequireGlobalRole(authChecker, domainUser.RoleAdminFZAG), handlers.TeamHandler.ListTeams)

		teams.GET("/:id", middleware.RequireTeamRole(authChecker, "id", domainTeam.MemberRoleMember), handlers.TeamHandler.GetTeam)
		teams.PUT("/:id", middleware.RequireTeamRole(authChecker, "id", domainTeam.MemberRoleManager), handlers.TeamHandler.UpdateTeam)
		teams.DELETE("/:id", middleware.RequireTeamRole(authChecker, "id", domainTeam.MemberRoleOwner), handlers.TeamHandler.DeleteTeam)

		teams.POST("/:id/members", middleware.RequireTeamRole(authChecker, "id", domainTeam.MemberRoleManager), handlers.TeamHandler.AddMember)
		teams.GET("/:id/members", middleware.RequireTeamRole(authChecker, "id", domainTeam.MemberRoleMember), handlers.TeamHandler.ListMembers)
		teams.DELETE("/:id/members/:userId", middleware.RequireTeamRole(authChecker, "id", domainTeam.MemberRoleManager), handlers.TeamHandler.RemoveMember)
	}

	// Admin routes
	admin := protectedV1.Group("/admin")
	admin.Use(middleware.RequireGlobalRole(authChecker, domainUser.RoleAdminFZAG))
	{
		admin.POST("/users/:id/password-reset", handlers.AdminHandler.ResetUserPassword)
		admin.POST("/users/:id/disable", handlers.AdminHandler.DisableUser)
		admin.POST("/users/:id/enable", handlers.AdminHandler.EnableUser)
		admin.POST("/users/:id/lock", handlers.AdminHandler.LockUser)
		admin.POST("/users/:id/unlock", handlers.AdminHandler.UnlockUser)
		admin.POST("/users/:id/role", handlers.AdminHandler.SetUserRole)

		admin.GET("/login-attempts", handlers.AdminHandler.ListLoginAttempts)
	}

	// Auth routes (protected)
	authProtected := protectedV1.Group("/auth")
	{
		authProtected.GET("/me", handlers.AuthHandler.Me)
	}

	// Facility routes
	facility := protectedV1.Group("/facility")
	{
		facility.POST("/buildings/validate", handlers.FacilityValidationHandler.ValidateBuilding)
		facility.POST("/buildings", handlers.FacilityBuildingHandler.CreateBuilding)
		facility.GET("/buildings", handlers.FacilityBuildingHandler.ListBuildings)
		facility.GET("/buildings/:id", handlers.FacilityBuildingHandler.GetBuilding)
		facility.PUT("/buildings/:id", handlers.FacilityBuildingHandler.UpdateBuilding)
		facility.DELETE("/buildings/:id", handlers.FacilityBuildingHandler.DeleteBuilding)

		facility.POST("/system-types", handlers.FacilitySystemTypeHandler.CreateSystemType)
		facility.GET("/system-types", handlers.FacilitySystemTypeHandler.ListSystemTypes)
		facility.GET("/system-types/:id", handlers.FacilitySystemTypeHandler.GetSystemType)
		facility.PUT("/system-types/:id", handlers.FacilitySystemTypeHandler.UpdateSystemType)
		facility.DELETE("/system-types/:id", handlers.FacilitySystemTypeHandler.DeleteSystemType)

		facility.POST("/system-parts", handlers.FacilitySystemPartHandler.CreateSystemPart)
		facility.GET("/system-parts", handlers.FacilitySystemPartHandler.ListSystemParts)
		facility.GET("/system-parts/:id", handlers.FacilitySystemPartHandler.GetSystemPart)
		facility.PUT("/system-parts/:id", handlers.FacilitySystemPartHandler.UpdateSystemPart)
		facility.DELETE("/system-parts/:id", handlers.FacilitySystemPartHandler.DeleteSystemPart)

		facility.POST("/apparats", handlers.FacilityApparatHandler.CreateApparat)
		facility.GET("/apparats", handlers.FacilityApparatHandler.ListApparats)
		facility.GET("/apparats/:id", handlers.FacilityApparatHandler.GetApparat)
		facility.PUT("/apparats/:id", handlers.FacilityApparatHandler.UpdateApparat)
		facility.DELETE("/apparats/:id", handlers.FacilityApparatHandler.DeleteApparat)

		facility.POST("/control-cabinets/validate", handlers.FacilityValidationHandler.ValidateControlCabinet)
		facility.POST("/control-cabinets", handlers.FacilityControlCabinetHandler.CreateControlCabinet)
		facility.GET("/control-cabinets", handlers.FacilityControlCabinetHandler.ListControlCabinets)
		facility.GET("/control-cabinets/:id", handlers.FacilityControlCabinetHandler.GetControlCabinet)
		facility.GET("/control-cabinets/:id/delete-impact", handlers.FacilityControlCabinetHandler.GetControlCabinetDeleteImpact)
		facility.PUT("/control-cabinets/:id", handlers.FacilityControlCabinetHandler.UpdateControlCabinet)
		facility.DELETE("/control-cabinets/:id", handlers.FacilityControlCabinetHandler.DeleteControlCabinet)

		facility.POST("/field-devices", handlers.FacilityFieldDeviceHandler.CreateFieldDevice)
		facility.POST("/field-devices/multi-create", handlers.FacilityFieldDeviceHandler.MultiCreateFieldDevices)
		facility.GET("/field-devices/options", handlers.FacilityFieldDeviceHandler.GetFieldDeviceOptions)
		facility.GET("/field-devices/available-apparat-nr", handlers.FacilityFieldDeviceHandler.ListAvailableApparatNumbers)
		facility.GET("/field-devices", handlers.FacilityFieldDeviceHandler.ListFieldDevices)
		facility.GET("/field-devices/:id", handlers.FacilityFieldDeviceHandler.GetFieldDevice)
		facility.GET("/field-devices/:id/bacnet-objects", handlers.FacilityFieldDeviceHandler.ListFieldDeviceBacnetObjects)
		facility.POST("/field-devices/:id/specification", handlers.FacilityFieldDeviceHandler.CreateFieldDeviceSpecification)
		facility.PUT("/field-devices/:id/specification", handlers.FacilityFieldDeviceHandler.UpdateFieldDeviceSpecification)
		facility.PUT("/field-devices/:id", handlers.FacilityFieldDeviceHandler.UpdateFieldDevice)
		facility.DELETE("/field-devices/:id", handlers.FacilityFieldDeviceHandler.DeleteFieldDevice)
		facility.PATCH("/field-devices/bulk-update", handlers.FacilityFieldDeviceHandler.BulkUpdateFieldDevices)
		facility.DELETE("/field-devices/bulk-delete", handlers.FacilityFieldDeviceHandler.BulkDeleteFieldDevices)

		facility.POST("/bacnet-objects", handlers.FacilityBacnetObjectHandler.CreateBacnetObject)
		facility.PUT("/bacnet-objects/:id", handlers.FacilityBacnetObjectHandler.UpdateBacnetObject)

		facility.POST("/sps-controllers/validate", handlers.FacilityValidationHandler.ValidateSPSController)
		facility.POST("/sps-controllers", handlers.FacilitySPSControllerHandler.CreateSPSController)
		facility.GET("/sps-controllers", handlers.FacilitySPSControllerHandler.ListSPSControllers)
		facility.GET("/sps-controllers/next-ga-device", handlers.FacilitySPSControllerHandler.GetNextAvailableGADevice)
		facility.GET("/sps-controllers/:id", handlers.FacilitySPSControllerHandler.GetSPSController)
		facility.PUT("/sps-controllers/:id", handlers.FacilitySPSControllerHandler.UpdateSPSController)
		facility.DELETE("/sps-controllers/:id", handlers.FacilitySPSControllerHandler.DeleteSPSController)

		facility.GET("/state-texts", handlers.FacilityStateTextHandler.ListStateTexts)
		facility.GET("/state-texts/:id", handlers.FacilityStateTextHandler.GetStateText)
		facility.POST("/state-texts", handlers.FacilityStateTextHandler.CreateStateText)
		facility.PUT("/state-texts/:id", handlers.FacilityStateTextHandler.UpdateStateText)
		facility.DELETE("/state-texts/:id", handlers.FacilityStateTextHandler.DeleteStateText)

		facility.GET("/notification-classes", handlers.FacilityNotificationClassHandler.ListNotificationClasses)
		facility.GET("/notification-classes/:id", handlers.FacilityNotificationClassHandler.GetNotificationClass)
		facility.POST("/notification-classes", handlers.FacilityNotificationClassHandler.CreateNotificationClass)
		facility.PUT("/notification-classes/:id", handlers.FacilityNotificationClassHandler.UpdateNotificationClass)
		facility.DELETE("/notification-classes/:id", handlers.FacilityNotificationClassHandler.DeleteNotificationClass)

		facility.GET("/alarm-definitions", handlers.FacilityAlarmDefinitionHandler.ListAlarmDefinitions)
		facility.GET("/alarm-definitions/:id", handlers.FacilityAlarmDefinitionHandler.GetAlarmDefinition)
		facility.POST("/alarm-definitions", handlers.FacilityAlarmDefinitionHandler.CreateAlarmDefinition)
		facility.PUT("/alarm-definitions/:id", handlers.FacilityAlarmDefinitionHandler.UpdateAlarmDefinition)
		facility.DELETE("/alarm-definitions/:id", handlers.FacilityAlarmDefinitionHandler.DeleteAlarmDefinition)

		facility.GET("/object-data", handlers.FacilityObjectDataHandler.ListObjectData)
		facility.GET("/object-data/:id", handlers.FacilityObjectDataHandler.GetObjectData)
		facility.GET("/object-data/:id/bacnet-objects", handlers.FacilityObjectDataHandler.GetObjectDataBacnetObjects)
		facility.POST("/object-data", handlers.FacilityObjectDataHandler.CreateObjectData)
		facility.PUT("/object-data/:id", handlers.FacilityObjectDataHandler.UpdateObjectData)
		facility.DELETE("/object-data/:id", handlers.FacilityObjectDataHandler.DeleteObjectData)

		facility.GET("/sps-controller-system-types", handlers.FacilitySPSControllerSystemTypeHandler.ListSPSControllerSystemTypes)
	}
}
