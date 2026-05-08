package user

type Handlers struct {
	User         *UserHandler
	Registration *RegistrationHandler
	Admin        *AdminHandler
	Role         *RoleHandler
	Permission   *PermissionHandler
}

func NewHandlers(userService UserService, adminService AdminService, roleService RoleQueryService, directoryService UserDirectoryService, registrationService UserRegistrationService, permissionService PermissionService, rolePermissionService RolePermissionService) *Handlers {
	return &Handlers{
		User:         NewUserHandler(userService, roleService, directoryService, registrationService),
		Registration: NewRegistrationHandler(registrationService),
		Admin:        NewAdminHandler(adminService),
		Role:         NewRoleHandler(rolePermissionService),
		Permission:   NewPermissionHandler(permissionService),
	}
}
