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
		teams.POST("", middleware.RequireGlobalRole(authChecker, domainUser.RoleAdminFZAG), handler.CreateTeam)
		teams.GET("", middleware.RequireGlobalRole(authChecker, domainUser.RoleAdminFZAG), handler.ListTeams)

		teams.GET("/:id", middleware.RequireTeamRole(authChecker, "id", domainTeam.MemberRoleMember), handler.GetTeam)
		teams.PUT("/:id", middleware.RequireTeamRole(authChecker, "id", domainTeam.MemberRoleManager), handler.UpdateTeam)
		teams.DELETE("/:id", middleware.RequireTeamRole(authChecker, "id", domainTeam.MemberRoleOwner), handler.DeleteTeam)

		teams.POST("/:id/members", middleware.RequireTeamRole(authChecker, "id", domainTeam.MemberRoleManager), handler.AddMember)
		teams.GET("/:id/members", middleware.RequireTeamRole(authChecker, "id", domainTeam.MemberRoleMember), handler.ListMembers)
		teams.DELETE("/:id/members/:userId", middleware.RequireTeamRole(authChecker, "id", domainTeam.MemberRoleManager), handler.RemoveMember)
	}
}
