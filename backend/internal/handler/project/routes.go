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
		projects.GET("/:id/events", handlers.Project.StreamProjectEvents)
		projects.GET("/:id/collaboration", handlers.Project.StreamProjectCollaboration)
		projects.GET("/:id/field-device-options", handlers.FieldDeviceOptions.GetFieldDeviceOptionsForProject)
		projects.POST("/:id/users", handlers.Project.InviteProjectUser)
		projects.POST("/:id/control-cabinets", handlers.Project.CreateProjectControlCabinet)
		projects.GET("/:id/control-cabinets", handlers.Project.ListProjectControlCabinets)
		projects.POST("/:id/control-cabinets/:controlCabinetId/copy", handlers.Project.CopyProjectControlCabinet)
		projects.PUT("/:id/control-cabinets/:linkId", handlers.Project.UpdateProjectControlCabinet)
		projects.DELETE("/:id/control-cabinets/:linkId", handlers.Project.DeleteProjectControlCabinet)
		projects.POST("/:id/sps-controllers", handlers.Project.CreateProjectSPSController)
		projects.GET("/:id/sps-controllers", handlers.Project.ListProjectSPSControllers)
		projects.POST("/:id/sps-controllers/:spsControllerId/copy", handlers.Project.CopyProjectSPSController)
		projects.POST("/:id/sps-controller-system-types/:systemTypeId/copy", handlers.Project.CopyProjectSPSControllerSystemType)
		projects.PUT("/:id/sps-controllers/:linkId", handlers.Project.UpdateProjectSPSController)
		projects.DELETE("/:id/sps-controllers/:linkId", handlers.Project.DeleteProjectSPSController)
		projects.POST("/:id/field-devices", handlers.Project.CreateProjectFieldDevice)
		projects.GET("/:id/field-devices", handlers.Project.ListProjectFieldDevices)
		projects.PUT("/:id/field-devices/:linkId", handlers.Project.UpdateProjectFieldDevice)
		projects.DELETE("/:id/field-devices/:linkId", handlers.Project.DeleteProjectFieldDevice)
		projects.GET("/:id/users", handlers.Project.ListProjectUsers)
		projects.DELETE("/:id/users/:userId", handlers.Project.RemoveProjectUser)
		projects.GET("/:id/object-data", handlers.Project.ListProjectObjectData)
		projects.POST("/:id/object-data", handlers.Project.AddProjectObjectData)
		projects.DELETE("/:id/object-data/:objectDataId", handlers.Project.RemoveProjectObjectData)
		projects.PUT("/:id", handlers.Project.UpdateProject)
		projects.DELETE("/:id", handlers.Project.DeleteProject)
	}
}

func RegisterPhaseRoutes(protectedV1 *gin.RouterGroup, handlers *Handlers, authChecker middleware.AuthorizationChecker) {
	phases := protectedV1.Group("/phases")
	{
		phases.GET("", handlers.Phase.ListPhases)
		phases.GET("/:id", handlers.Phase.GetPhase)
		phases.POST("", middleware.RequireGlobalRole(authChecker, domainUser.RoleAdminFZAG), handlers.Phase.CreatePhase)
		phases.PUT("/:id", middleware.RequireGlobalRole(authChecker, domainUser.RoleAdminFZAG), handlers.Phase.UpdatePhase)
		phases.DELETE("/:id", middleware.RequireGlobalRole(authChecker, domainUser.RoleAdminFZAG), handlers.Phase.DeletePhase)
	}
}
