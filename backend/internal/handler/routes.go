package handler

import (
	facilityhandler "github.com/besart951/go_infra_link/backend/internal/handler/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	authsvc "github.com/besart951/go_infra_link/backend/internal/service/auth"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	ProjectHandler *ProjectHandler
	UserHandler    *UserHandler
	AuthHandler    *AuthHandler

	FacilityBuildingHandler       *facilityhandler.BuildingHandler
	FacilitySystemTypeHandler     *facilityhandler.SystemTypeHandler
	FacilitySystemPartHandler     *facilityhandler.SystemPartHandler
	FacilitySpecificationHandler  *facilityhandler.SpecificationHandler
	FacilityApparatHandler        *facilityhandler.ApparatHandler
	FacilityFieldDeviceHandler    *facilityhandler.FieldDeviceHandler
	FacilityControlCabinetHandler *facilityhandler.ControlCabinetHandler
	FacilitySPSControllerHandler  *facilityhandler.SPSControllerHandler
}

// RegisterRoutes registers all API routes
func RegisterRoutes(r *gin.Engine, handlers *Handlers, jwtService authsvc.JWTService) {
	// Public API v1 group (login only)
	publicV1 := r.Group("/api/v1")
	publicAuth := publicV1.Group("/auth")
	{
		publicAuth.POST("/login", handlers.AuthHandler.Login)
		publicAuth.POST("/dev-login", handlers.AuthHandler.DevLogin)
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
	protectedV1.Use(middleware.AuthGuard(jwtService))
	protectedV1.Use(middleware.CSRFMiddleware())

	// Project routes
	projects := protectedV1.Group("/projects")
	{
		projects.POST("", handlers.ProjectHandler.CreateProject)
		projects.GET("", handlers.ProjectHandler.ListProjects)
		projects.GET("/:id", handlers.ProjectHandler.GetProject)
		projects.PUT("/:id", handlers.ProjectHandler.UpdateProject)
		projects.DELETE("/:id", handlers.ProjectHandler.DeleteProject)
	}

	// User routes
	users := protectedV1.Group("/users")
	{
		users.POST("", handlers.UserHandler.CreateUser)
		users.GET("", handlers.UserHandler.ListUsers)
		users.GET("/:id", handlers.UserHandler.GetUser)
		users.PUT("/:id", handlers.UserHandler.UpdateUser)
		users.DELETE("/:id", handlers.UserHandler.DeleteUser)
	}

	// Auth routes (protected)
	authProtected := protectedV1.Group("/auth")
	{
		authProtected.GET("/me", handlers.AuthHandler.Me)
	}

	// Facility routes
	facility := protectedV1.Group("/facility")
	{
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

		facility.POST("/specifications", handlers.FacilitySpecificationHandler.CreateSpecification)
		facility.GET("/specifications", handlers.FacilitySpecificationHandler.ListSpecifications)
		facility.GET("/specifications/:id", handlers.FacilitySpecificationHandler.GetSpecification)
		facility.PUT("/specifications/:id", handlers.FacilitySpecificationHandler.UpdateSpecification)
		facility.DELETE("/specifications/:id", handlers.FacilitySpecificationHandler.DeleteSpecification)

		facility.POST("/apparats", handlers.FacilityApparatHandler.CreateApparat)
		facility.GET("/apparats", handlers.FacilityApparatHandler.ListApparats)
		facility.GET("/apparats/:id", handlers.FacilityApparatHandler.GetApparat)
		facility.PUT("/apparats/:id", handlers.FacilityApparatHandler.UpdateApparat)
		facility.DELETE("/apparats/:id", handlers.FacilityApparatHandler.DeleteApparat)

		facility.POST("/control-cabinets", handlers.FacilityControlCabinetHandler.CreateControlCabinet)
		facility.GET("/control-cabinets", handlers.FacilityControlCabinetHandler.ListControlCabinets)
		facility.GET("/control-cabinets/:id", handlers.FacilityControlCabinetHandler.GetControlCabinet)
		facility.PUT("/control-cabinets/:id", handlers.FacilityControlCabinetHandler.UpdateControlCabinet)
		facility.DELETE("/control-cabinets/:id", handlers.FacilityControlCabinetHandler.DeleteControlCabinet)

		facility.POST("/field-devices", handlers.FacilityFieldDeviceHandler.CreateFieldDevice)
		facility.GET("/field-devices", handlers.FacilityFieldDeviceHandler.ListFieldDevices)
		facility.GET("/field-devices/:id", handlers.FacilityFieldDeviceHandler.GetFieldDevice)
		facility.PUT("/field-devices/:id", handlers.FacilityFieldDeviceHandler.UpdateFieldDevice)
		facility.DELETE("/field-devices/:id", handlers.FacilityFieldDeviceHandler.DeleteFieldDevice)

		facility.POST("/sps-controllers", handlers.FacilitySPSControllerHandler.CreateSPSController)
		facility.GET("/sps-controllers", handlers.FacilitySPSControllerHandler.ListSPSControllers)
		facility.GET("/sps-controllers/:id", handlers.FacilitySPSControllerHandler.GetSPSController)
		facility.PUT("/sps-controllers/:id", handlers.FacilitySPSControllerHandler.UpdateSPSController)
		facility.DELETE("/sps-controllers/:id", handlers.FacilitySPSControllerHandler.DeleteSPSController)
	}
}
