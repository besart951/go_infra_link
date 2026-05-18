package userdirectory

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainTeam "github.com/besart951/go_infra_link/backend/internal/domain/team"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

var allowedOrderBy = map[string]func(a, b Item) bool{
	"first_name": func(a, b Item) bool { return strings.ToLower(a.User.FirstName) < strings.ToLower(b.User.FirstName) },
	"last_name":  func(a, b Item) bool { return strings.ToLower(a.User.LastName) < strings.ToLower(b.User.LastName) },
	"email": func(a, b Item) bool {
		return strings.ToLower(a.User.EmailValue()) < strings.ToLower(b.User.EmailValue())
	},
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

type RoleFilter struct {
	Role        domainUser.Role
	DisplayName string
	Count       int64
}

type Capabilities struct {
	CanUpdate     bool
	CanDelete     bool
	CanDisable    bool
	CanEnable     bool
	CanRestore    bool
	CanChangeRole bool
}

type PageCapabilities struct {
	CanCreateUser  bool
	CanReadDeleted bool
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
	Roles            []RoleFilter
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

func (s *Service) List(ctx context.Context, requesterID uuid.UUID, page, limit int, search, teamID, role, orderBy, order string, includeDeleted bool) (*ListResult, error) {
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

	allUsers, err := s.loadAllUsers(ctx, includeDeleted)
	if err != nil {
		return nil, err
	}

	requestedTeamID := uuid.Nil
	if strings.TrimSpace(teamID) != "" {
		parsed, parseErr := uuid.Parse(teamID)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid team_id: %w", parseErr)
		}
		requestedTeamID = parsed
	}
	requestedRole := domainUser.Role(strings.TrimSpace(role))
	if requestedRole != "" && !domainUser.IsValidRole(requestedRole) {
		return nil, fmt.Errorf("invalid role: %s", requestedRole)
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

	baseVisible := make([]Item, 0, len(allUsers))
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

		if !matchesSearch(candidate, search) {
			continue
		}

		itemTeams := make([]TeamView, 0, len(visibleTeamIDs))
		for id := range visibleTeamIDs {
			if name, ok := teamNames[id]; ok {
				itemTeams = append(itemTeams, TeamView{ID: id, Name: name})
				continue
			}
			if name, ok := candidateTeamNames[id]; ok {
				itemTeams = append(itemTeams, TeamView{ID: id, Name: name})
			}
		}
		sort.Slice(itemTeams, func(i, j int) bool { return strings.ToLower(itemTeams[i].Name) < strings.ToLower(itemTeams[j].Name) })

		baseVisible = append(baseVisible, Item{
			User:         *candidate,
			Teams:        itemTeams,
			Capabilities: buildCapabilities(requester.ID, requester.Role, requesterPermissions, *candidate, candidatePermissions, len(allUsers)),
		})
	}

	visible := make([]Item, 0, len(baseVisible))
	teamCounts := map[uuid.UUID]int64{}
	teamFilterNames := map[uuid.UUID]string{}
	roleCounts := map[domainUser.Role]int64{}
	for _, item := range baseVisible {
		matchesTeam := requestedTeamID == uuid.Nil || itemHasTeam(item, requestedTeamID)
		matchesRole := requestedRole == "" || item.User.Role == requestedRole

		if matchesRole {
			for _, team := range item.Teams {
				teamCounts[team.ID]++
				teamFilterNames[team.ID] = team.Name
			}
		}
		if matchesTeam {
			roleCounts[item.User.Role]++
		}
		if matchesTeam && matchesRole {
			visible = append(visible, item)
		}
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
		name := teamFilterNames[id]
		if name == "" {
			continue
		}
		filters = append(filters, TeamFilter{ID: id, Name: name, Count: count})
	}
	sort.Slice(filters, func(i, j int) bool { return strings.ToLower(filters[i].Name) < strings.ToLower(filters[j].Name) })
	roleFilters := make([]RoleFilter, 0, len(roleCounts))
	for _, role := range domainUser.AllRoles() {
		count, ok := roleCounts[role]
		if !ok {
			continue
		}
		roleFilters = append(roleFilters, RoleFilter{Role: role, DisplayName: domainUser.RoleDisplayName(role), Count: count})
	}

	return &ListResult{
		Items:      visible[offset:end],
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
		Teams:      filters,
		Roles:      roleFilters,
		PageCapabilities: PageCapabilities{
			CanCreateUser:  canCreateUser,
			CanReadDeleted: requesterPermissions.has(domainUser.PermissionUserReadDeleted),
		},
	}, nil
}

func itemHasTeam(item Item, teamID uuid.UUID) bool {
	for _, team := range item.Teams {
		if team.ID == teamID {
			return true
		}
	}
	return false
}

func buildCapabilities(requesterID uuid.UUID, requesterRole domainUser.Role, requesterPermissions permissionSet, target domainUser.User, targetPermissions permissionSet, totalUsers int) Capabilities {
	if requesterID == target.ID {
		return Capabilities{}
	}
	if !canMutateRole(requesterRole, target.Role) {
		return Capabilities{}
	}
	if !permissionSetContainsAll(requesterPermissions, targetPermissions) {
		return Capabilities{}
	}

	canMutateSuperAdmin := !(target.Role == domainUser.RoleSuperAdmin && totalUsers <= 1)
	canRestore := target.IsDeleted() &&
		!target.IsAnonymized() &&
		target.RestoreUntil != nil &&
		time.Now().UTC().Before(*target.RestoreUntil) &&
		requesterPermissions.has(domainUser.PermissionUserDelete) &&
		requesterPermissions.has(domainUser.PermissionUserReadDeleted) &&
		canMutateSuperAdmin
	canUseNormalActions := !target.IsDeleted() && !target.IsAnonymized()
	return Capabilities{
		CanUpdate:     canUseNormalActions && requesterPermissions.has(domainUser.PermissionUserUpdate),
		CanDelete:     canUseNormalActions && requesterPermissions.has(domainUser.PermissionUserDelete) && canMutateSuperAdmin,
		CanDisable:    canUseNormalActions && requesterPermissions.has(domainUser.PermissionUserUpdate) && target.IsActive && canMutateSuperAdmin,
		CanEnable:     canUseNormalActions && requesterPermissions.has(domainUser.PermissionUserUpdate) && !target.IsActive,
		CanRestore:    canRestore,
		CanChangeRole: canUseNormalActions && requesterPermissions.has(domainUser.PermissionUserUpdate),
	}
}

func canMutateRole(requesterRole, targetRole domainUser.Role) bool {
	if requesterRole == domainUser.RoleSuperAdmin {
		return true
	}
	requesterLevel := domainUser.RoleLevel(requesterRole)
	return requesterLevel > 0 && domainUser.RoleLevel(targetRole) < requesterLevel
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
		strings.Contains(strings.ToLower(candidate.EmailValue()), search) ||
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

func (s *Service) loadAllUsers(ctx context.Context, includeDeleted bool) ([]*domainUser.User, error) {
	result, err := s.users.GetPaginatedList(ctx, domain.PaginationParams{Page: 1, Limit: 10000, OrderBy: "last_login_at", Order: "desc", IncludeDeleted: includeDeleted})
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
