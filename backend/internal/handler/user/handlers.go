package user

type Handlers struct {
	User       *UserHandler
	Admin      *AdminHandler
	Role       *RoleHandler
	Permission *PermissionHandler
}

func NewHandlers(userService UserService, adminService AdminService, roleService RoleQueryService, directoryService UserDirectoryService, permissionService PermissionService, rolePermissionService RolePermissionService) *Handlers {
	return &Handlers{
		User:       NewUserHandler(userService, roleService, directoryService),
		Admin:      NewAdminHandler(adminService),
		Role:       NewRoleHandler(rolePermissionService),
		Permission: NewPermissionHandler(permissionService),
	}
}
