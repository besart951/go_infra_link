package admin

import (
	"context"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type Service struct {
	userRepo userReaderUpdater
	policy   MutationPolicy
}

type userReaderUpdater interface {
	GetByIds(ctx context.Context, ids []uuid.UUID) ([]*domainUser.User, error)
	Update(ctx context.Context, entity *domainUser.User) error
}

type MutationPolicy interface {
	CanDisableUser(ctx context.Context, actorID uuid.UUID, target domainUser.User) error
	CanEnableUser(ctx context.Context, actorID uuid.UUID, target domainUser.User) error
	CanAssignRole(ctx context.Context, actorID uuid.UUID, currentRole, nextRole domainUser.Role) error
}

func New(userRepo userReaderUpdater, policy MutationPolicy) *Service {
	return &Service{userRepo: userRepo, policy: policy}
}

func (s *Service) DisableUser(ctx context.Context, actorID, userID uuid.UUID) error {
	u, err := domain.GetByID(ctx, s.userRepo, userID)
	if err != nil {
		return err
	}
	if s.policy == nil {
		return domainUser.ErrRoleNotAssignable
	}
	if err := s.policy.CanDisableUser(ctx, actorID, *u); err != nil {
		return err
	}
	now := time.Now().UTC()
	u.DisabledAt = &now
	u.IsActive = false
	return s.userRepo.Update(ctx, u)
}

func (s *Service) EnableUser(ctx context.Context, actorID, userID uuid.UUID) error {
	u, err := domain.GetByID(ctx, s.userRepo, userID)
	if err != nil {
		return err
	}
	if s.policy == nil {
		return domainUser.ErrRoleNotAssignable
	}
	if err := s.policy.CanEnableUser(ctx, actorID, *u); err != nil {
		return err
	}
	u.DisabledAt = nil
	u.IsActive = true
	return s.userRepo.Update(ctx, u)
}

func (s *Service) SetUserRole(ctx context.Context, actorID, userID uuid.UUID, role domainUser.Role) error {
	u, err := domain.GetByID(ctx, s.userRepo, userID)
	if err != nil {
		return err
	}
	if s.policy == nil {
		return domainUser.ErrRoleNotAssignable
	}
	if err := s.policy.CanAssignRole(ctx, actorID, u.Role, role); err != nil {
		return err
	}
	u.Role = role
	return s.userRepo.Update(ctx, u)
}
