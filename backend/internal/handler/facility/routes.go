package facility

import (
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/handler/facility/alarm"
	"github.com/besart951/go_infra_link/backend/internal/handler/facility/fielddevice"
	"github.com/besart951/go_infra_link/backend/internal/handler/facility/hierarchy"
	"github.com/besart951/go_infra_link/backend/internal/handler/facility/internal/routing"
	"github.com/besart951/go_infra_link/backend/internal/handler/facility/objectdata"
	"github.com/besart951/go_infra_link/backend/internal/handler/facility/reference"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/gin-gonic/gin"
)

type routeDefinition = routing.Definition

func RegisterRoutes(protectedV1 *gin.RouterGroup, handlers *Handlers, authChecker middleware.AuthorizationChecker) {
	facility := protectedV1.Group("/facility")
	registerRoutes(facility, authChecker, routeDefinitions(handlers), handlers.Realtime)
	handlers.DeleteImpact.SetAuthorizationChecker(authChecker)
	handlers.ReferenceData.SetAuthorizationChecker(authChecker)
	handlers.Details.SetAuthorizationChecker(authChecker)
	facility.GET("/buildings/:id/detail", middleware.RequirePermission(authChecker, domainUser.PermissionBuildingRead), handlers.Details.GetBuildingDetail)
	facility.GET("/control-cabinets/:id/detail", middleware.RequirePermission(authChecker, domainUser.PermissionControlCabinetRead), handlers.Details.GetControlCabinetDetail)
	facility.GET("/sps-controllers/:id/detail", middleware.RequirePermission(authChecker, domainUser.PermissionSPSControllerRead), handlers.Details.GetSPSControllerDetail)
	facility.GET("/sps-controller-system-types/:id/detail", middleware.RequirePermission(authChecker, domainUser.PermissionSPSControllerSystemTypeRead), handlers.Details.GetSPSControllerSystemTypeDetail)
	facility.GET("/field-devices/:id/detail", middleware.RequirePermission(authChecker, domainUser.PermissionFieldDeviceRead), handlers.Details.GetFieldDeviceDetail)
	facility.GET("/delete-impacts", handlers.DeleteImpact.GetDeleteImpacts)
	facility.GET("/reference-data/stream", handlers.ReferenceData.StreamFacilityReferenceData)
	facility.GET("/jobs", handlers.FacilityJob.ListJobs)
	facility.GET("/jobs/:id", handlers.FacilityJob.GetJob)
	facility.POST("/jobs/:id/retry", handlers.FacilityJob.RetryJob)
	facility.GET("/jobs/:id/download", handlers.Export.DownloadExport)
}

func registerRoutes(group *gin.RouterGroup, authChecker middleware.AuthorizationChecker, routes []routeDefinition, broadcaster FacilityMutationBroadcaster) {
	for _, route := range routes {
		handler := route.Handler
		if change, ok := facilityMutationForRoute(route); ok && broadcaster != nil {
			handler = withFacilityChangeBroadcast(handler, broadcaster, change)
		}
		group.Handle(
			route.Method,
			route.Path,
			middleware.RequirePermission(authChecker, route.Permission),
			handler,
		)
	}
}

func routeDefinitions(handlers *Handlers) []routeDefinition {
	routes := make([]routeDefinition, 0, 96)
	routes = append(routes, hierarchy.Routes(hierarchyHandlers(handlers))...)
	routes = append(routes, reference.Routes(referenceHandlers(handlers))...)
	routes = append(routes, fielddevice.Routes(fieldDeviceHandlers(handlers))...)
	routes = append(routes, objectdata.Routes(objectDataHandlers(handlers))...)
	routes = append(routes, alarm.Routes(alarmHandlers(handlers))...)
	return routes
}
