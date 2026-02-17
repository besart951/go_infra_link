package user

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type Service struct {
	repo        domainUser.UserRepository
	passwordSvc domainUser.PasswordHasher
}

// New creates a user service with the given repository and password hasher.
func New(repo domainUser.UserRepository, passwordSvc domainUser.PasswordHasher) *Service {
	return &Service{repo: repo, passwordSvc: passwordSvc}
}

func (s *Service) Create(user *domainUser.User) error {
	if user.Role == "" {
		user.Role = domainUser.RoleEnterpreneur
	}
	return s.repo.Create(user)
}

func (s *Service) CreateWithPassword(user *domainUser.User, password string) error {
	hashedPassword, err := s.passwordSvc.Hash(password)
	if err != nil {
		return domainUser.ErrPasswordHashingFailed
	}

	if user.Role == "" {
		user.Role = domainUser.RoleEnterpreneur
	}

	user.Password = hashedPassword
	return s.repo.Create(user)
}

func (s *Service) GetByIds(ids []uuid.UUID) ([]*domainUser.User, error) {
	return s.repo.GetByIds(ids)
}

func (s *Service) GetByID(id uuid.UUID) (*domainUser.User, error) {
	return domain.GetByID(s.repo, id)
}

func (s *Service) Update(user *domainUser.User) error {
	return s.repo.Update(user)
}

func (s *Service) UpdateWithPassword(user *domainUser.User, password *string) error {
	if password != nil && *password != "" {
		hashedPassword, err := s.passwordSvc.Hash(*password)
		if err != nil {
			return domainUser.ErrPasswordHashingFailed
		}
		user.Password = hashedPassword
	}

	return s.repo.Update(user)
}

func (s *Service) DeleteByID(id uuid.UUID) error {
	return s.repo.DeleteByIds([]uuid.UUID{id})
}

func (s *Service) List(page, limit int, search, orderBy, order string) (*domain.PaginatedList[domainUser.User], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)

	// Default ordering by last_login_at descending
	if orderBy == "" {
		orderBy = "last_login_at"
		order = "desc"
	}

	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:    page,
		Limit:   limit,
		Search:  search,
		OrderBy: orderBy,
		Order:   order,
	})
}
