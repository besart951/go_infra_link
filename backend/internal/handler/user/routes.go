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
		users.GET("/directory", handlers.User.ListDirectory)
	}

	usersAdmin := protectedV1.Group("/users")
	{
		usersAdmin.POST("", middleware.RequirePermission(authChecker, domainUser.PermissionUserCreate), handlers.User.CreateUser)
		usersAdmin.GET("", middleware.RequirePermission(authChecker, domainUser.PermissionUserRead), handlers.User.ListUsers)
		usersAdmin.GET("/:id", middleware.RequirePermission(authChecker, domainUser.PermissionUserRead), handlers.User.GetUser)
		usersAdmin.PUT("/:id", middleware.RequirePermission(authChecker, domainUser.PermissionUserUpdate), handlers.User.UpdateUser)
		usersAdmin.DELETE("/:id", middleware.RequirePermission(authChecker, domainUser.PermissionUserDelete), handlers.User.DeleteUser)
	}
}

func RegisterRoleRoutes(protectedV1 *gin.RouterGroup, handlers *Handlers, authChecker middleware.AuthorizationChecker) {
	roles := protectedV1.Group("/roles")
	{
		roles.GET("", middleware.RequirePermission(authChecker, domainUser.PermissionRoleRead), handlers.Role.ListRoles)
		roles.PUT("/:role/permissions", middleware.RequirePermission(authChecker, domainUser.PermissionRoleUpdate), handlers.Role.UpdateRolePermissions)
		roles.POST("/:role/permissions", middleware.RequirePermission(authChecker, domainUser.PermissionRoleUpdate), handlers.Role.AddRolePermission)
		roles.DELETE("/:role/permissions/:permission", middleware.RequirePermission(authChecker, domainUser.PermissionRoleUpdate), handlers.Role.RemoveRolePermission)
	}

	permissions := protectedV1.Group("/permissions")
	{
		permissions.GET("", middleware.RequirePermission(authChecker, domainUser.PermissionPermissionRead), handlers.Permission.ListPermissions)
		permissions.POST("", middleware.RequirePermission(authChecker, domainUser.PermissionPermissionCreate), handlers.Permission.CreatePermission)
		permissions.PUT("/:id", middleware.RequirePermission(authChecker, domainUser.PermissionPermissionUpdate), handlers.Permission.UpdatePermission)
		permissions.DELETE("/:id", middleware.RequirePermission(authChecker, domainUser.PermissionPermissionDelete), handlers.Permission.DeletePermission)
	}
}

func RegisterAdminRoutes(protectedV1 *gin.RouterGroup, handlers *Handlers, authChecker middleware.AuthorizationChecker) {
	admin := protectedV1.Group("/admin")
	{
		admin.POST("/users/:id/disable", middleware.RequirePermission(authChecker, domainUser.PermissionUserUpdate), handlers.Admin.DisableUser)
		admin.POST("/users/:id/enable", middleware.RequirePermission(authChecker, domainUser.PermissionUserUpdate), handlers.Admin.EnableUser)
		admin.POST("/users/:id/role", middleware.RequirePermission(authChecker, domainUser.PermissionUserUpdate), handlers.Admin.SetUserRole)
	}
}
