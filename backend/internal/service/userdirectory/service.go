package userdirectory

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainTeam "github.com/besart951/go_infra_link/backend/internal/domain/team"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

var allowedOrderBy = map[string]func(a, b Item) bool{
	"first_name": func(a, b Item) bool { return strings.ToLower(a.User.FirstName) < strings.ToLower(b.User.FirstName) },
	"last_name":  func(a, b Item) bool { return strings.ToLower(a.User.LastName) < strings.ToLower(b.User.LastName) },
	"email":      func(a, b Item) bool { return strings.ToLower(a.User.Email) < strings.ToLower(b.User.Email) },
	"role":       func(a, b Item) bool { return domainUser.RoleLevel(a.User.Role) < domainUser.RoleLevel(b.User.Role) },
	"created_at": func(a, b Item) bool { return a.User.CreatedAt.Before(b.User.CreatedAt) },
	"last_login_at": func(a, b Item) bool {
		if a.User.LastLoginAt == nil {
			return false
		}
		if b.User.LastLoginAt == nil {
			return true
		}
		return a.User.LastLoginAt.Before(*b.User.LastLoginAt)
	},
}

type TeamView struct {
	ID   uuid.UUID
	Name string
}

type TeamFilter struct {
	ID    uuid.UUID
	Name  string
	Count int64
}

type Capabilities struct {
	CanUpdate     bool
	CanDelete     bool
	CanDisable    bool
	CanEnable     bool
	CanChangeRole bool
}

type PageCapabilities struct {
	CanCreateUser bool
}

type Item struct {
	User         domainUser.User
	Teams        []TeamView
	Capabilities Capabilities
}

type ListResult struct {
	Items            []Item
	Total            int64
	Page             int
	TotalPages       int
	Teams            []TeamFilter
	PageCapabilities PageCapabilities
}

type UserReader interface {
	GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainUser.User], error)
	GetByIds(ctx context.Context, ids []uuid.UUID) ([]*domainUser.User, error)
}

type TeamReader interface {
	GetByIds(ctx context.Context, ids []uuid.UUID) ([]*domainTeam.Team, error)
}

type TeamMembershipReader interface {
	ListByUser(ctx context.Context, userID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainTeam.TeamMember], error)
}

type Service struct {
	users           UserReader
	teams           TeamReader
	memberships     TeamMembershipReader
	rolePermissions RolePermissionReader
}

type RolePermissionReader interface {
	ListByRole(ctx context.Context, role domainUser.Role) ([]domainUser.RolePermission, error)
}

type permissionSet map[string]struct{}

const superAdminPermissionWildcard = "*"

func New(users UserReader, teams TeamReader, memberships TeamMembershipReader, rolePermissions RolePermissionReader) *Service {
	return &Service{users: users, teams: teams, memberships: memberships, rolePermissions: rolePermissions}
}

func (s *Service) List(ctx context.Context, requesterID uuid.UUID, page, limit int, search, teamID, orderBy, order string) (*ListResult, error) {
	requester, err := domain.GetByID(ctx, s.users, requesterID)
	if err != nil {
		return nil, err
	}

	requesterRolePerms, err := s.rolePermissions.ListByRole(ctx, requester.Role)
	if err != nil {
		return nil, err
	}
	requesterPermissions := permissionSetForRole(requester.Role, requesterRolePerms)
	if !requesterPermissions.has(domainUser.PermissionUserRead) {
		return nil, domainUser.ErrForbiddenUserDirectory
	}

	requesterTeams, teamNames, err := s.loadUserTeams(ctx, requesterID)
	if err != nil {
		return nil, err
	}

	allUsers, err := s.loadAllUsers(ctx)
	if err != nil {
		return nil, err
	}

	visible := make([]Item, 0, len(allUsers))
	teamCounts := map[uuid.UUID]int64{}
	requestedTeamID := uuid.Nil
	if strings.TrimSpace(teamID) != "" {
		parsed, parseErr := uuid.Parse(teamID)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid team_id: %w", parseErr)
		}
		requestedTeamID = parsed
	}

	rolePermissionCache := map[domainUser.Role]permissionSet{
		requester.Role: requesterPermissions,
	}
	resolveRolePermissions := func(role domainUser.Role) (permissionSet, error) {
		if cached, ok := rolePermissionCache[role]; ok {
			return cached, nil
		}
		permissions, err := s.rolePermissions.ListByRole(ctx, role)
		if err != nil {
			return nil, err
		}
		resolved := permissionSetForRole(role, permissions)
		rolePermissionCache[role] = resolved
		return resolved, nil
	}

	canCreateUser := requesterPermissions.has(domainUser.PermissionUserCreate)

	for _, candidate := range allUsers {
		candidateTeams, candidateTeamNames, err := s.loadUserTeams(ctx, candidate.ID)
		if err != nil {
			return nil, err
		}

		candidatePermissions, err := resolveRolePermissions(candidate.Role)
		if err != nil {
			return nil, err
		}

		visibleTeamIDs := intersectVisibleTeamIDs(requesterPermissions.has(domainUser.PermissionTeamRead), requesterTeams, candidateTeams)
		if !canSeeUser(requester.ID, requesterPermissions, candidate, visibleTeamIDs, candidatePermissions) {
			continue
		}

		if requestedTeamID != uuid.Nil {
			if _, ok := visibleTeamIDs[requestedTeamID]; !ok {
				continue
			}
		}

		if !matchesSearch(candidate, search) {
			continue
		}

		itemTeams := make([]TeamView, 0, len(visibleTeamIDs))
		for id := range visibleTeamIDs {
			if name, ok := teamNames[id]; ok {
				itemTeams = append(itemTeams, TeamView{ID: id, Name: name})
				teamCounts[id]++
				continue
			}
			if name, ok := candidateTeamNames[id]; ok {
				itemTeams = append(itemTeams, TeamView{ID: id, Name: name})
				teamCounts[id]++
			}
		}
		sort.Slice(itemTeams, func(i, j int) bool { return strings.ToLower(itemTeams[i].Name) < strings.ToLower(itemTeams[j].Name) })

		visible = append(visible, Item{
			User:         *candidate,
			Teams:        itemTeams,
			Capabilities: buildCapabilities(requester.ID, requesterPermissions, *candidate, candidatePermissions, len(allUsers)),
		})
	}

	sortVisible(visible, orderBy, order)

	page, limit = domain.NormalizePagination(page, limit, 10)
	total := int64(len(visible))
	offset := (page - 1) * limit
	offset = min(offset, len(visible))
	end := offset + limit
	end = min(end, len(visible))

	filters := make([]TeamFilter, 0, len(teamCounts))
	for id, count := range teamCounts {
		name := teamNames[id]
		if name == "" {
			continue
		}
		filters = append(filters, TeamFilter{ID: id, Name: name, Count: count})
	}
	sort.Slice(filters, func(i, j int) bool { return strings.ToLower(filters[i].Name) < strings.ToLower(filters[j].Name) })

	return &ListResult{
		Items:            visible[offset:end],
		Total:            total,
		Page:             page,
		TotalPages:       domain.CalculateTotalPages(total, limit),
		Teams:            filters,
		PageCapabilities: PageCapabilities{CanCreateUser: canCreateUser},
	}, nil
}

func buildCapabilities(requesterID uuid.UUID, requesterPermissions permissionSet, target domainUser.User, targetPermissions permissionSet, totalUsers int) Capabilities {
	if requesterID == target.ID {
		return Capabilities{}
	}
	if !permissionSetContainsAll(requesterPermissions, targetPermissions) {
		return Capabilities{}
	}

	canMutateSuperAdmin := !(target.Role == domainUser.RoleSuperAdmin && totalUsers <= 1)
	return Capabilities{
		CanUpdate:     requesterPermissions.has(domainUser.PermissionUserUpdate),
		CanDelete:     requesterPermissions.has(domainUser.PermissionUserDelete) && canMutateSuperAdmin,
		CanDisable:    requesterPermissions.has(domainUser.PermissionUserUpdate) && target.IsActive && canMutateSuperAdmin,
		CanEnable:     requesterPermissions.has(domainUser.PermissionUserUpdate) && !target.IsActive,
		CanChangeRole: requesterPermissions.has(domainUser.PermissionUserUpdate),
	}
}

func canSeeUser(requesterID uuid.UUID, requesterPermissions permissionSet, candidate *domainUser.User, visibleTeamIDs map[uuid.UUID]struct{}, candidatePermissions permissionSet) bool {
	if requesterID == candidate.ID {
		return true
	}
	if permissionSetContainsAll(requesterPermissions, candidatePermissions) {
		return true
	}
	return len(visibleTeamIDs) > 0
}

func matchesSearch(candidate *domainUser.User, search string) bool {
	search = strings.TrimSpace(strings.ToLower(search))
	if search == "" {
		return true
	}
	fullName := strings.ToLower(strings.TrimSpace(candidate.FirstName + " " + candidate.LastName))
	return strings.Contains(strings.ToLower(candidate.FirstName), search) ||
		strings.Contains(strings.ToLower(candidate.LastName), search) ||
		strings.Contains(strings.ToLower(candidate.Email), search) ||
		strings.Contains(fullName, search)
}

func sortVisible(items []Item, orderBy, order string) {
	less, ok := allowedOrderBy[orderBy]
	if !ok {
		less = allowedOrderBy["last_login_at"]
	}
	desc := !strings.EqualFold(order, "asc")
	sort.SliceStable(items, func(i, j int) bool {
		if desc {
			return less(items[j], items[i])
		}
		return less(items[i], items[j])
	})
}

func intersectVisibleTeamIDs(canReadAllTeams bool, requesterTeams map[uuid.UUID]struct{}, candidateTeams map[uuid.UUID]struct{}) map[uuid.UUID]struct{} {
	if canReadAllTeams {
		return candidateTeams
	}
	result := map[uuid.UUID]struct{}{}
	for id := range candidateTeams {
		if _, ok := requesterTeams[id]; ok {
			result[id] = struct{}{}
		}
	}
	return result
}

func permissionSetFromRolePermissions(permissions []domainUser.RolePermission) permissionSet {
	set := permissionSet{}
	for _, rolePermission := range permissions {
		set[rolePermission.Permission] = struct{}{}
	}
	return set
}

func permissionSetForRole(role domainUser.Role, permissions []domainUser.RolePermission) permissionSet {
	if role == domainUser.RoleSuperAdmin {
		return permissionSet{superAdminPermissionWildcard: {}}
	}
	return permissionSetFromRolePermissions(permissions)
}

func permissionSetContainsAll(granted permissionSet, required permissionSet) bool {
	for permission := range required {
		if !granted.has(permission) {
			return false
		}
	}
	return true
}

func (s permissionSet) has(permission string) bool {
	if _, ok := s[superAdminPermissionWildcard]; ok {
		return true
	}
	_, ok := s[permission]
	return ok
}

func (s *Service) loadAllUsers(ctx context.Context) ([]*domainUser.User, error) {
	result, err := s.users.GetPaginatedList(ctx, domain.PaginationParams{Page: 1, Limit: 10000, OrderBy: "last_login_at", Order: "desc"})
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(result.Items))
	for _, item := range result.Items {
		ids = append(ids, item.ID)
	}
	return s.users.GetByIds(ctx, ids)
}

func (s *Service) loadUserTeams(ctx context.Context, userID uuid.UUID) (map[uuid.UUID]struct{}, map[uuid.UUID]string, error) {
	memberships, err := s.memberships.ListByUser(ctx, userID, domain.PaginationParams{Page: 1, Limit: 1000})
	if err != nil {
		return nil, nil, err
	}
	teamIDs := make([]uuid.UUID, 0, len(memberships.Items))
	teamSet := map[uuid.UUID]struct{}{}
	for _, member := range memberships.Items {
		teamSet[member.TeamID] = struct{}{}
		teamIDs = append(teamIDs, member.TeamID)
	}
	teams, err := s.teams.GetByIds(ctx, teamIDs)
	if err != nil {
		return nil, nil, err
	}
	teamNames := map[uuid.UUID]string{}
	for _, team := range teams {
		teamNames[team.ID] = team.Name
	}
	return teamSet, teamNames, nil
}
