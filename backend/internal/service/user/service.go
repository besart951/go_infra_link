package user

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainAuth "github.com/besart951/go_infra_link/backend/internal/domain/auth"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type Service struct {
	repo        domainUser.UserRepository
	passwordSvc domainUser.PasswordHasher
	policy      MutationPolicy
}

type MutationPolicy interface {
	CanDirectCreateUser(ctx context.Context, actorID uuid.UUID, targetRole domainUser.Role) error
	CanUpdateProfile(ctx context.Context, actorID uuid.UUID, target domainUser.User) error
	CanDeleteUser(ctx context.Context, actorID uuid.UUID, target domainUser.User) error
}

// New creates a user service with the given repository and password hasher.
func New(repo domainUser.UserRepository, passwordSvc domainUser.PasswordHasher, mutationPolicy ...MutationPolicy) *Service {
	var policy MutationPolicy
	if len(mutationPolicy) > 0 {
		policy = mutationPolicy[0]
	}
	return &Service{repo: repo, passwordSvc: passwordSvc, policy: policy}
}

func (s *Service) Create(ctx context.Context, user *domainUser.User) error {
	if user.Role == "" {
		user.Role = domainUser.RoleEnterpreneur
	}
	return s.repo.Create(ctx, user)
}

func (s *Service) CreateWithPassword(ctx context.Context, user *domainUser.User, password string) error {
	return s.createWithPassword(ctx, user, password)
}

func (s *Service) CreateWithPasswordForActor(ctx context.Context, actorID uuid.UUID, user *domainUser.User, password string) error {
	if user.Role == "" {
		user.Role = domainUser.RoleEnterpreneur
	}
	if s.policy == nil {
		return domainUser.ErrRoleNotAssignable
	}
	if err := s.policy.CanDirectCreateUser(ctx, actorID, user.Role); err != nil {
		return err
	}
	user.CreatedByID = &actorID
	return s.createWithPassword(ctx, user, password)
}

func (s *Service) createWithPassword(ctx context.Context, user *domainUser.User, password string) error {
	hashedPassword, err := s.passwordSvc.Hash(password)
	if err != nil {
		return domainUser.ErrPasswordHashingFailed
	}

	if user.Role == "" {
		user.Role = domainUser.RoleEnterpreneur
	}

	user.Password = hashedPassword
	return s.repo.Create(ctx, user)
}

func (s *Service) GetByIds(ctx context.Context, ids []uuid.UUID) ([]*domainUser.User, error) {
	return s.repo.GetByIds(ctx, ids)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domainUser.User, error) {
	return domain.GetByID(ctx, s.repo, id)
}

func (s *Service) Update(ctx context.Context, user *domainUser.User) error {
	return s.repo.Update(ctx, user)
}

func (s *Service) UpdateWithPassword(ctx context.Context, user *domainUser.User, password *string) error {
	if password != nil && *password != "" {
		hashedPassword, err := s.passwordSvc.Hash(*password)
		if err != nil {
			return domainUser.ErrPasswordHashingFailed
		}
		user.Password = hashedPassword
	}

	return s.repo.Update(ctx, user)
}

func (s *Service) UpdateProfileForActor(ctx context.Context, actorID uuid.UUID, user *domainUser.User) error {
	if user == nil {
		return domain.ErrInvalidArgument
	}
	if s.policy == nil {
		return domainUser.ErrRoleNotAssignable
	}
	if err := s.policy.CanUpdateProfile(ctx, actorID, *user); err != nil {
		return err
	}
	return s.repo.Update(ctx, user)
}

func (s *Service) UpdatePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) (*domainUser.User, error) {
	if userID == uuid.Nil || currentPassword == "" || newPassword == "" {
		return nil, domain.ErrInvalidArgument
	}
	usr, err := s.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := s.passwordSvc.Compare(usr.Password, currentPassword); err != nil {
		return nil, domainAuth.ErrInvalidCredentials
	}
	hashedPassword, err := s.passwordSvc.Hash(newPassword)
	if err != nil {
		return nil, domainUser.ErrPasswordHashingFailed
	}
	usr.Password = hashedPassword
	if err := s.repo.Update(ctx, usr); err != nil {
		return nil, err
	}
	return usr, nil
}

func (s *Service) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteByIds(ctx, []uuid.UUID{id})
}

func (s *Service) DeleteByIDForActor(ctx context.Context, actorID, userID uuid.UUID) error {
	usr, err := s.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if s.policy == nil {
		return domainUser.ErrRoleNotAssignable
	}
	if err := s.policy.CanDeleteUser(ctx, actorID, *usr); err != nil {
		return err
	}
	return s.DeleteByID(ctx, userID)
}

func (s *Service) List(ctx context.Context, page, limit int, search, orderBy, order string) (*domain.PaginatedList[domainUser.User], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)

	// Default ordering by last_login_at descending
	if orderBy == "" {
		orderBy = "last_login_at"
		order = "desc"
	}

	return s.repo.GetPaginatedList(ctx, domain.PaginationParams{
		Page:    page,
		Limit:   limit,
		Search:  search,
		OrderBy: orderBy,
		Order:   order,
	})
}
