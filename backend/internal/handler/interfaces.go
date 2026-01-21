package handler

import (
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainAuth "github.com/besart951/go_infra_link/backend/internal/domain/auth"
	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/besart951/go_infra_link/backend/internal/domain/team"
	"github.com/besart951/go_infra_link/backend/internal/domain/user"
	authsvc "github.com/besart951/go_infra_link/backend/internal/service/auth"
	"github.com/google/uuid"
)

type ProjectService interface {
	Create(project *project.Project) error
	GetByID(id uuid.UUID) (*project.Project, error)
	List(page, limit int, search string) (*domain.PaginatedList[project.Project], error)
	Update(project *project.Project) error
	DeleteByIds(ids []uuid.UUID) error
}

type UserService interface {
	CreateWithPassword(user *user.User, password string) error
	UpdateWithPassword(user *user.User, password *string) error
	GetByID(id uuid.UUID) (*user.User, error)
	List(page, limit int, search string) (*domain.PaginatedList[user.User], error)
	DeleteByIds(ids []uuid.UUID) error
}

type TeamService interface {
	Create(team *team.Team) error
	GetByID(id uuid.UUID) (*team.Team, error)
	List(page, limit int, search string) (*domain.PaginatedList[team.Team], error)
	Update(team *team.Team) error
	DeleteByIds(ids []uuid.UUID) error

	AddMember(teamID, userID uuid.UUID, role team.MemberRole) error
	RemoveMember(teamID, userID uuid.UUID) error
	ListMembers(teamID uuid.UUID, page, limit int) (*domain.PaginatedList[team.TeamMember], error)
}

type AdminService interface {
	DisableUser(userID uuid.UUID) error
	EnableUser(userID uuid.UUID) error
	LockUserUntil(userID uuid.UUID, until time.Time) error
	UnlockUser(userID uuid.UUID) error
	SetUserRole(userID uuid.UUID, role user.Role) error
}

type AuthService interface {
	Login(email, password string, userAgent, ip *string) (*authsvc.LoginResult, error)
	Refresh(refreshToken string, userAgent, ip *string) (*authsvc.LoginResult, error)
	Logout(refreshToken string) error
	CreatePasswordResetToken(adminID, userID uuid.UUID) (string, time.Time, error)
	ConfirmPasswordReset(token, newPassword string) error
	ListLoginAttempts(page, limit int, search string) (*domain.PaginatedList[domainAuth.LoginAttempt], error)
}
