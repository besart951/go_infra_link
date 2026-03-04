package dashboard

import (
	"sort"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainTeam "github.com/besart951/go_infra_link/backend/internal/domain/team"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/google/uuid"
)

const onlineWindow = 15 * time.Minute

type Service struct {
	projectRepo    domainProject.ProjectRepository
	phaseRepo      domainProject.PhaseRepository
	teamRepo       domainTeam.TeamRepository
	teamMemberRepo domainTeam.TeamMemberRepository
	userRepo       domainUser.UserRepository
	now            func() time.Time
}

func New(
	projectRepo domainProject.ProjectRepository,
	phaseRepo domainProject.PhaseRepository,
	teamRepo domainTeam.TeamRepository,
	teamMemberRepo domainTeam.TeamMemberRepository,
	userRepo domainUser.UserRepository,
) *Service {
	return &Service{
		projectRepo:    projectRepo,
		phaseRepo:      phaseRepo,
		teamRepo:       teamRepo,
		teamMemberRepo: teamMemberRepo,
		userRepo:       userRepo,
		now:            time.Now,
	}
}

func (s *Service) GetUserDashboard(userID uuid.UUID) (*dto.DashboardResponse, error) {
	now := s.now().UTC()
	response := &dto.DashboardResponse{
		Teams:       make([]dto.DashboardTeamSummaryResponse, 0),
		OnlineUsers: make([]dto.DashboardUserPresenceResponse, 0),
	}

	projects, err := s.projectRepo.GetPaginatedListForUser(domain.PaginationParams{Page: 1, Limit: 1}, userID)
	if err != nil {
		return nil, err
	}

	presenceUsers := make(map[uuid.UUID]domainUser.User)

	if len(projects.Items) > 0 {
		lastProject := projects.Items[0]
		phaseName := "Unknown"
		phase, err := domain.GetByID(s.phaseRepo, lastProject.PhaseID)
		if err == nil {
			phaseName = phase.Name
		}

		response.LastProject = &dto.DashboardProjectResponse{
			ID:        lastProject.ID,
			Name:      lastProject.Name,
			Status:    lastProject.Status,
			Phase:     phaseName,
			UpdatedAt: lastProject.UpdatedAt,
		}

		projectUsers, err := s.projectRepo.ListUsers(lastProject.ID)
		if err != nil {
			return nil, err
		}
		for _, usr := range projectUsers {
			presenceUsers[usr.ID] = usr
		}
	}

	memberships, err := s.teamMemberRepo.ListByUser(userID, domain.PaginationParams{Page: 1, Limit: 20})
	if err != nil {
		return nil, err
	}

	teamIDs := make([]uuid.UUID, 0, len(memberships.Items))
	for _, member := range memberships.Items {
		teamIDs = append(teamIDs, member.TeamID)
	}

	teamsByID := make(map[uuid.UUID]domainTeam.Team)
	if len(teamIDs) > 0 {
		teams, err := s.teamRepo.GetByIds(teamIDs)
		if err != nil {
			return nil, err
		}
		for _, t := range teams {
			teamsByID[t.ID] = *t
		}
	}

	for _, member := range memberships.Items {
		name := "Team"
		if t, ok := teamsByID[member.TeamID]; ok {
			name = t.Name
		}
		response.Teams = append(response.Teams, dto.DashboardTeamSummaryResponse{
			ID:       member.TeamID,
			Name:     name,
			Role:     string(member.Role),
			JoinedAt: member.JoinedAt,
		})
	}

	if len(memberships.Items) > 0 {
		primaryMembership := memberships.Items[0]
		primaryMembers, err := s.teamMemberRepo.ListByTeam(primaryMembership.TeamID, domain.PaginationParams{Page: 1, Limit: 200})
		if err != nil {
			return nil, err
		}

		memberUserIDs := make([]uuid.UUID, 0, len(primaryMembers.Items))
		memberRoleByUser := make(map[uuid.UUID]string)
		for _, member := range primaryMembers.Items {
			memberUserIDs = append(memberUserIDs, member.UserID)
			memberRoleByUser[member.UserID] = string(member.Role)
		}

		memberUsers, err := s.userRepo.GetByIds(memberUserIDs)
		if err != nil {
			return nil, err
		}

		teamName := "Team"
		if t, ok := teamsByID[primaryMembership.TeamID]; ok {
			teamName = t.Name
		}

		teamMembers := make([]dto.DashboardTeamMemberResponse, 0, len(memberUsers))
		for _, usr := range memberUsers {
			presenceUsers[usr.ID] = *usr
			teamMembers = append(teamMembers, dto.DashboardTeamMemberResponse{
				UserID:      usr.ID,
				FirstName:   usr.FirstName,
				LastName:    usr.LastName,
				Email:       usr.Email,
				Role:        memberRoleByUser[usr.ID],
				LastLoginAt: usr.LastLoginAt,
				IsOnline:    isUserOnline(*usr, now),
			})
		}

		sort.Slice(teamMembers, func(i, j int) bool {
			if teamMembers[i].IsOnline != teamMembers[j].IsOnline {
				return teamMembers[i].IsOnline
			}
			return teamMembers[i].FirstName+teamMembers[i].LastName < teamMembers[j].FirstName+teamMembers[j].LastName
		})

		response.PrimaryTeam = &dto.DashboardTeamResponse{
			ID:      primaryMembership.TeamID,
			Name:    teamName,
			Role:    string(primaryMembership.Role),
			Members: teamMembers,
		}
	}

	response.OnlineUsers = buildOnlineUsers(presenceUsers, now)

	return response, nil
}

func buildOnlineUsers(users map[uuid.UUID]domainUser.User, now time.Time) []dto.DashboardUserPresenceResponse {
	online := make([]dto.DashboardUserPresenceResponse, 0)
	for _, usr := range users {
		if !isUserOnline(usr, now) {
			continue
		}
		online = append(online, dto.DashboardUserPresenceResponse{
			ID:          usr.ID,
			FirstName:   usr.FirstName,
			LastName:    usr.LastName,
			Email:       usr.Email,
			LastLoginAt: usr.LastLoginAt,
			IsOnline:    true,
		})
	}

	sort.Slice(online, func(i, j int) bool {
		left := online[i].LastLoginAt
		right := online[j].LastLoginAt
		if left == nil {
			return false
		}
		if right == nil {
			return true
		}
		if left.Equal(*right) {
			return online[i].FirstName+online[i].LastName < online[j].FirstName+online[j].LastName
		}
		return left.After(*right)
	})

	return online
}

func isUserOnline(usr domainUser.User, now time.Time) bool {
	if !usr.IsActive || usr.DisabledAt != nil {
		return false
	}
	if usr.LockedUntil != nil && usr.LockedUntil.After(now) {
		return false
	}
	if usr.LastLoginAt == nil {
		return false
	}
	return now.Sub(*usr.LastLoginAt) <= onlineWindow
}
