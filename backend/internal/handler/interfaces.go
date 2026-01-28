package handler

import (
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainAuth "github.com/besart951/go_infra_link/backend/internal/domain/auth"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
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
	InviteUser(projectID, userID uuid.UUID) error
	CreateControlCabinet(projectID, controlCabinetID uuid.UUID) (*project.ProjectControlCabinet, error)
	UpdateControlCabinet(linkID, projectID, controlCabinetID uuid.UUID) (*project.ProjectControlCabinet, error)
	DeleteControlCabinet(linkID, projectID uuid.UUID) error
	CreateSPSController(projectID, spsControllerID uuid.UUID) (*project.ProjectSPSController, error)
	UpdateSPSController(linkID, projectID, spsControllerID uuid.UUID) (*project.ProjectSPSController, error)
	DeleteSPSController(linkID, projectID uuid.UUID) error
	CreateFieldDevice(projectID, fieldDeviceID uuid.UUID) (*project.ProjectFieldDevice, error)
	UpdateFieldDevice(linkID, projectID, fieldDeviceID uuid.UUID) (*project.ProjectFieldDevice, error)
	DeleteFieldDevice(linkID, projectID uuid.UUID) error
	ListControlCabinets(projectID uuid.UUID, page, limit int) (*domain.PaginatedList[project.ProjectControlCabinet], error)
	ListSPSControllers(projectID uuid.UUID, page, limit int) (*domain.PaginatedList[project.ProjectSPSController], error)
	ListFieldDevices(projectID uuid.UUID, page, limit int) (*domain.PaginatedList[project.ProjectFieldDevice], error)
	ListObjectData(projectID uuid.UUID, page, limit int) (*domain.PaginatedList[domainFacility.ObjectData], error)
	Update(project *project.Project) error
	DeleteByID(id uuid.UUID) error
}

type UserService interface {
	CreateWithPassword(user *user.User, password string) error
	UpdateWithPassword(user *user.User, password *string) error
	GetByID(id uuid.UUID) (*user.User, error)
	List(page, limit int, search, orderBy, order string) (*domain.PaginatedList[user.User], error)
	DeleteByID(id uuid.UUID) error
}

type TeamService interface {
	Create(team *team.Team) error
	GetByID(id uuid.UUID) (*team.Team, error)
	List(page, limit int, search string) (*domain.PaginatedList[team.Team], error)
	Update(team *team.Team) error
	DeleteByID(id uuid.UUID) error

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
