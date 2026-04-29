package middleware

import (
	"net/http"

	domainTeam "github.com/besart951/go_infra_link/backend/internal/domain/team"
	"github.com/besart951/go_infra_link/backend/internal/requestutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequirePermission(authz AuthorizationChecker, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := GetUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		ctx := c.Request.Context()
		role, err := authz.GetGlobalRole(ctx, userID)
		if err != nil {
			if requestutil.ShouldSuppressErrorResponse(ctx, err) {
				c.Abort()
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": "authorization_failed"})
			c.Abort()
			return
		}

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
