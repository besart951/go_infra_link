package project

import (
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterProjectRoutes(protectedV1 *gin.RouterGroup, handlers *Handlers) {
	projects := protectedV1.Group("/projects")
	{
		projects.POST("", handlers.Project.CreateProject)
		projects.GET("", handlers.Project.ListProjects)
		projects.GET("/:id", handlers.Project.GetProject)
		projects.GET("/:id/collaboration", handlers.Project.StreamProjectCollaboration)
		projects.GET("/:id/field-device-options", handlers.FieldDeviceOptions.GetFieldDeviceOptionsForProject)
		projects.POST("/:id/users", handlers.Membership.InviteProjectUser)
		projects.POST("/:id/control-cabinets", handlers.ControlCabinet.CreateProjectControlCabinet)
		projects.GET("/:id/control-cabinets", handlers.ControlCabinet.ListProjectControlCabinets)
		projects.POST("/:id/control-cabinets/:controlCabinetId/copy", handlers.ControlCabinet.CopyProjectControlCabinet)
		projects.PUT("/:id/control-cabinets/:linkId", handlers.ControlCabinet.UpdateProjectControlCabinet)
		projects.DELETE("/:id/control-cabinets/:linkId", handlers.ControlCabinet.DeleteProjectControlCabinet)
		projects.POST("/:id/sps-controllers", handlers.SPSController.CreateProjectSPSController)
		projects.GET("/:id/sps-controllers", handlers.SPSController.ListProjectSPSControllers)
		projects.POST("/:id/sps-controllers/:spsControllerId/copy", handlers.SPSController.CopyProjectSPSController)
		projects.POST("/:id/sps-controller-system-types/:systemTypeId/copy", handlers.SPSController.CopyProjectSPSControllerSystemType)
		projects.PUT("/:id/sps-controllers/:linkId", handlers.SPSController.UpdateProjectSPSController)
		projects.DELETE("/:id/sps-controllers/:linkId", handlers.SPSController.DeleteProjectSPSController)
		projects.POST("/:id/field-devices", handlers.FieldDevice.CreateProjectFieldDevice)
		projects.POST("/:id/field-devices/multi-create", handlers.FieldDevice.MultiCreateProjectFieldDevices)
		projects.GET("/:id/field-devices", handlers.FieldDevice.ListProjectFieldDevices)
		projects.PUT("/:id/field-devices/:linkId", handlers.FieldDevice.UpdateProjectFieldDevice)
		projects.DELETE("/:id/field-devices/:linkId", handlers.FieldDevice.DeleteProjectFieldDevice)
		projects.GET("/:id/users", handlers.Membership.ListProjectUsers)
		projects.DELETE("/:id/users/:userId", handlers.Membership.RemoveProjectUser)
		projects.GET("/:id/object-data", handlers.ObjectData.ListProjectObjectData)
		projects.POST("/:id/object-data", handlers.ObjectData.AddProjectObjectData)
		projects.DELETE("/:id/object-data/:objectDataId", handlers.ObjectData.RemoveProjectObjectData)
		projects.PATCH("/:id", handlers.Project.UpdateProject)
		projects.PUT("/:id", handlers.Project.UpdateProject)
		projects.DELETE("/:id", handlers.Project.DeleteProject)
	}
}

func RegisterPhaseRoutes(protectedV1 *gin.RouterGroup, handlers *Handlers, authChecker middleware.AuthorizationChecker) {
	phases := protectedV1.Group("/phases")
	{
		phases.GET("", handlers.Phase.ListPhases)
		phases.GET("/:id", handlers.Phase.GetPhase)
		phases.POST("", middleware.RequirePermission(authChecker, domainUser.PermissionPhaseCreate), handlers.Phase.CreatePhase)
		phases.PATCH("/:id", middleware.RequirePermission(authChecker, domainUser.PermissionPhaseUpdate), handlers.Phase.UpdatePhase)
		phases.PUT("/:id", middleware.RequirePermission(authChecker, domainUser.PermissionPhaseUpdate), handlers.Phase.UpdatePhase)
		phases.DELETE("/:id", middleware.RequirePermission(authChecker, domainUser.PermissionPhaseDelete), handlers.Phase.DeletePhase)
	}
}
