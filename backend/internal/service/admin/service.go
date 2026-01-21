package admin

import (
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type Service struct {
	userRepo domainUser.UserRepository
}

func New(userRepo domainUser.UserRepository) *Service {
	return &Service{userRepo: userRepo}
}

func (s *Service) DisableUser(userID uuid.UUID) error {
	now := time.Now().UTC()
	users, err := s.userRepo.GetByIds([]uuid.UUID{userID})
	if err != nil {
		return err
	}
	if len(users) == 0 {
		return domain.ErrNotFound
	}
	u := users[0]
	u.DisabledAt = &now
	u.IsActive = false
	return s.userRepo.Update(u)
}

func (s *Service) EnableUser(userID uuid.UUID) error {
	users, err := s.userRepo.GetByIds([]uuid.UUID{userID})
	if err != nil {
		return err
	}
	if len(users) == 0 {
		return domain.ErrNotFound
	}
	u := users[0]
	u.DisabledAt = nil
	u.IsActive = true
	return s.userRepo.Update(u)
}

func (s *Service) LockUserUntil(userID uuid.UUID, until time.Time) error {
	users, err := s.userRepo.GetByIds([]uuid.UUID{userID})
	if err != nil {
		return err
	}
	if len(users) == 0 {
		return domain.ErrNotFound
	}
	u := users[0]
	u.LockedUntil = &until
	return s.userRepo.Update(u)
}

func (s *Service) UnlockUser(userID uuid.UUID) error {
	users, err := s.userRepo.GetByIds([]uuid.UUID{userID})
	if err != nil {
		return err
	}
	if len(users) == 0 {
		return domain.ErrNotFound
	}
	u := users[0]
	u.LockedUntil = nil
	u.FailedLoginAttempts = 0
	return s.userRepo.Update(u)
}

func (s *Service) SetUserRole(userID uuid.UUID, role domainUser.Role) error {
	users, err := s.userRepo.GetByIds([]uuid.UUID{userID})
	if err != nil {
		return err
	}
	if len(users) == 0 {
		return domain.ErrNotFound
	}
	u := users[0]
	u.Role = role
	return s.userRepo.Update(u)
}
