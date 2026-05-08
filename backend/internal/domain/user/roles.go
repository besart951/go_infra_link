package user

// AllRoles returns the ordered list of supported roles (highest to lowest).
func AllRoles() []Role {
	return []Role{
		RoleSuperAdmin,
		RoleAdminFZAG,
		RoleFZAG,
		RoleAdminPlaner,
		RolePlaner,
		RoleAdminEnterpreneur,
		RoleEnterpreneur,
	}
}

// AssignableRoles returns roles that requesterRole may assign by hierarchy.
// Superadmin may assign every role; other roles may assign only roles below them.
func AssignableRoles(requesterRole Role) []Role {
	roles := AllRoles()
	if requesterRole == RoleSuperAdmin {
		return append([]Role{}, roles...)
	}

	requesterLevel := RoleLevel(requesterRole)
	if requesterLevel == 0 {
		return []Role{}
	}

	assignable := make([]Role, 0, len(roles))
	for _, role := range roles {
		if RoleLevel(role) < requesterLevel {
			assignable = append(assignable, role)
		}
	}
	return assignable
}

// IsValidRole reports whether the role is a known role constant.
func IsValidRole(role Role) bool {
	switch role {
	case RoleSuperAdmin,
		RoleAdminFZAG,
		RoleFZAG,
		RoleAdminPlaner,
		RolePlaner,
		RoleAdminEnterpreneur,
		RoleEnterpreneur:
		return true
	default:
		return false
	}
}

// RoleDisplayName returns a human-friendly label for a role.
func RoleDisplayName(role Role) string {
	switch role {
	case RoleSuperAdmin:
		return "Super Administrator"
	case RoleAdminFZAG:
		return "FZAG Administrator"
	case RoleFZAG:
		return "FZAG"
	case RoleAdminPlaner:
		return "Planner Administrator"
	case RolePlaner:
		return "Planner"
	case RoleAdminEnterpreneur:
		return "Entrepreneur Administrator"
	case RoleEnterpreneur:
		return "Entrepreneur"
	default:
		return string(role)
	}
}

// RoleDescription returns a short description for a role.
func RoleDescription(role Role) string {
	switch role {
	case RoleSuperAdmin:
		return "Full system access with all administrative capabilities"
	case RoleAdminFZAG:
		return "FZAG administrator with user and team management capabilities"
	case RoleFZAG:
		return "FZAG user with project and limited user management"
	case RoleAdminPlaner:
		return "Planner administrator with user and project management"
	case RolePlaner:
		return "Planner with project access and limited management"
	case RoleAdminEnterpreneur:
		return "Entrepreneur administrator with limited user creation"
	case RoleEnterpreneur:
		return "Basic read-only access to teams and projects"
	default:
		return ""
	}
}
