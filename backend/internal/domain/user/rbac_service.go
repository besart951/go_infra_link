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
