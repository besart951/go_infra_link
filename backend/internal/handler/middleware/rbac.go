package middleware

import (
	"net/http"

	domainTeam "github.com/besart951/go_infra_link/backend/internal/domain/team"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	rbacsvc "github.com/besart951/go_infra_link/backend/internal/service/rbac"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequireGlobalRole(rbac *rbacsvc.Service, min domainUser.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := GetUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		role, err := rbac.GetGlobalRole(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "authorization_failed"})
			c.Abort()
			return
		}
		if rbacsvc.GlobalRoleLevel(role) < rbacsvc.GlobalRoleLevel(min) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequireTeamRole(rbac *rbacsvc.Service, teamIDParam string, min domainTeam.MemberRole) gin.HandlerFunc {
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

		globalRole, err := rbac.GetGlobalRole(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "authorization_failed"})
			c.Abort()
			return
		}
		if rbacsvc.IsGlobalAdmin(globalRole) {
			c.Next()
			return
		}

		teamID, err := uuid.Parse(c.Param(teamIDParam))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_team_id"})
			c.Abort()
			return
		}

		role, err := rbac.GetTeamRole(teamID, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "authorization_failed"})
			c.Abort()
			return
		}
		if role == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}
		if rbacsvc.TeamRoleLevel(*role) < rbacsvc.TeamRoleLevel(min) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}
		c.Next()
	}
}
