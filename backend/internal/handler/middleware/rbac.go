package middleware

import (
	"net/http"
	"strings"

	domainTeam "github.com/besart951/go_infra_link/backend/internal/domain/team"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/requestutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequireAnyRole(authz AuthorizationChecker, roles ...domainUser.Role) gin.HandlerFunc {
	allowed := make(map[domainUser.Role]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		role, ok := requireGlobalRole(c, authz)
		if !ok {
			return
		}
		if _, allowed := allowed[role]; !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequireSuperAdminForRoleParam(authz AuthorizationChecker, roleParam string) gin.HandlerFunc {
	if roleParam == "" {
		roleParam = "role"
	}

	return func(c *gin.Context) {
		if domainUser.Role(c.Param(roleParam)) != domainUser.RoleSuperAdmin {
			c.Next()
			return
		}

		role, ok := requireGlobalRole(c, authz)
		if !ok {
			return
		}
		if role != domainUser.RoleSuperAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequirePermission(authz AuthorizationChecker, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := requireGlobalRole(c, authz)
		if !ok {
			return
		}

		ctx := c.Request.Context()
		hasPermission, err := authz.HasPermission(ctx, role, permission)
		if err != nil {
			if requestutil.ShouldSuppressErrorResponse(ctx, err) {
				c.Abort()
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": "authorization_failed"})
			c.Abort()
			return
		}
		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequirePermissionWhenQueryTrue(authz AuthorizationChecker, queryParam string, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isTruthyQuery(c.Query(queryParam)) {
			c.Next()
			return
		}

		role, ok := requireGlobalRole(c, authz)
		if !ok {
			return
		}

		ctx := c.Request.Context()
		hasPermission, err := authz.HasPermission(ctx, role, permission)
		if err != nil {
			if requestutil.ShouldSuppressErrorResponse(ctx, err) {
				c.Abort()
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": "authorization_failed"})
			c.Abort()
			return
		}
		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func isTruthyQuery(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func requireGlobalRole(c *gin.Context, authz AuthorizationChecker) (domainUser.Role, bool) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		c.Abort()
		return "", false
	}

	ctx := c.Request.Context()
	role, err := authz.GetGlobalRole(ctx, userID)
	if err != nil {
		if requestutil.ShouldSuppressErrorResponse(ctx, err) {
			c.Abort()
			return "", false
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "authorization_failed"})
		c.Abort()
		return "", false
	}

	return role, true
}

func RequireTeamPermission(authz AuthorizationChecker, teamIDParam string, permission string) gin.HandlerFunc {
	if teamIDParam == "" {
		teamIDParam = "id"
	}
	return func(c *gin.Context) {
		userID, ok := GetUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		ctx := c.Request.Context()
		if permission != "" {
			globalRole, err := authz.GetGlobalRole(ctx, userID)
			if err != nil {
				if requestutil.ShouldSuppressErrorResponse(ctx, err) {
					c.Abort()
					return
				}

				c.JSON(http.StatusInternalServerError, gin.H{"error": "authorization_failed"})
				c.Abort()
				return
			}

			hasPermission, err := authz.HasPermission(ctx, globalRole, permission)
			if err != nil {
				if requestutil.ShouldSuppressErrorResponse(ctx, err) {
					c.Abort()
					return
				}

				c.JSON(http.StatusInternalServerError, gin.H{"error": "authorization_failed"})
				c.Abort()
				return
			}
			if hasPermission {
				c.Next()
				return
			}
		}

		teamID, err := uuid.Parse(c.Param(teamIDParam))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_team_id"})
			c.Abort()
			return
		}

		role, err := authz.GetTeamRole(ctx, teamID, userID)
		if err != nil {
			if requestutil.ShouldSuppressErrorResponse(ctx, err) {
				c.Abort()
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": "authorization_failed"})
			c.Abort()
			return
		}
		if role == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}
		if permission == "" || !domainTeam.HasPermission(*role, permission) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}
		c.Next()
	}
}
