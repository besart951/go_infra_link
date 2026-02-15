package user

// RoleInfo provides role metadata along with permissions.
type RoleInfo struct {
	Name        Role
	DisplayName string
	Description string
	Level       int
	Permissions []string
	CanManage   []Role
}
