package team

// RoleLevel returns the hierarchical level of a team member role.
func RoleLevel(r MemberRole) int {
	switch r {
	case MemberRoleOwner:
		return 100
	case MemberRoleManager:
		return 50
	case MemberRoleMember:
		return 10
	default:
		return 10
	}
}
