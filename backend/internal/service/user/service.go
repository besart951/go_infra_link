package user

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainAuth "github.com/besart951/go_infra_link/backend/internal/domain/auth"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type Service struct {
	repo            Store
	passwordSvc     domainUser.PasswordHasher
	policy          MutationPolicy
	deletionService *DeletionService
}

const deleteRestoreRetention = 30 * 24 * time.Hour

type Store interface {
	domain.Repository[domainUser.User]
	domainUser.UserEmailRepository
	ListDueForAnonymization(ctx context.Context, now time.Time, limit int) ([]*domainUser.User, error)
}

type MutationPolicy interface {
	CanDirectCreateUser(ctx context.Context, actorID uuid.UUID, targetRole domainUser.Role) error
	CanUpdateProfile(ctx context.Context, actorID uuid.UUID, target domainUser.User) error
	CanDeleteUser(ctx context.Context, actorID uuid.UUID, target domainUser.User) error
	CanRestoreUser(ctx context.Context, actorID uuid.UUID, target domainUser.User) error
}

// New creates a user service with the given repository and password hasher.
func New(repo Store, passwordSvc domainUser.PasswordHasher, mutationPolicy ...MutationPolicy) *Service {
	var policy MutationPolicy
	if len(mutationPolicy) > 0 {
		policy = mutationPolicy[0]
	}
	svc := &Service{repo: repo, passwordSvc: passwordSvc, policy: policy}
	// Initialize DeletionService to own soft-delete lifecycle
	svc.deletionService = NewDeletionService(repo, policy)
	return svc
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
	email := domainUser.NormalizeEmail(user.EmailValue())
	if email == "" {
		return domain.ErrInvalidArgument
	}
	user.SetEmail(email)
	if existing, err := s.repo.GetByEmail(ctx, email); err == nil {
		if existing.IsDeleted() && !existing.IsAnonymized() {
			if existing.RestoreUntil != nil && existing.RestoreUntil.After(time.Now().UTC()) {
				return domainUser.ErrDeletedUserRestorable
			}
		}
		return domain.ErrConflict
	} else if !errors.Is(err, domain.ErrNotFound) {
		return err
	}

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
	return s.deletionService.DeleteByID(ctx, id)
}

func (s *Service) DeleteByIDForActor(ctx context.Context, actorID, userID uuid.UUID) error {
	return s.deletionService.DeleteByIDForActor(ctx, actorID, userID)
}

func (s *Service) List(ctx context.Context, page, limit int, search, orderBy, order string, includeDeleted bool) (*domain.PaginatedList[domainUser.User], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)

	// Default ordering by last_login_at descending
	if orderBy == "" {
		orderBy = "last_login_at"
		order = "desc"
	}

	result, err := s.repo.GetPaginatedList(ctx, domain.PaginationParams{
		Page:           page,
		Limit:          limit,
		Search:         search,
		OrderBy:        orderBy,
		Order:          order,
		IncludeDeleted: includeDeleted,
	})
	if err != nil {
		return nil, err
	}
	if includeDeleted {
		return result, nil
	}

	return result, nil
}

func (s *Service) RestoreByIDForActor(ctx context.Context, actorID, userID uuid.UUID) error {
	return s.deletionService.RestoreByIDForActor(ctx, actorID, userID)
}

func (s *Service) PurgeDeletedUsers(ctx context.Context, limit int) error {
	if limit <= 0 {
		limit = 100
	}
	now := time.Now().UTC()
	candidates, err := s.repo.ListDueForAnonymization(ctx, now, limit)
	if err != nil {
		return err
	}
	for _, candidate := range candidates {
		if candidate == nil || candidate.IsAnonymized() {
			continue
		}
		// Delegate anonymization to DeletionService
		if err := s.deletionService.Anonymize(ctx, candidate.ID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) StartDeletedUserPurgeWorker(interval time.Duration, batchSize int) func() {
	if interval <= 0 {
		interval = time.Hour
	}
	if batchSize <= 0 {
		batchSize = 100
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		s.purgeDeletedUsersWithLog(ctx, batchSize)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.purgeDeletedUsersWithLog(ctx, batchSize)
			}
		}
	}()
	return cancel
}

func (s *Service) purgeDeletedUsersWithLog(ctx context.Context, batchSize int) {
	if err := s.PurgeDeletedUsers(ctx, batchSize); err != nil {
		slog.Warn("deleted user purge failed", "err", err)
	}
}

func markUserDeleted(usr *domainUser.User, actorID uuid.UUID, now time.Time) {
	restoreUntil := now.Add(deleteRestoreRetention)
	usr.DeletedAt = &now
	if actorID != uuid.Nil {
		usr.DeletedByID = &actorID
	}
	usr.RestoreUntil = &restoreUntil
	usr.ScheduledPurgeAt = &restoreUntil
	usr.DisabledAt = &now
	usr.IsActive = false
	usr.LockedUntil = nil
	usr.FailedLoginAttempts = 0
	usr.LastLoginAt = nil
	if email := domainUser.NormalizeEmail(usr.EmailValue()); email != "" {
		hash := sha256.Sum256([]byte(email))
		hashText := hex.EncodeToString(hash[:])
		usr.DeletedEmailHash = &hashText
	}
}

func clearDeleteMarkers(usr *domainUser.User) {
	usr.DeletedAt = nil
	usr.DeletedByID = nil
	usr.RestoreUntil = nil
	usr.ScheduledPurgeAt = nil
	usr.AnonymizedAt = nil
	usr.DeletedEmailHash = nil
}

func anonymizeUser(usr *domainUser.User, now time.Time) {
	if usr.DeletedEmailHash == nil {
		if email := domainUser.NormalizeEmail(usr.EmailValue()); email != "" {
			hash := sha256.Sum256([]byte(email))
			hashText := hex.EncodeToString(hash[:])
			usr.DeletedEmailHash = &hashText
		}
	}
	usr.FirstName = "Deleted"
	usr.LastName = "User"
	usr.Email = nil
	usr.Password = ""
	usr.IsActive = false
	usr.DisabledAt = &now
	usr.RestoreUntil = nil
	usr.ScheduledPurgeAt = nil
	usr.AnonymizedAt = &now
	usr.LockedUntil = nil
	usr.LastLoginAt = nil
	usr.FailedLoginAttempts = 0
}
