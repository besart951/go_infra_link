package user

import (
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(protectedV1 *gin.RouterGroup, handlers *Handlers, authChecker middleware.AuthorizationChecker) {
	users := protectedV1.Group("/users")
	{
		users.GET("/allowed-roles", handlers.User.GetAllowedRoles)
	}

	usersAdmin := protectedV1.Group("/users")
	usersAdmin.Use(middleware.RequireGlobalRole(authChecker, domainUser.RoleAdminFZAG))
	{
		usersAdmin.POST("", handlers.User.CreateUser)
		usersAdmin.GET("", handlers.User.ListUsers)
		usersAdmin.GET("/:id", handlers.User.GetUser)
		usersAdmin.PUT("/:id", handlers.User.UpdateUser)
		usersAdmin.DELETE("/:id", handlers.User.DeleteUser)
	}
}

func RegisterRoleRoutes(protectedV1 *gin.RouterGroup, handlers *Handlers, authChecker middleware.AuthorizationChecker) {
	roles := protectedV1.Group("/roles")
	roles.Use(middleware.RequireGlobalRole(authChecker, domainUser.RoleAdminFZAG))
	{
		roles.GET("", handlers.Role.ListRoles)
		roles.PUT("/:role/permissions", handlers.Role.UpdateRolePermissions)
		roles.POST("/:role/permissions", handlers.Role.AddRolePermission)
		roles.DELETE("/:role/permissions/:permission", handlers.Role.RemoveRolePermission)
	}

	permissions := protectedV1.Group("/permissions")
	permissions.Use(middleware.RequireGlobalRole(authChecker, domainUser.RoleAdminFZAG))
	{
		permissions.GET("", handlers.Permission.ListPermissions)
		permissions.POST("", handlers.Permission.CreatePermission)
		permissions.PUT("/:id", handlers.Permission.UpdatePermission)
		permissions.DELETE("/:id", handlers.Permission.DeletePermission)
	}
}

func RegisterAdminRoutes(protectedV1 *gin.RouterGroup, handlers *Handlers, authChecker middleware.AuthorizationChecker) {
	admin := protectedV1.Group("/admin")
	admin.Use(middleware.RequireGlobalRole(authChecker, domainUser.RoleAdminFZAG))
	{
		admin.POST("/users/:id/disable", handlers.Admin.DisableUser)
		admin.POST("/users/:id/enable", handlers.Admin.EnableUser)
		admin.POST("/users/:id/role", handlers.Admin.SetUserRole)
	}
}
