package user

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type Service struct {
	repo domainUser.UserRepository
}

func New(repo domainUser.UserRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(user *domainUser.User) error {
	return s.repo.Create(user)
}

func (s *Service) GetByIds(ids []uuid.UUID) ([]*domainUser.User, error) {
	return s.repo.GetByIds(ids)
}

func (s *Service) GetById(id uuid.UUID) (*domainUser.User, error) {
	users, err := s.repo.GetByIds([]uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, nil
	}
	return users[0], nil
}

func (s *Service) Update(user *domainUser.User) error {
	return s.repo.Update(user)
}

func (s *Service) DeleteByIds(ids []uuid.UUID) error {
	return s.repo.DeleteByIds(ids)
}

func (s *Service) List(page, limit int, search string) (*domain.PaginatedList[domainUser.User], error) {
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}
