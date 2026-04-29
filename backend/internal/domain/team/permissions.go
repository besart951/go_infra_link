package team

const (
	PermissionTeamView         = "team.read"
	PermissionTeamEdit         = "team.update"
	PermissionTeamDelete       = "team.delete"
	PermissionTeamMemberList   = "team.read"
	PermissionTeamMemberAdd    = "team.update"
	PermissionTeamMemberRemove = "team.update"
)

func PermissionsForMemberRole(role MemberRole) []string {
	switch role {
	case MemberRoleOwner:
		return []string{
			PermissionTeamView,
			PermissionTeamEdit,
			PermissionTeamDelete,
		}
	case MemberRoleManager:
		return []string{
			PermissionTeamView,
			PermissionTeamEdit,
		}
	case MemberRoleMember:
		return []string{
			PermissionTeamView,
		}
	default:
		return []string{}
	}
}

func HasPermission(role MemberRole, permission string) bool {
	for _, granted := range PermissionsForMemberRole(role) {
		if granted == permission {
			return true
		}
	}
	return false
}