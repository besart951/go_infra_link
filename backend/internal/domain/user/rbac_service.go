package user

// RBACService defines the interface for role-based access control operations
type RBACService interface {
	// CanManageRole checks if a user with requesterRole can manage/create a user with targetRole
	CanManageRole(requesterRole Role, targetRole Role) bool

	// GetAllowedRoles returns the list of roles that a user with the given role can assign to others
	GetAllowedRoles(requesterRole Role) []Role

	// GetRoleLevel returns the hierarchical level of a role (higher = more privileged)
	GetRoleLevel(role Role) int
}

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

// IsAdmin returns true if the role has at least admin_fzag level privileges.
func IsAdmin(role Role) bool {
	return RoleLevel(role) >= RoleLevel(RoleAdminFZAG)
}

// CanManageRole checks if a user with requesterRole can manage/create a user with targetRole.
func CanManageRole(requesterRole, targetRole Role) bool {
	if requesterRole == RoleEnterpreneur {
		return false
	}
	return RoleLevel(requesterRole) > RoleLevel(targetRole)
}

// GetAllowedRoles returns the list of roles that a user with the given role can assign to others.
func GetAllowedRoles(requesterRole Role) []Role {
	switch requesterRole {
	case RoleSuperAdmin:
		return []Role{
			RoleSuperAdmin,
			RoleAdminFZAG,
			RoleFZAG,
			RoleAdminPlaner,
			RolePlaner,
			RoleAdminEnterpreneur,
			RoleEnterpreneur,
		}
	case RoleAdminFZAG:
		return []Role{
			RoleFZAG,
			RoleAdminPlaner,
			RolePlaner,
			RoleAdminEnterpreneur,
			RoleEnterpreneur,
		}
	case RoleFZAG:
		return []Role{
			RoleAdminPlaner,
			RolePlaner,
			RoleAdminEnterpreneur,
			RoleEnterpreneur,
		}
	case RoleAdminPlaner:
		return []Role{
			RolePlaner,
			RoleAdminEnterpreneur,
			RoleEnterpreneur,
		}
	case RolePlaner:
		return []Role{
			RoleAdminEnterpreneur,
			RoleEnterpreneur,
		}
	case RoleAdminEnterpreneur:
		return []Role{
			RoleEnterpreneur,
		}
	case RoleEnterpreneur:
		return []Role{}
	default:
		return []Role{}
	}
}
