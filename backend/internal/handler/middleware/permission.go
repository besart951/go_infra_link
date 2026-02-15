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

		hasPerm, err := authz.HasPermission(role, permission)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "authorization_failed"})
			c.Abort()
			return
		}

		if !hasPerm {
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
