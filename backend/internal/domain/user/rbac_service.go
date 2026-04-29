package user

// RoleLevel returns the hierarchical level of a global role (higher = more privileged).
func RoleLevel(role Role) int {
	switch role {
	case RoleSuperAdmin:
		return 100
	case RoleAdminFZAG:
		return 90
	case RoleFZAG:
		return 80
	case RoleAdminPlaner:
		return 70
	case RolePlaner:
		return 60
	case RoleAdminEnterpreneur:
		return 50
	case RoleEnterpreneur:
		return 40
	default:
		return 0
	}
}
