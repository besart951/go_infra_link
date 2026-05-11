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
		users.GET(
			"/directory",
			middleware.RequirePermissionWhenQueryTrue(authChecker, "include_deleted", domainUser.PermissionUserReadDeleted),
			handlers.User.ListDirectory,
		)
		users.PUT("/me/password", handlers.User.UpdateOwnPassword)
	}

	usersAdmin := protectedV1.Group("/users")
	{
		usersAdmin.POST("", middleware.RequirePermission(authChecker, domainUser.PermissionUserCreate), handlers.User.CreateUser)
		usersAdmin.POST("/invitations", middleware.RequirePermission(authChecker, domainUser.PermissionUserCreate), handlers.Registration.CreateInvitation)
		usersAdmin.GET(
			"",
			middleware.RequirePermission(authChecker, domainUser.PermissionUserRead),
			middleware.RequirePermissionWhenQueryTrue(authChecker, "include_deleted", domainUser.PermissionUserReadDeleted),
			handlers.User.ListUsers,
		)
		usersAdmin.GET("/:id", middleware.RequirePermission(authChecker, domainUser.PermissionUserRead), handlers.User.GetUser)
		usersAdmin.GET("/:id/registration", middleware.RequirePermission(authChecker, domainUser.PermissionUserRead), handlers.Registration.GetProcess)
		usersAdmin.POST("/:id/registration/resend", middleware.RequirePermission(authChecker, domainUser.PermissionUserCreate), handlers.Registration.ResendInvitation)
		usersAdmin.PUT("/:id", middleware.RequirePermission(authChecker, domainUser.PermissionUserUpdate), handlers.User.UpdateUser)
		usersAdmin.DELETE("/:id", middleware.RequirePermission(authChecker, domainUser.PermissionUserDelete), handlers.User.DeleteUser)
	}
}

func RegisterRoleRoutes(protectedV1 *gin.RouterGroup, handlers *Handlers, authChecker middleware.AuthorizationChecker) {
	roleAdmins := middleware.RequireAnyRole(authChecker, domainUser.RoleSuperAdmin, domainUser.RoleAdminFZAG)
	superAdmins := middleware.RequireAnyRole(authChecker, domainUser.RoleSuperAdmin)
	superAdminRoleTarget := middleware.RequireSuperAdminForRoleParam(authChecker, "role")

	roles := protectedV1.Group("/roles", roleAdmins)
	{
		roles.GET("", handlers.Role.ListRoles)
		roles.PUT("/:role/permissions", superAdminRoleTarget, handlers.Role.UpdateRolePermissions)
		roles.POST("/:role/permissions", superAdminRoleTarget, handlers.Role.AddRolePermission)
		roles.DELETE("/:role/permissions/:permission", superAdminRoleTarget, handlers.Role.RemoveRolePermission)
	}

	permissions := protectedV1.Group("/permissions", roleAdmins)
	{
		permissions.GET("", handlers.Permission.ListPermissions)
		permissions.POST("", superAdmins, handlers.Permission.CreatePermission)
		permissions.PUT("/:id", superAdmins, handlers.Permission.UpdatePermission)
		permissions.DELETE("/:id", superAdmins, handlers.Permission.DeletePermission)
	}
}

func RegisterAdminRoutes(protectedV1 *gin.RouterGroup, handlers *Handlers, authChecker middleware.AuthorizationChecker) {
	admin := protectedV1.Group("/admin")
	{
		admin.POST("/users/:id/disable", middleware.RequirePermission(authChecker, domainUser.PermissionUserUpdate), handlers.Admin.DisableUser)
		admin.POST("/users/:id/enable", middleware.RequirePermission(authChecker, domainUser.PermissionUserUpdate), handlers.Admin.EnableUser)
		admin.POST("/users/:id/restore",
			middleware.RequirePermission(authChecker, domainUser.PermissionUserDelete),
			middleware.RequirePermission(authChecker, domainUser.PermissionUserReadDeleted),
			handlers.Admin.RestoreUser,
		)
		admin.POST("/users/:id/role", middleware.RequirePermission(authChecker, domainUser.PermissionUserUpdate), handlers.Admin.SetUserRole)
	}
}
