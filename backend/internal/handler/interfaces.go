package handler

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/project"
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

type AuthService interface {
	Login(email, password string, userAgent, ip *string) (*authsvc.LoginResult, error)
	Refresh(refreshToken string, userAgent, ip *string) (*authsvc.LoginResult, error)
	Logout(refreshToken string) error
}
