package team

import (
	domainTeam "github.com/besart951/go_infra_link/backend/internal/domain/team"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(protectedV1 *gin.RouterGroup, handler *TeamHandler, authChecker middleware.AuthorizationChecker) {
	teams := protectedV1.Group("/teams")
	{
		teams.POST("", middleware.RequirePermission(authChecker, domainUser.PermissionTeamCreate), handler.CreateTeam)
		teams.GET("", middleware.RequirePermission(authChecker, domainUser.PermissionTeamRead), handler.ListTeams)

		teams.GET("/:id", middleware.RequireTeamPermission(authChecker, "id", domainTeam.PermissionTeamView), handler.GetTeam)
		teams.PUT("/:id", middleware.RequireTeamPermission(authChecker, "id", domainTeam.PermissionTeamEdit), handler.UpdateTeam)
		teams.DELETE("/:id", middleware.RequireTeamPermission(authChecker, "id", domainTeam.PermissionTeamDelete), handler.DeleteTeam)

		teams.POST("/:id/members", middleware.RequireTeamPermission(authChecker, "id", domainTeam.PermissionTeamMemberAdd), handler.AddMember)
		teams.GET("/:id/members", middleware.RequireTeamPermission(authChecker, "id", domainTeam.PermissionTeamMemberList), handler.ListMembers)
		teams.DELETE("/:id/members/:userId", middleware.RequireTeamPermission(authChecker, "id", domainTeam.PermissionTeamMemberRemove), handler.RemoveMember)
	}
}
