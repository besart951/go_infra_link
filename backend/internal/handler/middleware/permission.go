package middleware

import (
	"net/http"

	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/gin-gonic/gin"
)

// RequirePermission creates middleware that checks if the current user has a specific permission
func RequirePermission(authz AuthorizationChecker, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := GetUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		role, err := authz.GetGlobalRole(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "authorization_failed"})
			c.Abort()
			return
		}

		// Check if role has the required permission
		// For now, we'll use a simple mapping - this can be extended to use the Permission entities
		if !hasPermission(role, permission) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireCanManageRole creates middleware that checks if the current user can manage a specific role
func RequireCanManageRole(authz AuthorizationChecker, targetRole domainUser.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := GetUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		role, err := authz.GetGlobalRole(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "authorization_failed"})
			c.Abort()
			return
		}

		if !domainUser.CanManageRole(role, targetRole) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "You cannot manage users with this role"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// hasPermission checks if a role has a specific permission
// This is a simplified implementation - in a full system, this would query the database
func hasPermission(role domainUser.Role, permission string) bool {
	// Define permission mappings
	permissions := map[domainUser.Role][]string{
		domainUser.RoleSuperAdmin: {
			"user.create", "user.read", "user.update", "user.delete",
			"team.create", "team.read", "team.update", "team.delete",
			"project.create", "project.read", "project.update", "project.delete",
			"system.configure",
		},
		domainUser.RoleAdminFZAG: {
			"user.create", "user.read", "user.update", "user.delete",
			"team.create", "team.read", "team.update", "team.delete",
			"project.create", "project.read", "project.update", "project.delete",
		},
		domainUser.RoleFZAG: {
			"user.create", "user.read", "user.update",
			"team.read", "team.update",
			"project.create", "project.read", "project.update", "project.delete",
		},
		domainUser.RoleAdminPlaner: {
			"user.create", "user.read", "user.update",
			"team.read",
			"project.create", "project.read", "project.update",
		},
		domainUser.RolePlaner: {
			"user.read",
			"team.read",
			"project.read", "project.update",
		},
		domainUser.RoleAdminEnterpreneur: {
			"user.create", "user.read",
			"team.read",
			"project.read",
		},
		domainUser.RoleEnterpreneur: {
			"team.read",
			"project.read",
		},
	}

	rolePerms, ok := permissions[role]
	if !ok {
		return false
	}

	for _, perm := range rolePerms {
		if perm == permission {
			return true
		}
	}

	return false
}
