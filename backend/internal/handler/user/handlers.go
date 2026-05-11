package user

type Handlers struct {
	User         *UserHandler
	Registration *RegistrationHandler
	Admin        *AdminHandler
	Role         *RoleHandler
	Permission   *PermissionHandler
}

type Dependencies struct {
	Users                 UserService
	Admin                 AdminService
	RoleService           RoleQueryService
	DirectoryService      UserDirectoryService
	RegistrationService   UserRegistrationService
	PermissionService     PermissionService
	RolePermissionService RolePermissionService
}

func NewHandlers(deps Dependencies) *Handlers {
	return &Handlers{
		User:         NewUserHandler(deps.Users, deps.RoleService, deps.DirectoryService, deps.RegistrationService),
		Registration: NewRegistrationHandler(deps.RegistrationService),
		Admin:        NewAdminHandler(deps.Admin, deps.Users),
		Role:         NewRoleHandler(deps.RolePermissionService),
		Permission:   NewPermissionHandler(deps.PermissionService),
	}
}
